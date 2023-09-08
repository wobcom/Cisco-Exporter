package config

import (
	"github.com/gobwas/glob"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const defaultConnectTimeout int = 5
const defaultCommandTimeout int = 20
const defaultPort int = 22

// Config provides means of reading the configuration file
type Config struct {
	DeviceGroups map[string]*DeviceGroupConfig `yaml:"devices,omitempty"`
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

// GetAllOsVersions returns all known and valid os version
func GetAllOsVersions() []OSVersion {
	return []OSVersion{IOSXE, IOS, NXOS}
}

// OSVersionToString converts OSVersion to a string
func (o OSVersion) String() string {
	mapping := map[OSVersion]string{
		IOSXE: "ios-xe",
		IOS:   "ios",
		NXOS:  "nxos",
	}
	name, found := mapping[o]
	if found {
		return name
	}
	return "unknown/invalid"
}

// DeviceGroupConfig is used to read device configuration from the config file
// DeviceGroupConfig describe how to connect to a remote device and what metrics
// to extract from the remote device.
type DeviceGroupConfig struct {
	OSVersion         OSVersion
	StaticName        *string   `yaml:"-"`
	Matcher           glob.Glob `yaml:"-"`
	Port              int       `yaml:"port,omitempty"`
	Username          string    `yaml:"username"`
	KeyFile           string    `yaml:"key_file,omitempty"`
	Password          string    `yaml:"password,omitempty"`
	ConnectTimeout    int       `yaml:"connect_timeout,omitempty"`
	CommandTimeout    int       `yaml:"command_timeout,omitempty"`
	EnabledCollectors []string  `yaml:"enabled_collectors,flow"`
	Interfaces        []string  `yaml:"interfaces,flow"`
	EnabledVLANs      []string  `yaml:"enabled_vlans,flow"`
}

func newConfig() *Config {
	config := &Config{
		DeviceGroups: make(map[string]*DeviceGroupConfig, 0),
	}
	return config
}

func (c *Config) GetDeviceGroup(device string) *DeviceGroupConfig {
	for _, config := range c.DeviceGroups {
		if config.Matcher == nil {
			continue
		}

		if config.Matcher.Match(device) {
			return config
		}
	}

	return nil

}

func (c *Config) GetStaticDevices() []string {
	staticDeviceNames := make([]string, 0)

	for _, config := range c.DeviceGroups {
		if config.StaticName != nil {
			staticDeviceNames = append(staticDeviceNames, *config.StaticName)
		}
	}

	return staticDeviceNames
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

	for matchStr, groupConfig := range config.DeviceGroups {

		groupConfig.Matcher = glob.MustCompile(matchStr)

		// A glob is static, if there are no special meta signs to quote.
		// Therefore, QuoteMeta should be a no op for static strings.
		if glob.QuoteMeta(matchStr) == matchStr {
			s := matchStr
			groupConfig.StaticName = &s
		}

		if groupConfig.ConnectTimeout == 0 {
			groupConfig.ConnectTimeout = defaultConnectTimeout
		}
		if groupConfig.CommandTimeout == 0 {
			groupConfig.CommandTimeout = defaultCommandTimeout
		}
		if groupConfig.Port == 0 {
			groupConfig.Port = defaultPort
		}
	}

	return config, nil
}
