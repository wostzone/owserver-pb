package internal

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"github.com/wostzone/owserver-pb/internal/eds"
)

// PluginID is the default ID of this service. Used to name the configuration file
// and as the publisher ID portion of the Thing ID (zoneID:publisherID:deviceID:deviceType)
const PluginID = "owserver-pb"

// OWServerPBConfig contains the plugin configuration
type OWServerPBConfig struct {
	// The Thing publisher-id, default is the pluginID
	// Must be unique on the hub. Recommended is to add a '-1' in case of multiple instances.
	ClientID string `yaml:"clientID"`
	// OWServer address. Default is auto-discover using DNS-SD
	EdsAddress string `yaml:"owserverAddress,omitempty"`
	// Login to the EDS OWserver using Basic Auth.
	LoginName string `yaml:"loginName,omitempty"`
	Password  string `yaml:"password,omitempty"`
	// publish the TD of this service, default is False
	PublishTD bool `yaml:"publishTD,omitempty"`
	// interval of republishing the full TD, default is 1 hours
	TDInterval int `yaml:"tdInterval,omitempty"`
	// interval of republishing modified Thing property values, default is 60 seconds
	ValueInterval int `yaml:"valueInterval,omitempty"`
}

// OWServerPB is the hub protocol binding plugin for capturing 1-wire OWServer V2 Data
type OWServerPB struct {
	// Configuration of this protocol binding
	Config OWServerPBConfig

	// EDS OWServer client API
	edsAPI *eds.EdsAPI

	// Hub CA certificate to validate client connections
	caCert *x509.Certificate

	// Client certificate of this service
	pluginCert *tls.Certificate

	// Hub MQTT broker address:port to use for publishing TD and events
	mqttHostPort string

	// Hub MQTT client instance
	hubClient *mqttclient.MqttClient

	// 1-wire nodes retrieved from the owserver gateway device
	// map of node/device ID to node info
	nodeInfo map[string]*eds.OneWireNode

	// Map of node/device ID to exposed thing created for each published node
	eThings map[string]*mqttbinding.MqttExposedThing

	// flag, this service is up and running
	running bool
	mu      sync.Mutex

	// the zone of the plugin publications, default is local
	zone string
}

// heartbeat polls the EDS server every X seconds and updates the Exposed TD's
func (pb *OWServerPB) heartbeat() {
	var tdCountDown = 0
	var valueCountDown = 0
	for {
		pb.mu.Lock()
		isRunning := pb.running
		pb.mu.Unlock()
		if !isRunning {
			break
		}

		tdCountDown--
		if tdCountDown <= 0 {
			// create ExposedThing's as they are discovered
			_ = pb.PollTDs()
			tdCountDown = pb.Config.TDInterval
		}
		valueCountDown--
		if valueCountDown <= 0 {
			pb.UpdatePropertyValues()
			valueCountDown = pb.Config.ValueInterval
		}
		time.Sleep(time.Second)
	}
}

// PublishServiceTD publishes the Thing Description document of the service itself
// This is only published if 'publishTD' is set in the configuration
// The publisher of this TD is the hub with the deviceID the plugin-ID
// TD attributes of this service includes are:
//    'address' - gateway address
func (pb *OWServerPB) PublishServiceTD() {
	if !pb.Config.PublishTD {
		return
	}
	deviceType := vocab.DeviceTypeService
	thingID := thing.CreatePublisherID(pb.zone, pb.Config.ClientID, pb.Config.ClientID, deviceType)
	logrus.Infof("Publishing this service TD %s", thingID)

	// Create the TD document for this protocol binding
	tdoc := thing.CreateTD(thingID, "OWServer Service", deviceType)
	tdoc.UpdateTitleDescription("EDS OWServer-V2 Protocol binding",
		"This service publishes information on The EDS OWServer 1-wire gateway and its connected sensors")

	// Include the service properties (attributes and configuration)
	tdoc.AddProperty(vocab.PropNameAddress, "Gateway Address", vocab.WoTDataTypeString)
	eThing := mqttbinding.CreateExposedThing(tdoc, pb.hubClient)
	eThing.SetPropertyWriteHandler("", func(propName string, value mqttbinding.InteractionOutput) error {
		//
		return nil
	})
	err := eThing.Expose()
	if err != nil {
		logrus.Errorf("PublishServiceTD: Error publishing service TD: %s", err)
	}
}

//// PublishTDs publishes all the TD of Things
//func (pb *OWServerPB) PublishTDs(tds map[string]thing.ThingTD) error {
//	var err error
//	for thingID, td := range tds {
//		err = pb.hubClient.PublishTD(thingID, td)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

// Start the OWServer protocol binding
// This:
//   1. connects to the hub message bus
//   2. publish this service as a Thing as its own publisher
//   3. periodic poll the OWServer gateway for metadata and values of 1-wire devices
//   	a. create a TD and an exposed thing for each 1-wire device connected to the OWServer gateway
//      b. expose (publish) the TD of newly added or modified exposed things
//      c. publish the values of 1-wire devices via the exposed thing
func (pb *OWServerPB) Start() error {
	var err error

	// Connect using the MQTT protocol
	// Todo consideration: move transport protocol into ExposedThing factory
	pb.hubClient = mqttclient.NewMqttClient(pb.Config.ClientID, pb.caCert, 0)
	err = pb.hubClient.ConnectWithClientCert(pb.mqttHostPort, pb.pluginCert)
	if err != nil {
		logrus.Errorf("Protocol Binding for OWServer startup failed")
		return err
	}

	// Publish the OWServer service as a Thing
	pb.PublishServiceTD()

	// Periodic polling of the OWServer
	pb.running = true
	go pb.heartbeat()

	logrus.Infof("Service OWServer startup completed")
	return nil
}

// Stop the service
func (pb *OWServerPB) Stop() {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if pb.running {
		pb.running = false

		logrus.Info("Stopping service OWServer")
		// FIXME, wait until discovery has completed if running
		time.Sleep(time.Second)

		pb.hubClient.Close()
	}
}

// NewOWServerPB creates a new OWServer Protocol Binding service with the provided configuration
func NewOWServerPB(config OWServerPBConfig, mqttHostPort string,
	caCert *x509.Certificate, pluginCert *tls.Certificate) *OWServerPB {

	// these are from hub configuration
	pb := &OWServerPB{
		mqttHostPort: mqttHostPort,
		caCert:       caCert,
		pluginCert:   pluginCert,
		nodeInfo:     make(map[string]*eds.OneWireNode),
		eThings:      make(map[string]*mqttbinding.MqttExposedThing),
		running:      false,
	}
	pb.Config = config
	// ensure valid defaults
	if config.ClientID == "" {
		config.ClientID = PluginID
	}
	if config.TDInterval == 0 {
		config.TDInterval = 3600
	}
	if config.ValueInterval == 0 {
		config.ValueInterval = 60
	}

	// Create the adapter for the OWServer 1-wire gateway
	pb.edsAPI = eds.NewEdsAPI(config.EdsAddress, config.LoginName, config.Password)
	return pb
}
