// Package internal handles input set command
package internal

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"time"
)

// HandleActionRequest handles requests to activate inputs
// Not supported
func (pb *OWServerPB) HandleActionRequest(
	eThing *mqttbinding.MqttExposedThing,
	actionName string,
	io mqttbinding.InteractionOutput) error {

	logrus.Infof("HandleActionRequest for Thing %s. Action=%s Value=%s",
		eThing.GetThingDescription().GetID(), actionName, io.ValueAsString())

	// If the action name is converted to a standardized vocabulary then convert the name
	// to the EDS writable property name.
	// FIXME lookup of the action affordance should be in the ExposedThing
	actionAffordance := eThing.GetThingDescription().GetAction(actionName)
	if actionAffordance == nil {
		return errors.New("Unknown action " + actionName)
	}

	err := pb.edsAPI.WriteData(eThing.DeviceID, actionName, io.ValueAsString())
	if err == nil {
		time.Sleep(time.Second)
		_ = pb.UpdatePropertyValues()
		// The EDS is slow, retry in case it was missed
		time.Sleep(time.Second * 2)
		err = pb.UpdatePropertyValues()
	}
	return err
}
