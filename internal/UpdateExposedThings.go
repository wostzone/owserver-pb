package internal

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/owserver-pb/internal/eds"
)

// PollProperties polls the OWServer hub and updates ExposedThing properties if needed
func (pb *OWServerPB) PollProperties() (err error) {

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
	eThing := pb.eFactory.Expose(node.NodeID, tdoc)
	eThing.SetPropertyWriteHandler("", pb.HandleConfigRequest)
	eThing.SetActionHandler("", pb.HandleActionRequest)

	pb.eThings[node.NodeID] = eThing
	//} else {
	//	// Node metadata doesn't change
	//	_ = eThing.Expose()
	//}
}
