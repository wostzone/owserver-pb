package internal_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/owserver/internal"
)

// Read EDS test data from file
func TestReadEdsFromFile(t *testing.T) {
	edsAPI := internal.NewEdsAPI("file://../test/owserver-details.xml", "", "")
	rootNode, err := edsAPI.ReadEds()
	assert.NoError(t, err)
	require.NotNil(t, rootNode, "Expected root node")
	assert.True(t, len(rootNode.Nodes) == 20, "Expected 20 parameters and nested")
}

// Read EDS test data from file
func TestReadEdsFromInvalidFile(t *testing.T) {
	// error case, unknown file
	edsAPI := internal.NewEdsAPI("file://../doesnotexist.xml", "", "")
	rootNode, err := edsAPI.ReadEds()
	assert.Error(t, err)
	assert.Nil(t, rootNode, "Did not expect root node")

}

// Read EDS device and check if more than 1 node is returned. A minimum of 1 is expected if the
// device is online with an additional node for each connected node.
// NOTE: This requires a live gateway on the 'edsAddress'
func TestReadEdsFromGateway(t *testing.T) {

	edsAddress := "10.3.3.33"
	edsAPI := internal.NewEdsAPI(edsAddress, "", "")
	rootNode, err := edsAPI.ReadEds()

	assert.NoError(t, err, "Failed reading EDS gateway")
	require.NotNil(t, rootNode, "Expected root node")
	assert.GreaterOrEqual(t, len(rootNode.Nodes), 3, "Expected at least 3 nodes")

}
func TestReadEdsFromInvalidAddress(t *testing.T) {

	// error case - bad gateway
	// error case, unknown file
	edsAddress := "doesnoteexist"
	edsAPI := internal.NewEdsAPI(edsAddress, "", "")
	rootNode, err := edsAPI.ReadEds()
	assert.Error(t, err)
	assert.Nil(t, rootNode)
}

// Parse the nodes xml file and test for correct results
func TestParseNodeFile(t *testing.T) {
	// remove cached nodes first
	os.Remove("../test/onewire-nodes.json")
	edsAddress := "file://../test/owserver-details.xml"
	edsAPI := internal.NewEdsAPI(edsAddress, "", "")

	rootNode, err := edsAPI.ReadEds()
	require.NoError(t, err)
	require.NotNil(t, rootNode)

	// The test file has gateway parameters and 3 connected nodes
	gwParams, deviceNodes := edsAPI.ParseNodeParams(rootNode)
	assert.Len(t, gwParams, 17, "Expected multiple gateway parameters")
	assert.Lenf(t, deviceNodes, 3, "Expected 3 gateway nodes")
}
