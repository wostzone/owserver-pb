package internal

import (
	"fmt"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"github.com/wostzone/hub/lib/client/pkg/vocab"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/owserver-pb/internal/eds"
)

// CreateTDFromNode converts the node into a TD that describes the node.
// - All attributes will be added as node properties
// - Writable non-sensors attributes are marked as writable configuration
// - Sensors are also added as events.
// - Writable sensors are also added as actions.
// This is only used when a new Exposed Thing is created
func (pb *OWServerPB) CreateTDFromNode(node *eds.OneWireNode) (tdoc *thing.ThingTD) {
	thingID := thing.CreatePublisherID(pb.zone, PluginID, node.NodeID, node.DeviceType)
	tdoc = thing.CreateTD(thingID, node.Name, node.DeviceType)
	tdoc.UpdateTitleDescription(node.Name, node.Description)

	// Map node attribute to Thing properties
	for attrName, attr := range node.Attr {
		prop := tdoc.AddProperty(attrName, attr.Name, vocab.WoTDataTypeString)
		prop.Unit = attr.Unit

		if attr.IsSensor {
			prop.AtType = string(vocab.PropertyTypeOutput)
			// sensors emit events
			evAff := tdoc.AddEvent(attrName, attrName, vocab.WoTDataTypeString)
			evAff.Data.Type = vocab.WoTDataTypeString
			evAff.Data.Unit = prop.Unit

			// writable sensors are actuators and can be triggered with actions
			if attr.Writable {
				actionAff := tdoc.AddAction(attrName, attrName, vocab.WoTDataTypeString)
				actionAff.Input.Type = vocab.WoTDataTypeString
				actionAff.Input.Unit = prop.Unit
			}
		} else {
			// non-sensors are attributes. Writable attributes are configuration.
			if attr.Writable {
				prop.AtType = string(vocab.PropertyTypeConfig)
				prop.ReadOnly = false
			} else {
				prop.AtType = string(vocab.PropertyTypeAttr)
				prop.ReadOnly = true
			}
		}
	}

	return
}

// PollTDs polls the OWServer hub and updates the ExposedThing if needed
func (pb *OWServerPB) PollTDs() (err error) {

	if pb.edsAPI == nil || !pb.running {
		err := fmt.Errorf("EDS API not initialized")
		logrus.Error(err)
		return err
	}

	rootNode, err := pb.edsAPI.ReadEds()
	if err != nil {
		// if pb.gatewayInfo.thingTD != nil {
		// 	// The EDS cannot be reached. Set its error status
		// 	td.SetThingErrorStatus(pb.gatewayInfo.thingTD, err.Error())
		// }
		return err
	}
	nodeList := pb.edsAPI.ParseOneWireNodes(rootNode, 0, true)

	for _, node := range nodeList {
		pb.UpdateExposedThingFromNode(node)
	}
	return nil
}

// UpdateExposedThingFromNode ensures that an exposed thing exists for the onewire node
// This updates the schema.
func (pb *OWServerPB) UpdateExposedThingFromNode(node *eds.OneWireNode) {
	//eThing, found := pb.eThings[node.NodeID]
	//if !found {
	tdoc := pb.CreateTDFromNode(node)
	eThing := mqttbinding.CreateExposedThing(tdoc, pb.hubClient)
	pb.eThings[node.NodeID] = eThing
	_ = eThing.Expose()
	//} else {
	//	// Node metadata doesn't change
	//	_ = eThing.Expose()
	//}
}
