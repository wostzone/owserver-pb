package eds_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/owserver-pb/internal/eds"
)

// Some tests require a living OWServer
const edsAddress = ""

// simulation file for testing without OWServer gateway
const owserverSimulation = "../../test/owserver-details.xml"

func TestDiscover(t *testing.T) {
	edsAPI := eds.NewEdsAPI("", "", "")
	addr, err := edsAPI.Discover()
	assert.NoError(t, err)
	assert.NotEmpty(t, addr, "EDS OWserver V2 not found")
}

// Read EDS test data from file
func TestReadEdsFromFile(t *testing.T) {
	edsAPI := eds.NewEdsAPI("file://"+owserverSimulation, "", "")
	rootNode, err := edsAPI.ReadEds()
	assert.NoError(t, err)
	require.NotNil(t, rootNode, "Expected root node")
	assert.True(t, len(rootNode.Nodes) == 20, "Expected 20 parameters and nested")
}

// Read EDS test data from file
func TestReadEdsFromInvalidFile(t *testing.T) {
	// error case, unknown file
	edsAPI := eds.NewEdsAPI("file://../doesnotexist.xml", "", "")
	rootNode, err := edsAPI.ReadEds()
	assert.Error(t, err)
	assert.Nil(t, rootNode, "Did not expect root node")
}

// Read EDS device and check if more than 1 node is returned. A minimum of 1 is expected if the
// device is online with an additional node for each connected node.
// NOTE: This requires a live hub on the 'edsAddress'
func TestReadEdsFromHub(t *testing.T) {

	// NOTE: This requires a live hub on the 'edsAddress'
	edsAPI := eds.NewEdsAPI(edsAddress, "", "")
	rootNode, err := edsAPI.ReadEds()

	assert.NoError(t, err, "Failed reading EDS hub")
	require.NotNil(t, rootNode, "Expected root node")
	assert.GreaterOrEqual(t, len(rootNode.Nodes), 3, "Expected at least 3 nodes")
}
func TestReadEdsFromInvalidAddress(t *testing.T) {

	// error case - bad hub
	// error case, unknown file
	edsAddress := "doesnoteexist"
	edsAPI := eds.NewEdsAPI(edsAddress, "", "")
	rootNode, err := edsAPI.ReadEds()
	assert.Error(t, err)
	assert.Nil(t, rootNode)
}

// Parse the nodes xml file and test for correct results
func TestParseNodeFile(t *testing.T) {
	// remove cached nodes first
	os.Remove("../../test/onewire-nodes.json")
	edsAddress := "file://../../test/owserver-details.xml"
	edsAPI := eds.NewEdsAPI(edsAddress, "", "")

	rootNode, err := edsAPI.ReadEds()
	require.NoError(t, err)
	require.NotNil(t, rootNode)

	// The test file has hub parameters and 3 connected nodes
	deviceNodes := edsAPI.ParseOneWireNodes(rootNode, 0, true)
	assert.Lenf(t, deviceNodes, 4, "Expected 4 nodes")
}
