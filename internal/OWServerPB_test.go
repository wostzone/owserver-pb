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

var pluginConfig *internal.PluginConfig = &internal.PluginConfig{} // use defaults
var hubConfig *hubconfig.HubConfig
var setupOnce = false

// --- THIS REQUIRES A RUNNING HUB OR MESSAGE BUS ---

// Use the project app folder during testing
func setup() {
	// if setupOnce {
	// 	return
	// }
	// setupOnce = true
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	// pluginConfig = &internal.PluginConfig{}
	// // remove VSCode testing arguments
	// os.Args = append(os.Args[0:1], strings.Split("", " ")...)
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, pluginID, pluginConfig)
	hubConfig.Messenger.CertsFolder = "/etc/mosquitto/certs"
}
func teardown() {
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")
	setup()
	svc := internal.OWServerPB{}
	err := svc.Start(hubConfig, pluginConfig)
	assert.NoError(t, err)
	svc.Stop()
	teardown()
}

func TestPollTDs(t *testing.T) {
	logrus.Infof("--- TestPollOnce ---")
	setup()
	os.Remove("../test/onewire-nodes.json")

	svc := internal.OWServerPB{}
	err := svc.Start(hubConfig, pluginConfig)
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
	setup()
	os.Remove("../test/onewire-nodes.json")

	svc := internal.OWServerPB{}
	err := svc.Start(hubConfig, pluginConfig)
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

	setup()
	// error cases - don't panic when polling without address
	os.Remove("../test/onewire-nodes.json")
	svc := internal.OWServerPB{}
	badConfig := *pluginConfig
	badConfig.EdsAddress = "http://invalidAddress/"
	// err := svc.Start(gwConfig, &badConfig)
	tds, err := svc.PollTDs()
	_ = tds
	assert.Error(t, err)
	teardown()

}
