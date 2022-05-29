// Package internal handles node configuration commands
package internal

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/owserver-pb/internal/eds"
	"github.com/wostzone/wost-go/pkg/exposedthing"
	"github.com/wostzone/wost-go/pkg/thing"
	"time"
)

// HandleConfigRequest handles requests to update a Thing's configuration
// There are currently no node configurations to update to onewire
func (pb *OWServerPB) HandleConfigRequest(
	eThing *exposedthing.ExposedThing, propName string, io *thing.InteractionOutput) error {
	logrus.Infof("HandleConfigRequest for Thing %s. propName=%s", eThing.GetThingDescription().GetID(), propName)
	// for now accept all configuration

	// If the property name is converted to a standardized vocabulary then convert the name
	// to the EDS writable property name.
	edsName := eds.LookupEdsName(propName)

	err := pb.edsAPI.WriteData(eThing.DeviceID, edsName, io.ValueAsString())
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
