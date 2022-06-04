package internal

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/wost-go/pkg/vocab"
)

// PollNodeValues obtains thing property values of each Thing and converts the EDS property
// names to vocabulary names.
// This returns a map of device/node IDs containing a maps of property name-value pairs
func (pb *OWServerPB) PollNodeValues() (nodeValues map[string](map[string]interface{}), err error) {

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
func (pb *OWServerPB) PublishValues(thingValues map[string](map[string]interface{}), onlyChanges bool) error {
	if thingValues == nil {
		err := errors.New("missing values")
		logrus.Errorf("thingValues is nil")
		return err
	}
	logrus.Infof("%d things", len(thingValues))
	for deviceID, propValues := range thingValues {
		pb.mu.Lock()
		eThing, found := pb.eThings[deviceID]
		pb.mu.Unlock()
		if found {
			// submit each property in turn
			for propName, newVal := range propValues {
				err := eThing.EmitPropertyChange(propName, newVal, onlyChanges)
				if err != nil {
					return err
				}
			}

			// Publish property values that have changed
			//err := eThing.EmitPropertyChange(propValues, true)
			//if err != nil {
			//	return err
			//}
		} else {
			logrus.Errorf("Device with ID %s has no Exposed Thing", deviceID)
		}
	}
	return nil
}

// UpdatePropertyValues polls the OWServer hub for Thing property values and pass updates
// to the Exposed Thing.
//  onlyChanges only submit changed values
func (pb *OWServerPB) UpdatePropertyValues(onlyChanges bool) error {
	nodeValueMap, err := pb.PollNodeValues()
	if err == nil {
		err = pb.PublishValues(nodeValueMap, onlyChanges)
	}
	return err
}
