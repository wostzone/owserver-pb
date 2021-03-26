// Package internal queries the EDS for device, node and parameter information
package internal

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/pkg/td"
)

// Poll the OWServer hub for device updates
func (pb *OWServerPB) Poll() error {
	if pb.edsAPI == nil {
		err := fmt.Errorf("EDS API not initialized")
		logrus.Error(err)
		return err
	}

	startTime := time.Now()
	rootNode, err := pb.edsAPI.ReadEds()
	endTime := time.Now()
	latency := endTime.Sub(startTime)

	if err != nil {
		if pb.gatewayInfo.thingTD != nil {
			// The EDS cannot be reached. Set its error status
			td.SetThingErrorStatus(pb.gatewayInfo.thingTD, err.Error())
		}
		return err
	}
	// td.SetThingErrorStatus(pb.gatewayTD, "")

	gwParams, deviceNodes := pb.edsAPI.ParseNodeParams(rootNode)
	// Update the EDS Gateway Thing
	gwID, gwTD, newValues := pb.CreateGatewayTD(gwParams, latency)

	pb.gatewayInfo.thingTD = gwTD
	// TODO: Only republish TD at a given interval
	pb.hubClient.PublishTD(gwID, gwTD)
	pb.hubClient.PublishPropertyValues(gwID, newValues)

	// (re)discover any new sensor nodes and publish when changed
	for _, node := range deviceNodes {
		nodeID, nodeTD := pb.CreateNodeTD(&node)
		pb.nodeInfo[nodeID] = ThingInfo{
			thingTD:        nodeTD,
			propertyValues: make(map[string]string),
		}
		// TODO, only republish at a given interval
		pb.hubClient.PublishTD(nodeID, nodeTD)
	}
	return nil
}
