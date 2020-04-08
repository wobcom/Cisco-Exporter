package config

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const defaultConnectTimeout int = 5
const defaultCommandTimeout int = 20
const defaultPort int = 22

// Config provides means of reading the configuration file
type Config struct {
	Devices map[string]*DeviceConfig `yaml:"devices,omitempty"`
}

// OSVersion is a type to refere to the remote device's operating system.
// Different parsers / commands might be needed due to differences in the Cisco CLI.
type OSVersion int

const (
	// INVALID The remote device's OS could not be fingerprinted.
	INVALID OSVersion = 0
	// IOSXE The remote device is running IOS XE.
	IOSXE OSVersion = 1
	// IOS The remote device is running IOS.
	IOS OSVersion = 2
	// NXOS The remote device is running NX OS.
	NXOS OSVersion = 3
)

// DeviceConfig is used to read device configuration from the config file
// DeviceConfig describe how to connect to a remote device and what metrics
// to extract from the remote device.
type DeviceConfig struct {
	OSVersion         OSVersion
	Host              string
	Port              int      `yaml:"port,omitempty"`
	Username          string   `yaml:"username"`
	KeyFile           string   `yaml:"key_file,omitempty"`
	Password          string   `yaml:"password,omitempty"`
	ConnectTimeout    int      `yaml:"connect_timeout,omitempty"`
	CommandTimeout    int      `yaml:"command_timeout,omitempty"`
	EnabledCollectors []string `yaml:"enabled_collectors,flow"`
	Interfaces        []string `yaml:"interfaces,flow"`
}

func newConfig() *Config {
	config := &Config{
		Devices: make(map[string]*DeviceConfig, 0),
	}
	return config
}

// Load loads the configuration from the given reader.
func Load(reader io.Reader) (*Config, error) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	config := newConfig()
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, err
	}

	for hostName, device := range config.Devices {
		device.Host = hostName
		if device.ConnectTimeout == 0 {
			device.ConnectTimeout = defaultConnectTimeout
		}
		if device.CommandTimeout == 0 {
			device.CommandTimeout = defaultCommandTimeout
		}
		if device.Port == 0 {
			device.Port = defaultPort
		}
	}

	return config, nil
}
