package internal_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/hubclient-go/pkg/config"
	"github.com/wostzone/hubclient-go/pkg/mqttclient"
	"github.com/wostzone/hubclient-go/pkg/testenv"
	"github.com/wostzone/owserver-pb/internal"
)

var homeFolder string

// const zone = "test"
const testPluginID = "owserver-test"

var mqttHostPort string

var hubConfig *config.HubConfig

// var testCerts testenv.TestCerts

var mosquittoCmd *exec.Cmd

// TestMain run mosquitto and use the project test folder as the home folder.
func TestMain(m *testing.M) {
	// setup environment
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	certsFolder := path.Join(homeFolder, config.DefaultCertsFolder)
	testCerts := testenv.CreateCertBundle()
	testenv.SaveCerts(&testCerts, certsFolder)

	// load the plugin config with client cert
	hubConfig = config.CreateDefaultHubConfig(homeFolder)
	config.LoadHubConfig("", internal.PluginID, hubConfig)
	mqttHostPort = fmt.Sprintf("%s:%d", hubConfig.MqttAddress, hubConfig.MqttPortCert)

	// run the test mosquitto server. Use only certificate authentication
	mosquittoCmd, _ = testenv.StartMosquitto(hubConfig.ConfigFolder, hubConfig.CertsFolder, &testCerts)
	if mosquittoCmd == nil {
		logrus.Fatalf("Unable to setup mosquitto")
	}

	os.Remove("../test/onewire-nodes.json")

	result := m.Run()
	time.Sleep(time.Second)
	mosquittoCmd.Process.Kill()

	os.Exit(result)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")

	// svcConfig := internal.PluginConfig{}
	// hubConfig, err := config.LoadConfig(nil, homeFolder, "plugin", &svcConfig)
	// assert.NoError(t, err)

	svc := internal.NewOWServerPB(testPluginID, mqttHostPort, hubConfig.CaCert, hubConfig.PluginCert)

	err := svc.Start()
	assert.NoError(t, err)
	svc.Stop()
}

func TestPollTDs(t *testing.T) {
	var rxMsg []byte
	var rxThingID string

	logrus.Infof("--- TestPollOnce ---")

	svc := internal.NewOWServerPB(testPluginID, mqttHostPort, hubConfig.CaCert, hubConfig.PluginCert)
	assert.NotNil(t, svc)

	err := svc.Start()
	assert.NoError(t, err)

	// listener should receive the TD
	// FIXME: consumer connection port should not be hidden
	// hostPort := fmt.Sprintf("%s:%d", hubConfig.Messenger.Address, hubConfig.Messenger.CertPortMqtt)
	// caCertFile := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	testClient := mqttclient.NewMqttHubClient("testplugin", hubConfig.CaCert)
	err = testClient.ConnectWithClientCert(mqttHostPort, hubConfig.PluginCert)
	assert.NoError(t, err)
	testClient.Subscribe("", func(thingID string, msgType string, message []byte, senderID string) {
		rxMsg = message
		rxThingID = thingID
	})
	time.Sleep(time.Second)

	// svc.Start(gwConfig, pluginConfig)
	tds, err := svc.PollTDs()
	assert.NoError(t, err)
	err = svc.PublishTDs(tds)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 100)
	testClient.Close()
	assert.NotEmpty(t, rxThingID, "Did not receive a message")
	assert.NotEmpty(t, rxMsg, "Did not receive message data")

	svc.Stop()
}

func TestPollValues(t *testing.T) {
	logrus.Infof("--- TestPollOnce ---")

	svc := internal.NewOWServerPB(testPluginID, mqttHostPort, hubConfig.CaCert, hubConfig.PluginCert)
	assert.NotNil(t, svc)

	err := svc.Start()
	assert.NoError(t, err)

	// Get and publish the Things
	tds, err := svc.PollTDs()
	require.NoError(t, err)
	err = svc.PublishTDs(tds)
	assert.NoError(t, err)

	// Get and publish Thing values
	values, err := svc.PollValues()
	require.NoError(t, err)
	err = svc.PublishValues(values)
	assert.NoError(t, err)

	svc.Stop()
}

func TestPollValuesNotInitialized(t *testing.T) {
	logrus.Infof("--- TestPollValuesNotInitialized ---")

	svc := internal.NewOWServerPB(testPluginID, mqttHostPort, hubConfig.CaCert, hubConfig.PluginCert)
	_, err := svc.PollValues()
	require.Error(t, err)
	_, err = svc.PollTDs()
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
func TestPollInvalidAddress(t *testing.T) {
	logrus.Infof("--- TestPollInvalidAddress ---")

	svc := internal.NewOWServerPB(testPluginID, mqttHostPort, hubConfig.CaCert, hubConfig.PluginCert)
	assert.NotNil(t, svc)

	svc.Config.EdsAddress = "http://invalidAddress/"
	err := svc.Start()
	assert.NoError(t, err)

	tds, err := svc.PollTDs()
	_ = tds
	assert.Error(t, err)
	svc.Stop()

}

func TestPublishServiceTD(t *testing.T) {
	logrus.Infof("--- TestPublishServiceTD ---")

	svc := internal.NewOWServerPB(testPluginID, mqttHostPort, hubConfig.CaCert, hubConfig.PluginCert)
	svc.Config.PublishTD = true
	err := svc.Start()
	assert.NoError(t, err)
	// svc.PublishServiceTD()
	svc.Stop()

}

func TestPublishServiceTDBadAddress(t *testing.T) {
	logrus.Infof("--- TestPublishServiceTD ---")

	svc := internal.NewOWServerPB(testPluginID, "badmqtt:port", hubConfig.CaCert, hubConfig.PluginCert)
	svc.Config.PublishTD = true
	err := svc.Start()
	assert.Error(t, err)
	values, err := svc.PollValues()
	assert.NoError(t, err)
	err = svc.PublishValues(values)
	assert.Error(t, err)
	svc.Stop()

}
