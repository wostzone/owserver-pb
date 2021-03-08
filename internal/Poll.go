// Package internal queries the EDS for device, node and parameter information
package internal

const EdsThingID = "eds"

// UpdateEdsError updates the TD thing error status
//  err with error to set, nil to clear
func UpdateEdsError(err error) {
	thing := td.GetThing(EdsThingID)
	if thing != nil {
	thing.SetErrorStatus(err)

}

// UpdateEdsGateway updates the TD of the EDS gateway
func UpdateEdsGateway(gwParam map[string]string) {
	thing := td.GetThing(EdsThingID)
	if thing == nil {
		thing = td.NewThing(EdsThingID, "EDS OWServer 1-wire Gateway")
	}
	for prop, val := range gwParam {
		thing.SetProperty(prop, val)
	}
	// td := NewTD(edsID, "EDS OWServer 1-wire Gateway")
	// td.AddProperty(NewStringProperty("address", eds.address, "EDS Gateway IP address"))

}

// UpdateEdsNode updates the TD of a 1-wire node
func UpdateEdsNode(nodeParam XMLNode) {
}

// Poll the EDS gateway for updates to nodes and sensors
func Poll(eds *EdsAPI) {
	rootNode, err := eds.ReadEds()
	if err != nil {
		// The EDS cannot be reached. Set its error status
		UpdateEdsError(err)
		return
	}

	gwParams, deviceNodes := eds.ParseNodeParams(rootNode)

	// clear any errors
	UpdateEdsError(err)
	UpdateEdsGateway(gwParams)

	// (re)discover the nodes on the gateway
	// td.UpdateGateway(gwParams)
	// 	app.updateGateway(gwParams)
	// 	pub.UpdateNodeStatus(gwID, map[types.NodeStatus]string{
	// 		types.NodeStatusRunState:    string(types.NodeRunStateReady),
	// 		types.NodeStatusLastError:   "",
	// 		types.NodeStatusLatencyMSec: fmt.Sprintf("%d", latency.Milliseconds()),
	// 	})

	// (re)discover any new sensor nodes and publish when changed
	for _, node := range deviceNodes {
		UpdateEdsNode(&node)
	}
}
