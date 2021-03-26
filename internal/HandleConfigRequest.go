// Package internal handles node configuration commands
package internal

import (
	"github.com/sirupsen/logrus"
)

// HandleConfigRequest handles requests to update a Thing's configuration
// There are currently no node configurations to update to onewire
func (pb *OWServerPB) HandleConfigRequest(thingID string, config map[string]interface{}) {
	logrus.Infof("HandleConfigRequest for Thing %s.", thingID)
	// for now accept all configuration
	// pb.UpdateThingConfigValues(nodeHWID, config)
}
