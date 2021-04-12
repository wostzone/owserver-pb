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

	"github.com/wostzone/hubapi/pkg/certsetup"
	"github.com/wostzone/hubapi/pkg/hubclient"
	"github.com/wostzone/hubapi/pkg/hubconfig"
	"github.com/wostzone/hubapi/pkg/testenv"
	"github.com/wostzone/owserver/internal"
)

var homeFolder string
var hubConfig *hubconfig.HubConfig

const testPluginID = "owserver-test"

var mcmd *exec.Cmd

// Use the project test folder during testing
func setup() *internal.OWServerPB {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	mcmd = testenv.Setup(homeFolder, 0)

	os.Remove("../test/onewire-nodes.json")
	svc := internal.NewOWServerPB()

	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	return svc
}
func teardown() {
	testenv.Teardown(mcmd)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")
	svc := setup()
	err := svc.Start(hubConfig)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond)
	svc.Stop()
	teardown()
}

func TestPollTDs(t *testing.T) {
	var rxMsg []byte
	var rxThingID string

	logrus.Infof("--- TestPollOnce ---")

	svc := setup()
	err := svc.Start(hubConfig)
	assert.NoError(t, err)

	// listener should receive the TD
	// FIXME: consumer connection port should not be hidden
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Messenger.Address, hubConfig.Messenger.Port+1)
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
	teardown()
}
func TestPollValues(t *testing.T) {
	logrus.Infof("--- TestPollOnce ---")
	svc := setup()

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
	teardown()
}

func TestPollInvalidAddress(t *testing.T) {
	logrus.Infof("--- TestPollInvalidAddress ---")

	svc := setup()
	svc.Config.EdsAddress = "http://invalidAddress/"
	// err := svc.Start(gwConfig, &badConfig)
	tds, err := svc.PollTDs()
	_ = tds
	assert.Error(t, err)
	teardown()

}
