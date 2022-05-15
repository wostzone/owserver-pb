package internal

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

// PollValues obtains thing property values of each Thing and converts the EDS property
// names to vocabulary names.
// This returns a map of device/node IDs containing a maps of property name-value pairs
func (pb *OWServerPB) PollValues() (nodeValues map[string](map[string]interface{}), err error) {

	if pb.edsAPI == nil || !pb.running {
		err = fmt.Errorf("EDS API not initialized")
		logrus.Error(err)
		return
	}
	nodeValues, err = pb.edsAPI.PollValues()
	// update service properties if enabled
	if pb.Config.PublishTD {
		serviceProps := make(map[string]interface{})
		serviceProps[vocab.PropNameGatewayAddress] = pb.edsAPI.GetLastAddress()
		nodeValues[pb.Config.ClientID] = serviceProps
	}
	return nodeValues, err
}

// PublishValues publishes updated thing property values of each TD
// This takes a map of device IDs and properties [device IDs] (property map)
//  and emits the properties as an update event.
func (pb *OWServerPB) PublishValues(thingValues map[string](map[string]interface{})) error {
	if thingValues == nil {
		err := errors.New("missing values")
		logrus.Errorf("OWServerPB.PublishValues. thingValues is nil")
		return err
	}
	logrus.Infof("OWServerPB.PublishValues for %d things", len(thingValues))
	for deviceID, propValues := range thingValues {
		eThing, found := pb.eThings[deviceID]
		if found {
			// Publish property values that have changed
			err := eThing.EmitPropertyChanges(propValues, true)
			if err != nil {
				return err
			}
		} else {
			logrus.Errorf("PublishValues. Device with ID %s has no Exposed Thing", deviceID)
		}
	}
	return nil
}

// UpdatePropertyValues polls the OWServer hub for Thing property values and pass updates
// to the Exposed Thing.
func (pb *OWServerPB) UpdatePropertyValues() error {
	nodeValueMap, err := pb.PollValues()
	if err == nil {
		err = pb.PublishValues(nodeValueMap)
	}
	return err
}
