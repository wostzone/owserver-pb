package internal

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/owserver-pb/internal/eds"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/vocab"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "owserver-pb"

// PluginConfig with owserver plugin configuration
type PluginConfig struct {
	ClientID      string `yaml:"clientID"` // custom unique client ID, default is the pluginID
	EdsAddress    string `yaml:"owserverAddress"`
	LoginName     string `yaml:"loginName"`
	Password      string `yaml:"password"`
	PublishTD     bool   `yaml:"publishTD"`     // publish the TD of this service
	TDInterval    int    `yaml:"tdInterval"`    // interval of republishing the full TD, default is 1 hours
	ValueInterval int    `yaml:"valueInterval"` // interval of republishing the Thing property values, default is 60 seconds
}

// OWServerPB is a  hub protocol binding plugin for capturing 1-wire OWServer V2 Data
type OWServerPB struct {
	Config    PluginConfig         // options for accessing EDS OWServer
	edsAPI    *eds.EdsAPI          // EDS device access
	hubConfig *hubconfig.HubConfig // hub based configuration
	hubClient *hubclient.MqttHubClient
	nodeInfo  map[string]*eds.OneWireNode // map of node ID to node info and thingID
	running   bool
}

// PublishServiceTD publishes the Thing Description of the service itself
func (pb *OWServerPB) PublishServiceTD() {
	if !pb.Config.PublishTD {
		return
	}
	deviceType := vocab.DeviceTypeService
	thingID := td.CreatePublisherThingID(pb.hubConfig.Zone, "hub", pb.Config.ClientID, deviceType)
	logrus.Infof("Publishing this service TD %s", thingID)
	thingTD := td.CreateTD(thingID, deviceType)
	// Include the service configuration properties
	prop := td.CreateProperty(vocab.PropNameAddress, "Gateway Address", vocab.PropertyTypeAttr)
	td.SetPropertyDataTypeString(prop, 0, 0)
	//
	td.AddTDProperty(thingTD, vocab.PropNameAddress, prop)
	td.SetThingDescription(thingTD, "EDS OWServer-V2 Protocol binding",
		"This service publishes information on The EDS OWServer 1-wire gateway and its connected sensors")
	pb.hubClient.PublishTD(thingID, thingTD)
}

// PublishThingsTD publishes the TD of Things
func (pb *OWServerPB) PublishTDs(tds map[string]td.ThingTD) error {
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
			tdCountDown = pb.Config.TDInterval
		}
		valueCountDown--
		if valueCountDown <= 0 {
			values, err := pb.PollValues()
			if err == nil {
				pb.PublishValues(values)
			}
			valueCountDown = pb.Config.ValueInterval
		}
		time.Sleep(time.Second)
	}
}

// Start connects to the hub internal message bus and starts polling
// the owserver.
func (pb *OWServerPB) Start(hubConfig *hubconfig.HubConfig) error {
	var err error
	pb.hubConfig = hubConfig
	// map of node thing info objects by thing ID
	pb.nodeInfo = make(map[string]*eds.OneWireNode)
	pb.edsAPI = eds.NewEdsAPI(pb.Config.EdsAddress, pb.Config.LoginName, pb.Config.Password)
	pb.hubClient = hubclient.NewMqttHubPluginClient(PluginID, hubConfig)
	err = pb.hubClient.Connect()
	if err != nil {
		logrus.Errorf("Protocol Binding for OWServer startup failed")
		return err
	}
	// publish the logger service thing
	pb.PublishServiceTD()

	pb.running = true
	go pb.heartbeat()
	logrus.Infof("Service OWServer startup completed")
	return nil
}

// Stop the service
func (pb *OWServerPB) Stop() {
	pb.running = false
	logrus.Info("Stopping service OWServer")
	pb.hubClient.Close()
}

// Create a new OWServer Protocol Binding service with default configuration
func NewOWServerPB() *OWServerPB {
	pb := &OWServerPB{}
	pb.Config = PluginConfig{
		ClientID:      PluginID,
		PublishTD:     false,
		TDInterval:    3600,
		ValueInterval: 60,
	}
	return pb
}
