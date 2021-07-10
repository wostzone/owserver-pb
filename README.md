# WoST OWServerV2 Protocol Binding

WoST OwServerV2 protocol binding is a hub plugin that reads one-wire sensor data from the EDS OwServer-v2 hub and publishes its TD and events onto the hub.

## Objective

Convert EDS OWServer 1-wire devices to WoT Things.

## Status 

The status of this plugin is Alpha. It is functional but breaking changes must be expected.

This plugin provides basic functionality:
1. Publish TDs for the EDS gateway and connected 1-wire devices
2. Publish value update messages

Planned:
3. Support for configuration updates
4. Improve conformance to WoT TD's with publishing updates and use of TD forms


## Audience

This project is aimed at software developers, system implementors and people with a keen interest in IoT that value security and interoperability.

## Dependencies

This plugin needs a EDS OWServer hub device on the local network. 

## Summary

With this plugin 1-wire devices connected to a OWServer V2 hub can be discovered and accessed via the WoST hub like any other WoST Thing.

This is a relative simple plugin that can serve as an example on writing plugins.


## Build and Installation

### System Requirements

This plugin runs as part of the WoST hub. It has no additional requirements other than a working hub.


### Manual Installation

See the hub README on plugin installation.


### Build From Source

Build with:

```
make all
```
The plugin can be found in dist/bin. Copy this to the hub bin directory.
An example configuration file is provided in config/owserver.yaml. Copy this to the hub config directory.


## Usage

Configure the owserver.yaml configuration file with the EDS OWServer V2 hub address and login credentials and restart the hub.

The EDS hub itself and the 1-wire devices that are connected can be found through the directory service. 
