# WoST OWServerV2 Protocol Binding

WoST OwServerV2 protocol binding is a gateway plugin that reads one-wire sensor data from the EDS OwServer-v2 gateway and publishes its TD and events onto the gateway.

## Objective

Convert EDS OWServer 1-wire devices to WoT Things.

## Status 

The status of this plugin is In Development.

This plugin is not yet functional.

## Audience

This project is aimed at software developers, system implementors and people with a keen interest in IoT. 

## Dependencies

This plugin needs a EDS OWServer gateway device on the local network. 

## Summary

With this plugin 1-wire devices connected to a OWServer V2 gateway can be discovered and accessed via the WoST gateway like any other WoST Thing.

This is a relative simple plugin that can serve as an example on writing plugins.


## Build and Installation

### System Requirements

This plugin runs as part of the WoST gateway. It has no additional requirements other than a working gateway.


### Manual Installation

See the gateway README on plugin installation.


### Build From Source

Build with:

```
make all
```
The plugin can be found in dist/bin for 64bit intel or amd processors, or dist/arm for 64 bit ARM processors. Copy this to the gateway bin or arm directory.
An example configuration file is provided in config/owserver.yaml. Copy this to the gateway config directory.


## Usage

Configure the owserver.yaml configuration file with the EDS OWServer V2 gateway address and login credentials and restart the gateway.

The EDS gateway itself and the 1-wire devices that are connected can be found through the directory service. 
