package internal

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/api"
	"github.com/wostzone/hubapi/pkg/td"
)

// family to device type. See also: http://owfs.sourceforge.net/simple_family.html
// Todo: get from config file so it is easy to update
var deviceTypeMap = map[string]api.DeviceType{
	"10": api.DeviceTypeThermometer,
	"28": api.DeviceTypeThermometer,
	"7E": api.DeviceTypeMultisensor,
}

// CreateTDFromNode converts the node into a TD
func (pb *OWServerPB) CreateTDFromNode(node *OneWireNode) (thingID string, thingTD api.ThingTD) {
	thingID = td.CreatePublisherThingID(pb.hubConfig.Zone, PluginID, node.NodeID, node.DeviceType)
	thingTD = td.CreateTD(thingID, api.DeviceTypeGateway)
	td.SetThingDescription(thingTD, node.Name, node.Description)

	// Map node attribute to Thing properties
	for attrName, attr := range node.Attr {
		prop := td.CreateTDProperty(attrName, "", attr.PropertyType)
		td.SetTDPropertyDataTypeString(prop, 0, 0)
		if attr.Unit != "" {
			td.SetTDPropertyUnit(prop, attr.Unit)
		}
		td.AddTDProperty(thingTD, attrName, prop)
	}

	return
}

// PollTDs polls the OWServer hub and converts the result to Thing Definitions
func (pb *OWServerPB) PollTDs() (map[string]api.ThingTD, error) {
	// tds is a map of ThingID:ThingTD
	tds := make(map[string]api.ThingTD, 0)

	if pb.edsAPI == nil {
		err := fmt.Errorf("EDS API not initialized")
		logrus.Error(err)
		return nil, err
	}

	rootNode, err := pb.edsAPI.ReadEds()
	if err != nil {
		// if pb.gatewayInfo.thingTD != nil {
		// 	// The EDS cannot be reached. Set its error status
		// 	td.SetThingErrorStatus(pb.gatewayInfo.thingTD, err.Error())
		// }
		return nil, err
	}
	nodeList := pb.edsAPI.ParseOneWireNodes(rootNode, 0, true)

	for _, node := range nodeList {
		thingID, td := pb.CreateTDFromNode(node)
		tds[thingID] = td
	}
	// // td.SetThingErrorStatus(pb.gatewayTD, "")
	// gwNodeID, gwThingID, gwTD := pb.CreateGatewayTD(nodeList[0])
	// pb.setThingInfo(gwNodeID, gwThingID)

	// // (re)discover any new sensor nodes and publish when changed
	// for _, node := range deviceNodes {
	// 	nodeID, thingID, nodeTD := pb.CreateNodeTD(&node)
	// 	tds[nodeID] = nodeTD
	// 	pb.setThingInfo(nodeID, thingID)
	// }
	return tds, nil
}
