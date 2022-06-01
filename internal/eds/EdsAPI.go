// Package eds with EDS OWServer API methods
package eds

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wostzone/wost-go/pkg/vocab"

	"github.com/sirupsen/logrus"
)

// family to device type. See also: http://owfs.sourceforge.net/simple_family.html
// Todo: get from config file so it is easy to update
var deviceTypeMap = map[string]vocab.DeviceType{
	"10": vocab.DeviceTypeThermometer,
	"28": vocab.DeviceTypeThermometer,
	"7E": vocab.DeviceTypeMultisensor,
}

// AttrVocab maps OWServer attribute names to IoT vocabulary
var AttrVocab = map[string]string{
	"MACAddress": vocab.PropNameMAC,
	//"DateTime":   vocab.PropNameDateTime,
	"DeviceName": vocab.PropNameName,
	"HostName":   vocab.PropNameHostname,
	// Exclude/ignore the following attributes as they are chatty and not useful
	"DateTime":     "",
	"RawData":      "",
	"Counter1":     "",
	"Counter2":     "",
	"PollCount":    "",
	"PrimaryValue": "",
}

// SensorTypeVocab maps OWServer sensor names to IoT vocabulary
var SensorTypeVocab = map[string]struct {
	name     string
	dataType string
	decimals int // number of decimals accuracy for this value
}{
	// "BarometricPressureHg": vocab.PropNameAtmosphericPressure, // unit Hg
	"BarometricPressureMb": {name: vocab.PropNameAtmosphericPressure,
		dataType: vocab.WoTDataTypeNumber, decimals: 0}, // unit Mb
	"DewPoint": {name: vocab.PropNameDewpoint,
		dataType: vocab.WoTDataTypeNumber, decimals: 1},
	"HeatIndex": {name: vocab.PropNameHeatIndex,
		dataType: vocab.WoTDataTypeNumber, decimals: 1},
	"Humidity": {name: vocab.PropNameHumidity,
		dataType: vocab.WoTDataTypeNumber, decimals: 0},
	"Humidex": {name: vocab.PropNameHumidex,
		dataType: vocab.WoTDataTypeNumber, decimals: 1},
	"Light": {name: vocab.PropNameLuminance,
		dataType: vocab.WoTDataTypeNumber, decimals: 0},
	"RelayState": {name: vocab.PropNameRelay,
		dataType: vocab.WoTDataTypeBool, decimals: 0},
	"Temperature": {name: vocab.PropNameTemperature,
		dataType: vocab.WoTDataTypeNumber, decimals: 1},
}

// UnitNameVocab maps OWServer unit names to IoT vocabulary
var UnitNameVocab = map[string]string{
	"PercentRelativeHumidity": vocab.UnitNamePercent,
	"Millibars":               vocab.UnitNameMillibar,
	"Centigrade":              vocab.UnitNameCelcius,
	"Fahrenheit":              vocab.UnitNameFahrenheit,
	"InchesOfMercury":         vocab.UnitNameMercury,
	"Lux":                     vocab.UnitNameLux,
	"#":                       vocab.UnitNameCount,
	"Volt":                    vocab.UnitNameVolt,
}

// EdsAPI EDS device API properties and methods
type EdsAPI struct {
	address         string     // EDS (IP) address or filename (file://./path/to/name.xml)
	loginName       string     // Basic Auth login name
	password        string     // Basic Auth password
	discoTimeoutSec int        // EDS OWServer discovery timeout
	readMutex       sync.Mutex // prevent concurrent discovery
}

// XMLNode XML parsing node. Pure magic...
//--- https://stackoverflow.com/questions/30256729/how-to-traverse-through-xml-data-in-golang
type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []XMLNode  `xml:",any"`
	// Possible attributes for subnodes, depending on the property name
	Description string `xml:"Description,attr"`
	Writable    string `xml:"Writable,attr"`
	Units       string `xml:"Units,attr"`
}

// OneWireAttr with info on each node attribute
type OneWireAttr struct {
	Name     string
	Unit     string
	Writable bool
	Value    string
	IsSensor bool // sensors emit events on change
}

// OneWireNode with info on each node
type OneWireNode struct {
	DeviceType vocab.DeviceType
	// ThingID     string
	NodeID      string // ROM ID
	Name        string
	Description string
	Attr        map[string]OneWireAttr // attribute by name
}

// Apply the vocabulary to the name
// This returns the translated name from the vocabulary or the original name if not in the vocabulary
func applyVocabulary(name string, vocab map[string]string) (vocabName string, hasName bool) {
	vocabName, hasName = vocab[name]
	if !hasName {
		vocabName = name
	}
	return vocabName, hasName
}

// LookupEdsName returns the EDS name of a sensor from the vocabulary name.
// If the name is not a vocabulary name then return the original name.
// intended for executing an action or writing a configuration.
// @param name is the standardized vocabulary for the property or sensor name
func LookupEdsName(name string) string {
	for edsName, sensorInfo := range SensorTypeVocab {
		if sensorInfo.name == name {
			return edsName
			break
		}
	}
	return name
}

// Discover any EDS OWServer ENet-2 on the local network for 3 seconds
// This uses a UDP Broadcast on port 30303 as stated in the manual
// If found, this sets the service address for further use
// Returns the address or an error if not found
func (edsAPI *EdsAPI) Discover() (addr string, err error) {
	logrus.Infof("Starting discovery")
	var addr2 *net.UDPAddr
	// listen
	pc, err := net.ListenPacket("udp4", ":30303")
	if err == nil {
		defer pc.Close()

		addr2, err = net.ResolveUDPAddr("udp4", "255.255.255.255:30303")
	}
	if err == nil {
		_, err = pc.WriteTo([]byte("D"), addr2)
	}
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024)
	// receive 2 messages, first the broadcast, followed by the response, if there is one
	// wait 3 seconds before giving up
	for {
		pc.SetReadDeadline(time.Now().Add(time.Second * time.Duration(edsAPI.discoTimeoutSec)))
		n, remoteAddr, err := pc.ReadFrom(buf)
		if err != nil {
			logrus.Infof("Discovery ended without results")
			return "", err
		} else if n > 1 {
			switch rxAddr := remoteAddr.(type) {
			case *net.UDPAddr:
				addr = rxAddr.IP.String()
				logrus.Infof("EdsAPI.Discover. Found at %s: %s", addr, buf[:n])
				return addr, nil
			}
		}
	}
}

// GetLastAddress returns the last used address of the gateway
// This is either the configured or the discovered address
func (edsAPI *EdsAPI) GetLastAddress() string {
	return edsAPI.address
}

// ParseOneWireNodes parses the owserver xml data and returns a list of nodes,
// including the owserver gateway, and their parameters.
// This also converts sensor values to a proper decimals. Eg temperature isn't 4 digits but 1.
//  xmlNode is the node to parse, its attribute and possibly subnodes
//  latency to add to the root node (gateway device)
//  isRootNode is set for the first node, eg the gateway itself
func (edsAPI *EdsAPI) ParseOneWireNodes(xmlNode *XMLNode, latency time.Duration, isRootNode bool) []*OneWireNode {
	owNodeList := make([]*OneWireNode, 0)

	owNode := OneWireNode{
		// ID:          xmlNode.Attrs["ROMId"],
		Name:        xmlNode.XMLName.Local,
		Description: xmlNode.Description,
		Attr:        make(map[string]OneWireAttr),
		DeviceType:  vocab.DeviceTypeGateway,
	}
	owNodeList = append(owNodeList, &owNode)
	// todo: find a better place for this
	if isRootNode {
		owAttr := OneWireAttr{
			Name:  vocab.PropNameLatency,
			Value: fmt.Sprintf("%.3f", latency.Seconds()),
			Unit:  "sec",
		}
		owNode.Attr[owAttr.Name] = owAttr
	}
	// parse attributes and round sensor values
	for _, node := range xmlNode.Nodes {
		// if the xmlnode has no subnodes then it is a parameter describing the current node
		if len(node.Nodes) == 0 {
			// standardize the naming of properties and property types
			writable := (strings.ToLower(node.Writable) == "true")
			attrName := node.XMLName.Local
			sensorInfo, isSensor := SensorTypeVocab[attrName]
			decimals := -1 // -1 means no conversion
			if isSensor {
				// this is a known sensor type. (writable sensors are actuators)
				attrName = sensorInfo.name
				decimals = sensorInfo.decimals
			} else {
				// this is an attribute. writable attributes are configuration
				attrName, _ = applyVocabulary(attrName, AttrVocab)
			}
			if attrName != "" {
				unit, _ := applyVocabulary(node.Units, UnitNameVocab)
				valueStr := string(node.Content)
				valueFloat, err := strconv.ParseFloat(valueStr, 32)
				// rounding of sensor values to decimals
				if err == nil && decimals >= 0 {
					ratio := math.Pow(10, float64(decimals))
					valueFloat = math.Round(valueFloat*ratio) / ratio
					valueStr = strconv.FormatFloat(valueFloat, 'f', decimals, 32)
				}

				owAttr := OneWireAttr{
					Name:     attrName,
					Value:    valueStr,
					Unit:     unit,
					IsSensor: isSensor,
					Writable: writable,
				}
				owNode.Attr[owAttr.Name] = owAttr
				// Family is used to determine device type, default is gateway
				if node.XMLName.Local == "Family" {
					deviceType := deviceTypeMap[owAttr.Value]
					if deviceType == "" {
						deviceType = vocab.DeviceTypeUnknown
					}
					owNode.DeviceType = deviceType
				} else if node.XMLName.Local == "ROMId" {
					// all subnodes use the ROMId as its ID
					owNode.NodeID = owAttr.Value
				} else if isRootNode && node.XMLName.Local == "DeviceName" {
					// The gateway itself uses the deviceName as its ID and name
					owNode.NodeID = owAttr.Value
					owNode.Name = owAttr.Value
					owNode.Description = "EDS OWServer Gateway"
				}
			}
		} else {
			// The node contains subnodes which contain one or more sensors.
			subNodes := edsAPI.ParseOneWireNodes(&node, 0, false)
			owNodeList = append(owNodeList, subNodes...)
		}
	}
	// owNode.ThingID = td.CreatePublisherThingID(pb.hubConfig.Zone, PluginID, owNode.NodeID, owNode.DeviceType)

	return owNodeList
}

// PollValues polls the OWServer gateway for Thing property values
// Returns a map of device/node ID's containing a map of property name:value pairs
// eg: map[nodeID](map[propName]propValue)
func (edsAPI *EdsAPI) PollValues() (map[string](map[string]interface{}), error) {
	logrus.Infof("EdsAPI.PollValues")

	// thingValues is a map of NodeID:{attr:value,...}
	thingValues := make(map[string](map[string]interface{}))

	// Read the values from the EDS gateway
	startTime := time.Now()
	rootNode, err := edsAPI.ReadEds()
	endTime := time.Now()
	latency := endTime.Sub(startTime)
	if err != nil {
		return nil, err
	}
	// Extract the nodes and convert properties to vocab names
	nodeList := edsAPI.ParseOneWireNodes(rootNode, latency, true)
	for _, node := range nodeList {
		propValues := make(map[string]interface{})
		for name, attr := range node.Attr {
			propValues[name] = attr.Value
		}
		thingValues[node.NodeID] = propValues
	}
	return thingValues, nil
}

// ReadEds reads EDS gateway and return the result as an XML node
// If edsAPI.address starts with file:// then read from file, otherwise from http
// If no address is configured, one will be auto discovered the first time.
func (edsAPI *EdsAPI) ReadEds() (rootNode *XMLNode, err error) {
	// don't discover or read concurrently
	edsAPI.readMutex.Lock()
	defer edsAPI.readMutex.Unlock()
	if edsAPI.address == "" {
		edsAPI.address, err = edsAPI.Discover()
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(edsAPI.address, "file://") {
		filename := edsAPI.address[7:]
		buffer, err := ioutil.ReadFile(filename)
		if err != nil {
			logrus.Errorf("Unable to read EDS file from %s: %v", filename, err)
			return nil, err
		}
		err = xml.Unmarshal(buffer, &rootNode)
		return rootNode, err
	}
	// not a file, continue with http request
	edsURL := "http://" + edsAPI.address + "/details.xml"
	req, _ := http.NewRequest("GET", edsURL, nil)

	req.SetBasicAuth(edsAPI.loginName, edsAPI.password)
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(req)

	// resp, err := http.Get(edsURL)
	if err != nil {
		logrus.Errorf("Unable to read EDS gateway from %s: %v", edsURL, err)
		return nil, err
	}
	// Decode the EDS response into XML
	dec := xml.NewDecoder(resp.Body)
	err = dec.Decode(&rootNode)
	_ = resp.Body.Close()

	return rootNode, err
}

// UnmarshalXML parse xml
func (n *XMLNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node XMLNode

	return d.DecodeElement((*node)(n), &start)
}

// WriteData writes a value to a variable
// this posts a request to devices.html?rom={romID}&variable={variable}&value={value}
func (edsAPI *EdsAPI) WriteData(romID string, variable string, value string) error {
	// TODO: auto config if this is http or https
	writeURL := "http://" + edsAPI.address + "/devices.htm" +
		"?rom=" + romID + "&variable=" + variable + "&value=" + value
	req, _ := http.NewRequest("GET", writeURL, nil)

	logrus.Infof("EdsAPI.WriteData: URL: %s", writeURL)
	req.SetBasicAuth(edsAPI.loginName, edsAPI.password)
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(req)
	_ = resp

	if err != nil {
		logrus.Errorf("EdsAPI.WriteData: Unable to write data to EDS gateway at %s: %v", writeURL, err)
	}
	return err
}

// NewEdsAPI creates a new NewEdsAPI instance
//  address is optional to override the discovery
//  loginName if needed, "" if not needed
//  password if needed, "" if not needed
func NewEdsAPI(address string, loginName string, password string) *EdsAPI {
	edsAPI := &EdsAPI{
		address:         address,
		loginName:       loginName,
		password:        password,
		discoTimeoutSec: 3, // discovery timeout
	}
	return edsAPI
}
