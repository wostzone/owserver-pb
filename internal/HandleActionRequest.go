// Package internal handles input set command
package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/owserver/internal/eds"
	"github.com/wostzone/wost-go/pkg/exposedthing"
	"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/vocab"
)

// HandleActionRequest handles requests to activate inputs
// Not supported
func (pb *OWServerPB) HandleActionRequest(
	eThing *exposedthing.ExposedThing,
	actionName string,
	io *thing.InteractionOutput) error {

	logrus.Infof("Thing %s. Action=%s Value=%s",
		eThing.GetThingDescription().GetID(), actionName, io.ValueAsString())

	// If the action name is converted to a standardized vocabulary then convert the name
	// to the EDS writable property name.
	// FIXME lookup of the action affordance should be in the ExposedThing
	actionAffordance := eThing.GetThingDescription().GetAction(actionName)
	if actionAffordance == nil {
		return errors.New("Unknown action " + actionName)
	}

	// lookup the action name used by the EDS
	edsName := eds.LookupEdsName(actionName)

	// determine the value. Booleans are submitted as integers
	actionValue := io.ValueAsString()
	if io.Schema.Type == vocab.WoTDataTypeBool {
		actionValue = fmt.Sprint(io.ValueAsInt())
	}

	err := pb.edsAPI.WriteData(eThing.DeviceID, edsName, actionValue)
	if err == nil {
		time.Sleep(time.Second)
		err = pb.UpdatePropertyValues(true)

		// The EDS is slow, retry in case it was missed
		time.Sleep(time.Second * 4)
		err = pb.UpdatePropertyValues(true)
	}
	return err
}
