package internal

import (
	"github.com/wostzone/owserver/internal/eds"
	"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/vocab"
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
				// what input type is expected?
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
