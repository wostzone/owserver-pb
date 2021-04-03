package internal_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/hubapi/pkg/hubconfig"
	"github.com/wostzone/owserver/internal"
)

var homeFolder string

const pluginID = "owserver-test"

var hubConfig *hubconfig.HubConfig
var setupOnce = false

// --- THIS REQUIRES A RUNNING HUB OR MESSAGE BUS ---

// Use the project app folder during testing
func setup() *internal.OWServerPB {
	os.Remove("../test/onewire-nodes.json")

	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	svc := internal.NewOWServerPB()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, pluginID, &svc.Config)
	hubConfig.Messenger.CertsFolder = "/etc/mosquitto/certs"
	return svc
}
func teardown() {
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
	logrus.Infof("--- TestPollOnce ---")

	svc := setup()
	err := svc.Start(hubConfig)
	require.NoError(t, err)

	// svc.Start(gwConfig, pluginConfig)
	tds, err := svc.PollTDs()
	require.NoError(t, err)
	err = svc.PublishTDs(tds)
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)
	teardown()
}
func TestPollValues(t *testing.T) {
	logrus.Infof("--- TestPollOnce ---")
	svc := setup()

	err := svc.Start(hubConfig)
	require.NoError(t, err)

	// svc.Start(gwConfig, pluginConfig)
	tds, err := svc.PollTDs()
	require.NoError(t, err)
	err = svc.PublishTDs(tds)
	assert.NoError(t, err)

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
