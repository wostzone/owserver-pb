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

	"github.com/wostzone/hubapi-go/pkg/certsetup"
	"github.com/wostzone/hubapi-go/pkg/hubclient"
	"github.com/wostzone/hubapi-go/pkg/hubconfig"
	"github.com/wostzone/hubapi-go/pkg/testenv"
	"github.com/wostzone/owserver-pb/internal"
)

var homeFolder string
var hubConfig *hubconfig.HubConfig

const testPluginID = "owserver-test"

var mosquittoCmd *exec.Cmd

// TestMain run mosquitto and use the project test folder as the home folder.
// Make sure the certificates exist.
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	mosquittoCmd = testenv.Setup(homeFolder, 0)
	if mosquittoCmd == nil {
		logrus.Fatalf("Unable to setup mosquitto")
	}
	os.Remove("../test/onewire-nodes.json")

	result := m.Run()
	testenv.Teardown(mosquittoCmd)

	os.Exit(result)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")
	svc := internal.NewOWServerPB()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	err := svc.Start(hubConfig)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond)
	svc.Stop()
}

func TestPollTDs(t *testing.T) {
	var rxMsg []byte
	var rxThingID string

	logrus.Infof("--- TestPollOnce ---")

	svc := internal.NewOWServerPB()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	err := svc.Start(hubConfig)
	assert.NoError(t, err)

	// listener should receive the TD
	// FIXME: consumer connection port should not be hidden
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Messenger.Address, hubConfig.Messenger.ClientPortMqtt)
	caCertFile := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	consumer := hubclient.NewHubClient(hostPort, caCertFile, "test-client", "")
	err = consumer.Start()
	assert.NoError(t, err)
	consumer.Subscribe("", func(thingID string, msgType string, message []byte, senderID string) {
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
	assert.NotEmpty(t, rxThingID, "Did not receive a message")
	assert.NotEmpty(t, rxMsg, "Did not receive message data")

	time.Sleep(3 * time.Second)
}

func TestPollValues(t *testing.T) {
	logrus.Infof("--- TestPollOnce ---")
	svc := internal.NewOWServerPB()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)

	err := svc.Start(hubConfig)
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

	time.Sleep(3 * time.Second)
}

func TestPollInvalidAddress(t *testing.T) {
	logrus.Infof("--- TestPollInvalidAddress ---")

	svc := internal.NewOWServerPB()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	svc.Config.EdsAddress = "http://invalidAddress/"
	// err := svc.Start(gwConfig, &badConfig)
	tds, err := svc.PollTDs()
	_ = tds
	assert.Error(t, err)

}
