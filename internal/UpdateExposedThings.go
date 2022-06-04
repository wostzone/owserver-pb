package internal

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/owserver/internal/eds"
	"github.com/wostzone/wost-go/pkg/exposedthing"
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
		prop := tdoc.AddProperty(attrName, attr.Name, attr.DataType)
		prop.Unit = attr.Unit

		// sensors are added as both properties and events
		if attr.IsSensor {
			// sensors emit events
			evAff := tdoc.AddEvent(attrName, attrName, attr.DataType)
			evAff.Data.Unit = prop.Unit

			// writable sensors are actuators and can be triggered with actions
			if attr.Writable {
				actionAff := tdoc.AddAction(attrName, attrName, attr.DataType)
				actionAff.Input.Unit = prop.Unit
			}
		} else {
			// non-sensors are attributes. Writable attributes are configuration.
			if attr.Writable {
				prop.ReadOnly = false
			} else {
				prop.ReadOnly = true
			}
		}
	}
	return
}

// CreateExposedThingFromNode ensures that an exposed thing exists for the onewire node
// This updates the schema.
func (pb *OWServerPB) CreateExposedThingFromNode(node *eds.OneWireNode) {
	//eThing, found := pb.eThings[node.NodeID]
	//if !found {
	tdoc := pb.CreateTDFromNode(node)
	eThing, found := pb.eFactory.Expose(node.NodeID, tdoc)
	if !found {
		eThing.SetPropertyWriteHandler("", pb.HandleConfigRequest)
		eThing.SetActionHandler("", pb.HandleActionRequest)
		pb.mu.Lock()
		pb.eThings[node.NodeID] = eThing
		pb.mu.Unlock()
	}
	//} else {
	//	// Node metadata doesn't change
	//	_ = eThing.Expose()
	//}
}

// CreateExposedThingForService creates the Thing Description document of the service itself
// and exposes it.
//
// TD attributes of this service includes are:
//    'address' - gateway address
func (pb *OWServerPB) CreateExposedThingForService() *exposedthing.ExposedThing {
	deviceType := vocab.DeviceTypeService
	thingID := thing.CreatePublisherID(pb.zone, pb.Config.ClientID, pb.Config.ClientID, deviceType)
	logrus.Infof("Publishing this service TD %s", thingID)

	// Create the TD document for this protocol binding
	tdoc := thing.CreateTD(thingID, "OWServer Service", deviceType)
	tdoc.UpdateTitleDescription("EDS OWServer-V2 Protocol binding",
		"This service publishes information on The EDS OWServer 1-wire gateway and its connected sensors")

	// Include the service properties (attributes and configuration)
	tdoc.AddProperty(vocab.PropNameGatewayAddress, "Gateway Address", vocab.WoTDataTypeString)

	eThing, found := pb.eFactory.Expose(pb.Config.ClientID, tdoc)
	if !found {
		pb.mu.Lock()
		pb.eThings[pb.Config.ClientID] = eThing
		pb.mu.Unlock()

		eThing.SetPropertyWriteHandler("",
			func(eThing *exposedthing.ExposedThing, propName string, value *thing.InteractionOutput) error {
				// TODO: add handle configuration changes (once there are any)
				return nil
			})
	}
	return eThing
}

// UpdateExposedThings polls the OWServer hub and makes sure an ExposedThing exist for each node
func (pb *OWServerPB) UpdateExposedThings() (err error) {

	pb.mu.Lock()
	isRunning := pb.running
	pb.mu.Unlock()

	if pb.edsAPI == nil || !isRunning {
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
		pb.CreateExposedThingFromNode(node)
	}
	return nil
}
