package connector

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"gitlab.com/wobcom/cisco-exporter/config"

	"github.com/prometheus/common/log"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const defaultPort = 22
const defaultTimeout = 5

// Option is a function called to configure the SSHConnectionManager
type Option func(*SSHConnectionManager)

// WithReconnectInterval allows specifying a reconnect interval
func WithReconnectInterval(duration time.Duration) Option {
	return func(connectionManager *SSHConnectionManager) {
		connectionManager.reconnectInterval = duration
	}
}

// WithKeepAliveInterval allows specifying a keep alive interval
func WithKeepAliveInterval(duration time.Duration) Option {
	return func(connectionManager *SSHConnectionManager) {
		connectionManager.keepAliveInterval = duration
	}
}

// WithKeepAliveTimeout allows specifying a keep alive timeout
func WithKeepAliveTimeout(duration time.Duration) Option {
	return func(connectionManager *SSHConnectionManager) {
		connectionManager.keepAliveTimeout = duration
	}
}

// SSHConnectionManager provides means of establishing and maintaining an SSH Connection to a remote deivce.
// SSH Connections are intentionally left open as long as possible, to reduce the number of logged logins as well as load on the TACACS server and the remote device.
type SSHConnectionManager struct {
	connections       map[string]*SSHConnection
	connectionsMutex  sync.Mutex
	reconnectInterval time.Duration
	keepAliveInterval time.Duration
	keepAliveTimeout  time.Duration
	mutexesMutex      sync.Mutex
	mutexes           map[string]sync.Mutex
}

// NewConnectionManager applies the specified options and returns a new SSHConnectionManager
func NewConnectionManager(options ...Option) *SSHConnectionManager {
	connectionManager := &SSHConnectionManager{
		connections:       make(map[string]*SSHConnection),
		mutexes:           make(map[string]sync.Mutex),
		reconnectInterval: 30 * time.Second,
		keepAliveInterval: 15 * time.Second,
		keepAliveTimeout:  15 * time.Second,
	}

	for _, option := range options {
		option(connectionManager)
	}

	return connectionManager
}

// GetConnection returns an SSHConnection to the given device.
// If the connection has not yet been established (or lost), establishing a connection is attempted.
// In case of error nil and the error are returned.
func (connMan *SSHConnectionManager) GetConnection(device *config.DeviceConfig) (*SSHConnection, error) {
	connMan.mutexesMutex.Lock()
	_, found := connMan.mutexes[device.Host]
	if !found {
		connMan.mutexes[device.Host] = sync.Mutex{}
	}
	mutex, _ := connMan.mutexes[device.Host]
	connMan.mutexesMutex.Unlock()

	mutex.Lock()
	defer mutex.Unlock()

	connMan.connectionsMutex.Lock()
	connection, found := connMan.connections[device.Host]
	connMan.connectionsMutex.Unlock()
	var err error = nil

	if found {
		if !connection.IsConnected() {
			log.Errorf("Connection to '%s' was lost, reconnecting.", device.Host)
			connection, err = connMan.establishConnection(device)
		} else if !connection.IsAuthenticated() {
			log.Errorf("Connection to '%s' is no longer authenticated, reconnecting.", device.Host)
			connection.Terminate()
			connection, err = connMan.establishConnection(device)
		} else {
			return connection, nil
		}
		if err != nil {
			return nil, err
		}
		connMan.connectionsMutex.Lock()
		connMan.connections[device.Host] = connection
		connMan.connectionsMutex.Unlock()
	} else {
		connection, err = connMan.establishConnection(device)
		if err != nil {
			return nil, err
		}
		connMan.connectionsMutex.Lock()
		connMan.connections[device.Host] = connection
		connMan.connectionsMutex.Unlock()
	}
	if connection == nil {
		log.Error("bleep hang")
	}
	return connection, err
}

func (connMan *SSHConnectionManager) establishConnection(device *config.DeviceConfig) (*SSHConnection, error) {
	sshClient, transportConnection, err := connMan.makeSSHClient(device)
	if err != nil {
		return nil, err
	}

	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open a new session for '%s'", device.Host)
	}

	stdin, _ := sshSession.StdinPipe()
	stdout, _ := sshSession.StdoutPipe()

	terminalModes := ssh.TerminalModes{
		ssh.ECHO:  0,
		ssh.OCRNL: 0,
	}

	sshSession.RequestPty("vt100", 0, 2000, terminalModes)
	sshSession.Shell()

	sshConnection := &SSHConnection{
		transportConnection: transportConnection,
		sshClient:           sshClient,
		Device:              device,
		done:                make(chan struct{}),
		stdin:               stdin,
		stdout:              stdout,
	}
	go connMan.keepAlive(sshConnection)

	err = sshConnection.DisablePagination()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not disable pagination on '%s': %v", device.Host, err)
	}
	device.OSVersion, err = sshConnection.IdentifyOSVersion()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not identify os version on '%s': %s", device.Host, err)
	}

	log.Infof("Established an SSH connection with '%s'", device.Host)
	return sshConnection, nil
}

func (connMan *SSHConnectionManager) makeSSHClient(device *config.DeviceConfig) (*ssh.Client, net.Conn, error) {
	clientConfig, err := connMan.makeSSHConfig(device)
	if err != nil {
		return nil, nil, err
	}

	transportConnection, err := net.DialTimeout("tcp", net.JoinHostPort(device.Host, strconv.Itoa(device.Port)), time.Duration(device.ConnectTimeout)*time.Second)
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("Could not connect to device '%s'", device.Host))
	}

	c, chans, reqs, err := ssh.NewClientConn(transportConnection, device.Host, clientConfig)
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("Could not establish SSH connection with '%s'", device.Host))
	}

	client := ssh.NewClient(c, chans, reqs)
	return client, transportConnection, nil
}

func (connMan *SSHConnectionManager) makeSSHConfig(device *config.DeviceConfig) (*ssh.ClientConfig, error) {
	if device.Port == 0 {
		device.Port = defaultPort
	}

	var config ssh.ClientConfig
	config.SetDefaults()

	auth, err := makeAuth(device)
	config.User = device.Username
	config.Auth = auth
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	config.Ciphers = append(config.Ciphers, "aes128-cbc", "aes256-cbc", "3des-cbc")
	return &config, err
}

func makeAuth(device *config.DeviceConfig) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	if device.KeyFile != "" {
		keyFile, err := os.Open(device.KeyFile)
		if err != nil {
			return nil, err
		}
		defer keyFile.Close()

		keyFileContents, err := ioutil.ReadAll(keyFile)
		if err != nil {
			return nil, errors.Wrapf(err, "Error reading private key file '%s'", device.KeyFile)
		}

		key, err := ssh.ParsePrivateKey(keyFileContents)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing private key file '%s'", device.KeyFile)
		}

		authMethods = append(authMethods, ssh.PublicKeys(key))
	}

	if device.Password != "" {
		authMethods = append(authMethods, ssh.Password(device.Password))
	}

	if len(authMethods) == 0 {
		return nil, errors.New(fmt.Sprintf("I don't know how to authenticate with '%s'", device.Host))
	}

	return authMethods, nil
}

func (connMan *SSHConnectionManager) keepAlive(connection *SSHConnection) {
	for {
		select {
		case <-time.After(connMan.keepAliveInterval):
			if !connection.IsConnected() {
				break
			}
			connection.transportConnection.SetDeadline(time.Now().Add(connMan.keepAliveTimeout))
			ctx := NewSSHCommandContext("")
			ctx.IgnoreOutputs()
			connection.RunCommand(ctx)
		case <-connection.done:
			return
		}
	}
}
