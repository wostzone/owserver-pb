// Package internal handles input set command
package internal

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
)

// HandleActionRequest handles requests to activate inputs
// TODO: support for controlling onewire inputs
func (pb *OWServerPB) HandleActionRequest(
	eThing *mqttbinding.MqttExposedThing,
	actionName string,
	io mqttbinding.InteractionOutput) error {

	logrus.Infof("HandleActionRequest for Thing %s. Action=%s", eThing.GetThingDescription().GetID(), actionName)
	// trigger an update of property values to get a quick response
	return nil
}
