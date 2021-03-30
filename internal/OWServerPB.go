package internal

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/hubapi/api"
	"github.com/wostzone/hubapi/pkg/hubclient"
	"github.com/wostzone/hubapi/pkg/hubconfig"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "owserver"

// PluginConfig with owserver plugin configuration
type PluginConfig struct {
	EdsAddress       string `yaml:"owserverAddress"`
	LoginName        string `yaml:"loginName"`
	Password         string `yaml:"password"`
	TDIntervalSec    int    `yaml:"tdInterval"`    // interval of republishing the full TD, default is 1 hours
	ValueIntervalSec int    `yaml:"valueInterval"` // interval of republishing the Thing property values, default is 60 seconds
}

// ThingInfo contains the last published property values
// type ThingInfo struct {
// 	nodeID         string            // node to thing ID mapping
// 	thingID        string            // provided by pollTD
// 	propertyValues map[string]string // provided by pollValues
// }

// OWServerPB is a  hub protocol binding plugin for capturing 1-wire OWServer V2 Data
type OWServerPB struct {
	config    PluginConfig         // options for accessing EDS OWServer
	edsAPI    *EdsAPI              // EDS device access
	hubConfig *hubconfig.HubConfig // hub based configuration
	hubClient api.IHubClient
	nodeInfo  map[string]*OneWireNode // map of node ID to node info and thingID
	running   bool
}

// PublishThingsTD publishes the TD of Things
func (pb *OWServerPB) PublishTDs(tds map[string]api.ThingTD) error {
	var err error
	for thingID, td := range tds {
		err = pb.hubClient.PublishTD(thingID, td)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishThingsTD publishes the TD of Things
//
func (pb *OWServerPB) PublishValues(thingValues map[string](map[string]interface{})) error {
	var err error
	for thingID, propValues := range thingValues {
		err = pb.hubClient.PublishPropertyValues(thingID, propValues)
		if err != nil {
			return err
		}
	}
	return nil
}

// heartbeat polls the EDS server every X seconds
func (pb *OWServerPB) heartbeat() {
	var tdCountDown = 0
	var valueCountDown = 0
	for pb.running {
		tdCountDown--
		if tdCountDown <= 0 {
			tds, err := pb.PollTDs()
			if err == nil {
				pb.PublishTDs(tds)
			}
			tdCountDown = pb.config.TDIntervalSec
		}
		valueCountDown--
		if valueCountDown <= 0 {
			values, err := pb.PollValues()
			if err == nil {
				pb.PublishValues(values)
			}
			valueCountDown = pb.config.ValueIntervalSec
		}
		time.Sleep(time.Second)
	}
}

// Start connects to the hub internal message bus and starts polling
// the owserver.
func (pb *OWServerPB) Start(hubConfig *hubconfig.HubConfig, pluginConfig *PluginConfig) error {
	var err error
	pb.config = *pluginConfig
	pb.hubConfig = hubConfig
	pb.nodeInfo = make(map[string]*OneWireNode, 0) // map of node thing info objects by thing ID
	pb.edsAPI = NewEdsAPI(pluginConfig.EdsAddress, pluginConfig.LoginName, pluginConfig.Password)
	pb.hubClient = hubclient.NewPluginClient(PluginID, hubConfig)
	err = pb.hubClient.Start(false)
	if err != nil {
		logrus.Errorf("Protocol Binding for OWServer startup failed")
		return err
	}
	pb.running = true
	go pb.heartbeat()
	logrus.Infof("Service OWServer startup completed")
	return nil
}

// Stop the service
func (pb *OWServerPB) Stop() {
	pb.running = false
	logrus.Info("Stopping service OWServer")
}
