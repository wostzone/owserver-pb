// Package internal handles node configuration commands
package internal

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"time"
)

// HandleConfigRequest handles requests to update a Thing's configuration
// There are currently no node configurations to update to onewire
func (pb *OWServerPB) HandleConfigRequest(
	eThing *mqttbinding.MqttExposedThing, propName string, io mqttbinding.InteractionOutput) error {
	logrus.Infof("HandleConfigRequest for Thing %s. propName=%s", eThing.GetThingDescription().GetID(), propName)
	// for now accept all configuration

	// If the property name is converted to a standardized vocabulary then convert the name
	// to the EDS writable property name.

	err := pb.edsAPI.WriteData(eThing.DeviceID, propName, io.ValueAsString())
	if err == nil {
		time.Sleep(time.Second)
		_ = pb.UpdatePropertyValues()
		// The EDS is slow, retry in case it was missed
		time.Sleep(time.Second * 2)
		err = pb.UpdatePropertyValues()
	} else {
		logrus.Errorf("HandleConfigRequest results in error: %s", err)
	}
	return err
}
