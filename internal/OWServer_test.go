package internal_test

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/gateway/pkg/config"
	"github.com/wostzone/gateway/pkg/smbserver"
	"github.com/wostzone/owserver/internal"
)

var homeFolder string

const pluginID = "owserver-test"

var pluginConfig *internal.PluginConfig = &internal.PluginConfig{} // use defaults
var gwConfig *config.GatewayConfig
var setupOnce = false

// Use the project app folder during testing
func setup() {
	if setupOnce {
		return
	}
	setupOnce = true
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../dist")
	pluginConfig = &internal.PluginConfig{}
	os.Args = append(os.Args[0:1], strings.Split("", " ")...)
	gwConfig, _ = config.SetupConfig(homeFolder, pluginID, pluginConfig)
}
func teardown() {
}

func TestStartStop(t *testing.T) {
	setup()
	server, err := smbserver.StartSmbServer(gwConfig)
	require.NoError(t, err)

	svc := internal.NewOWServer()
	err = svc.Start(gwConfig, pluginConfig)
	assert.NoError(t, err)
	svc.Stop()
	server.Stop()
	teardown()
}

func TestPollOnce(t *testing.T) {
	setup()
	os.Remove("../test/onewire-nodes.json")

	// pub, err := publisher.NewAppPublisher(AppID, configFolder, appConfig, "", false)
	// pub.SetSigningOnOff(false)
	// if !assert.NoError(t, err) {
	// 	return
	// }
	svc := internal.NewOWServer()
	// svc.Start(gwConfig, pluginConfig)
	err := svc.Poll()
	_ = err
	// assert.NoError(t, err)
	time.Sleep(3 * time.Second)
	teardown()

}
func TestPollInvalidAddress(t *testing.T) {
	setup()
	// error cases - don't panic when polling without address
	os.Remove("../test/onewire-nodes.json")
	svc := internal.NewOWServer()
	badConfig := *pluginConfig
	badConfig.EdsAddress = "http://invalidAddress/"
	// err := svc.Start(gwConfig, &badConfig)
	err := svc.Poll()
	assert.Error(t, err)
	teardown()

}
