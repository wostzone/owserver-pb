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

// UpdateEdsHub updates the TD of the EDS hub
func UpdateEdsHub(gwParam map[string]string) {
	thing := td.GetThing(EdsThingID)
	if thing == nil {
		thing = td.NewThing(EdsThingID, "EDS OWServer 1-wire Hub")
	}
	for prop, val := range gwParam {
		thing.SetProperty(prop, val)
	}
	// td := NewTD(edsID, "EDS OWServer 1-wire Hub")
	// td.AddProperty(NewStringProperty("address", eds.address, "EDS Hub IP address"))

}

// UpdateEdsNode updates the TD of a 1-wire node
func UpdateEdsNode(nodeParam XMLNode) {
}

// Poll the EDS hub for updates to nodes and sensors
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
	UpdateEdsHub(gwParams)

	// (re)discover the nodes on the hub
	// td.UpdateHub(gwParams)
	// 	app.updateHub(gwParams)
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
