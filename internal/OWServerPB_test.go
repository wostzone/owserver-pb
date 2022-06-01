package internal_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/wostzone/wost-go/pkg/consumedthing"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/mqttclient"
	"github.com/wostzone/wost-go/pkg/testenv"
	"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/vocab"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/owserver-pb/internal"
)

//var homeFolder string
const testPluginID = "owserver-test"

var tempFolder string
var mqttHostPort string
var testCerts testenv.TestCerts

//var hubConfig *config.HubConfig
var owsConfig internal.OWServerPBConfig

// var testCerts testenv.TestCerts
var owsSimulationFile string // simulation file
var mosquittoCmd *exec.Cmd

// TestMain run mosquitto and use the project test folder as the home folder.
// All tests are run using the simulation file.
func TestMain(m *testing.M) {
	// setup environment
	tempFolder = path.Join(os.TempDir(), "wost-owserver-test")
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../testdata")
	owsSimulationFile = "file://" + path.Join(homeFolder, "owserver-details.xml")
	logging.SetLogging("info", "")
	testCerts = testenv.CreateCertBundle()
	testenv.SaveCerts(&testCerts, tempFolder)

	// load the plugin config with client cert
	//hubConfig = config.CreateHubConfig(homeFolder)
	//hubConfig.Load("", internal.PluginID)
	mqttHostPort = fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortCert)
	owsConfig.ClientID = testPluginID
	owsConfig.EdsAddress = owsSimulationFile

	// run the test mosquitto server. Use only certificate authentication
	mosquittoCmd, _ = testenv.StartMosquitto(&testCerts, tempFolder)
	if mosquittoCmd == nil {
		logrus.Fatalf("Unable to setup mosquitto")
	}

	//os.Remove("../test/onewire-nodes.json")

	result := m.Run()
	time.Sleep(time.Second)
	testenv.StopMosquitto(mosquittoCmd, "")
	mosquittoCmd.Process.Kill()

	if result == 0 {
		os.RemoveAll(tempFolder)
	}

	os.Exit(result)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")
	var rxMsg []byte
	var rxTopic string
	var rxMutex = sync.Mutex{}

	// listen for TDs
	testClient := mqttclient.NewMqttClient(testPluginID+"-client", testCerts.CaCert, 0)
	err := testClient.ConnectWithClientCert(mqttHostPort, testCerts.PluginCert)
	require.NoError(t, err)

	serviceThingID := thing.CreatePublisherID(
		"", owsConfig.ClientID, owsConfig.ClientID, vocab.DeviceTypeService)
	serviceTopic := consumedthing.CreateTopic(serviceThingID, consumedthing.TopicTypeTD)

	testClient.Subscribe(serviceTopic, func(topic string, message []byte) {
		logrus.Infof("TestStartStop: received message for thingID: %s", topic)
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rxMsg = message
		rxTopic = topic
	})
	// startup
	svc := internal.NewOWServerPB(owsConfig,
		testenv.ServerAddress, testenv.MqttPortCert, testCerts.CaCert, testCerts.PluginCert)
	svc.Config.PublishTD = true
	err = svc.Start()
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// This should publish the Thing of the service
	rxMutex.Lock()
	assert.NotEmpty(t, rxTopic, "Did not receive a message")
	assert.NotEmpty(t, rxMsg, "Did not receive message data")
	rxMutex.Unlock()

	svc.Stop()
}

func TestPollTDs(t *testing.T) {
	var tdCount int = 0
	var rxMutex = sync.Mutex{}

	logrus.Infof("--- TestPollTDs ---")
	svc := internal.NewOWServerPB(owsConfig,
		testenv.ServerAddress, testenv.MqttPortCert, testCerts.CaCert, testCerts.PluginCert)
	assert.NotNil(t, svc)

	// Count the number of received TDs
	testClient := mqttclient.NewMqttClient(testPluginID+"-client", testCerts.CaCert, 0)
	err := testClient.ConnectWithClientCert(mqttHostPort, testCerts.PluginCert)
	assert.NoError(t, err)
	tdTopics := consumedthing.CreateTopic("+", consumedthing.TopicTypeTD)
	testClient.Subscribe(tdTopics, func(thingID string, message []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		tdCount++
	})
	time.Sleep(time.Second)

	// start the service which publishes TDs
	err = svc.Start()
	assert.NoError(t, err)

	err = svc.PollProperties()
	assert.NoError(t, err)
	//err = svc.PublishTDs(tds)
	//assert.NoError(t, err)

	time.Sleep(time.Millisecond * 500)
	testClient.Disconnect()
	// the simulation file contains 3 things. The service is 1 thing.
	rxMutex.Lock()
	assert.GreaterOrEqual(t, tdCount, 4)
	rxMutex.Unlock()

	svc.Stop()
}

func TestPollValues(t *testing.T) {
	logrus.Infof("--- TestPollOnce ---")
	var eventCount int = 0
	var rxMutex = sync.Mutex{}

	svc := internal.NewOWServerPB(owsConfig,
		testenv.ServerAddress, testenv.MqttPortCert, testCerts.CaCert, testCerts.PluginCert)
	assert.NotNil(t, svc)

	// Count the number of received value events
	testClient := mqttclient.NewMqttClient(testPluginID+"-client", testCerts.CaCert, 0)
	err := testClient.ConnectWithClientCert(mqttHostPort, testCerts.PluginCert)
	assert.NoError(t, err)
	eventTopics := consumedthing.CreateTopic("+", consumedthing.TopicTypeEvent) + "/#"
	testClient.Subscribe(eventTopics, func(thingID string, message []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		eventCount++
	})
	time.Sleep(time.Second)

	// start the heartbeat that publishes changes to property values
	err = svc.Start()
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// the simulation file contains 3 things + service is 4 property events + non property events
	rxMutex.Lock()
	assert.GreaterOrEqual(t, eventCount, 4)
	rxMutex.Unlock()

	svc.Stop()
}

func TestPollValuesNotInitialized(t *testing.T) {
	logrus.Infof("--- TestPollValuesNotInitialized ---")

	svc := internal.NewOWServerPB(owsConfig,
		testenv.ServerAddress, testenv.MqttPortCert, testCerts.CaCert, testCerts.PluginCert)
	_, err := svc.PollValues()
	require.Error(t, err)
	err = svc.PollProperties()
	require.Error(t, err)
}

// func TestPollValuesBadAddres(t *testing.T) {
// 	logrus.Infof("--- TestPollValuesBadAddres ---")

// 	svc := internal.NewOWServerPB(testPluginID, mqttHostPort, hubConfig.CaCert, hubConfig.PluginCert)
// 	// some address that is incorrect
// 	svc.Config.EdsAddress = "192.168.0.123"
// 	err := svc.Start()
// 	assert.NoError(t, err)
// 	_, err = svc.PollValues()
// 	require.Error(t, err)
// }
func TestPollInvalidEDSAddress(t *testing.T) {
	logrus.Infof("--- TestPollInvalidEDSAddress ---")

	owsConfig.EdsAddress = "http://invalidAddress/"
	svc := internal.NewOWServerPB(owsConfig,
		testenv.ServerAddress, testenv.MqttPortCert, testCerts.CaCert, testCerts.PluginCert)
	assert.NotNil(t, svc)

	err := svc.Start()
	assert.NoError(t, err)

	err = svc.PollProperties()
	assert.Error(t, err)
	svc.Stop()

}

func TestPublishServiceTD(t *testing.T) {
	logrus.Infof("--- TestPublishServiceTD ---")

	svc := internal.NewOWServerPB(owsConfig,
		testenv.ServerAddress, testenv.MqttPortCert, testCerts.CaCert, testCerts.PluginCert)
	svc.Config.PublishTD = true
	err := svc.Start()
	assert.NoError(t, err)
	// svc.PublishServiceTD()
	svc.Stop()

}

func TestPublishServiceTDBadAddress(t *testing.T) {
	logrus.Infof("--- TestPublishServiceTD ---")

	svc := internal.NewOWServerPB(owsConfig,
		"badmqttaddress", 22, testCerts.CaCert, testCerts.PluginCert)
	svc.Config.PublishTD = true
	err := svc.Start()
	assert.Error(t, err)
	values, err := svc.PollValues()
	assert.Error(t, err)
	err = svc.PublishValues(values)
	assert.Error(t, err)
	svc.Stop()

}
