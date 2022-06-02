package internal

import (
	"crypto/tls"
	"crypto/x509"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/wost-go/pkg/exposedthing"

	"github.com/wostzone/owserver/internal/eds"
)

// PluginID is the default ID of this service. Used to name the configuration file
// and as the publisher ID portion of the Thing ID (zoneID:publisherID:deviceID:deviceType)
const PluginID = "owserver"

// OWServerPBConfig contains the plugin configuration
type OWServerPBConfig struct {
	// The service instance ID, default is the pluginID
	// Must be unique on the hub. Recommended is to add a '-1' in case of multiple instances.
	ClientID string `yaml:"clientID"`
	// OWServer address. Default is auto-discover using DNS-SD
	EdsAddress string `yaml:"owserverAddress,omitempty"`
	// Login to the EDS OWserver using Basic Auth.
	LoginName string `yaml:"loginName,omitempty"`
	Password  string `yaml:"password,omitempty"`
	// PrettyJSON for testing to improve readability of JSON output, default is False
	PrettyJSON bool `yaml:"prettyJSON,omitempty"`
	// PublishTD enables publish the TD of this service, default is False
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

	// Hub MQTT broker address and port to use for publishing TD and events
	mqttAddress string
	mqttPort    int

	// Hub MQTT client instance
	//hubClient *mqttclient.MqttClient

	// 1-wire nodes retrieved from the owserver gateway device
	// map of node/device ID to node info
	nodeInfo map[string]*eds.OneWireNode

	// Factory for creating exposed things
	eFactory *exposedthing.ExposedThingFactory

	// Map of node/device ID to exposed thing created for each published node
	eThings map[string]*exposedthing.ExposedThing

	// exposed thing of the service itself. nil if disabled
	serviceEThing *exposedthing.ExposedThing

	// flag, this service is up and running
	running bool
	mu      sync.Mutex

	// the zone of the plugin publications, default is local
	zone string
}

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

	err = pb.eFactory.Connect(pb.mqttAddress, pb.mqttPort)
	if err != nil {
		logrus.Errorf("Exposed Thing factory connection failed")
		return err
	}

	// Publish the OWServer service as a Thing
	if pb.Config.PublishTD {
		pb.serviceEThing = pb.CreateExposedThingForService()
	}

	// Periodic polling of the OWServer
	pb.running = true
	go pb.heartBeat()

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

		pb.eFactory.Disconnect()
	}
}

// NewOWServerPB creates a new OWServer Protocol Binding service with the provided configuration
func NewOWServerPB(config OWServerPBConfig, mqttAddress string, mqttPort int,
	caCert *x509.Certificate, pluginCert *tls.Certificate) *OWServerPB {

	// these are from hub configuration
	pb := &OWServerPB{
		mqttAddress: mqttAddress,
		mqttPort:    mqttPort,
		caCert:      caCert,
		pluginCert:  pluginCert,
		nodeInfo:    make(map[string]*eds.OneWireNode),
		eThings:     make(map[string]*exposedthing.ExposedThing),
		eFactory:    exposedthing.CreateExposedThingFactory(config.ClientID, pluginCert, caCert),
		running:     false,
	}
	pb.Config = config
	// ensure valid defaults
	if pb.Config.ClientID == "" {
		pb.Config.ClientID = PluginID
	}
	if pb.Config.TDInterval == 0 {
		pb.Config.TDInterval = 3600
	}
	if pb.Config.ValueInterval == 0 {
		pb.Config.ValueInterval = 30
	}

	// Create the adapter for the OWServer 1-wire gateway
	pb.edsAPI = eds.NewEdsAPI(config.EdsAddress, config.LoginName, config.Password)
	return pb
}
