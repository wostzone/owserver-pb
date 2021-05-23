package internal

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi-go/pkg/td"
)

// PollValues polls the OWServer hub for Thing property values
// Returns a map of thingID's containing a map of property name:value pairs
func (pb *OWServerPB) PollValues() (map[string](map[string]interface{}), error) {
	// tds is a map of ThingID:{attr:value,...}
	thingValues := make(map[string](map[string]interface{}), 0)
	if pb.edsAPI == nil {
		err := fmt.Errorf("EDS API not initialized")
		logrus.Error(err)
		return nil, err
	}
	// Read the values from the EDS gateway
	startTime := time.Now()
	rootNode, err := pb.edsAPI.ReadEds()
	endTime := time.Now()
	latency := endTime.Sub(startTime)
	if err != nil {
		return nil, err
	}
	nodeList := pb.edsAPI.ParseOneWireNodes(rootNode, latency, true)
	for _, node := range nodeList {
		thingID := td.CreatePublisherThingID(pb.hubConfig.Zone, PluginID, node.NodeID, node.DeviceType)

		propValues := make(map[string]interface{})
		for name, attr := range node.Attr {
			propValues[name] = attr.Value
		}

		thingValues[thingID] = propValues
	}
	return thingValues, nil
}
