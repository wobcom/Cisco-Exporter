package connector

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"gitlab.com/wobcom/cisco-exporter/config"

	"github.com/pkg/errors"

	"golang.org/x/crypto/ssh"

	"github.com/prometheus/common/log"
)

// SSHConnection wraps an *ssh.Client and provides functions for executing commands on the remote device.
type SSHConnection struct {
	stdout              io.Reader
	stdin               io.WriteCloser
	sshClient           *ssh.Client
	mu                  sync.Mutex
	transportConnection net.Conn
	Device              *config.DeviceConfig
	done                chan struct{}
}

// SSHCommandContext provides context for running a command on the remote device.
// Lines received from the remote device are written to the Output chan.
// Errors encoutered during executing a command are written to the Erros chan.
// Once the command finished execution, an empty struct is written to the Done chan.
type SSHCommandContext struct {
	Command string
	Output  chan string
	Errors  chan error
	Done    chan struct{}
	Timeout int
}

// NewSSHCommandContext initializes the channels and returns a new SSHCommandContext.
func NewSSHCommandContext(command string) *SSHCommandContext {
	return &SSHCommandContext{
		Command: command,
		Output:  make(chan string),
		Errors:  make(chan error),
		Done:    make(chan struct{}, 1),
	}
}

// IgnoreOutputs ignores the outputs received from an SSHCommandContext and logs erros to the CLI.
func (ctx *SSHCommandContext) IgnoreOutputs() {
	go func() {
		for {
			select {
			case <-ctx.Done:
				return
			case <-ctx.Output:
				continue
			case err := <-ctx.Errors:
				log.Errorf("Ignoring error: %v", err)
			}
		}
	}()
}

// IsConnected returns whether the SSHConnection is still up and the remote end connected.
func (conn *SSHConnection) IsConnected() bool {
	return conn.transportConnection != nil
}

// IsAuthenticated tests if the authentication for this SSH Session has not yet expired
func (conn *SSHConnection) IsAuthenticated() bool {
	sshCtx := NewSSHCommandContext("")
	sshCtx.Timeout = 2
	go conn.RunCommand(sshCtx)

	var lastErr error = nil

	for {
		select {
		case <-sshCtx.Done:
			return lastErr == nil
		case lastErr = <-sshCtx.Errors:
			continue
		case <-sshCtx.Output:
			continue
		}
	}
}

// IdentifyOSVersion fingerprints the remote operating system.
func (conn *SSHConnection) IdentifyOSVersion() (config.OSVersion, error) {
	sshCtx := NewSSHCommandContext("show version")
	sshCtx.Timeout = 2
	go conn.RunCommand(sshCtx)

	var lastErr error = nil

	detectedVersion := config.INVALID
	versions := map[string]config.OSVersion{
		"IOS XE":       config.IOSXE,
		"NX-OS":        config.NXOS,
		"IOS Software": config.IOS,
	}

	for {
		select {
		case <-sshCtx.Done:
			return detectedVersion, lastErr
		case line := <-sshCtx.Output:
			for fingerprint, version := range versions {
				if strings.Contains(line, fingerprint) && detectedVersion == config.INVALID {
					detectedVersion = version
				}
			}
		case lastErr = <-sshCtx.Errors:
			continue
		}
	}
}

// DisablePagination disables the paginator on the remote end.
// This is required to parse the whole output of a command.
// Note that for `terminal length 0` certain privileges are required on the remote device.
func (conn *SSHConnection) DisablePagination() error {
	sshCtx := NewSSHCommandContext("terminal shell\nterminal length 0")
	sshCtx.Timeout = 2
	var lastErr error = nil
	go conn.RunCommand(sshCtx)

	for {
		select {
		case lastErr = <-sshCtx.Errors:
			continue
		case <-sshCtx.Done:
			return lastErr
		case <-sshCtx.Output:
			continue
		}
	}
}

// RunCommand runs a command on the remote device. All events / outputs are received to the provided context.
func (conn *SSHConnection) RunCommand(ctx *SSHCommandContext) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	defer func() {
		ctx.Done <- struct{}{}
	}()

	if ctx.Timeout == 0 {
		ctx.Timeout = conn.Device.CommandTimeout
	}

	if conn.transportConnection == nil {
		ctx.Errors <- errors.New(fmt.Sprintf("Cannot run command '%s' on target '%s': Not connected.", ctx.Command, conn.Device.Host))
		return
	}

	reader := bufio.NewReader(conn.stdout)
	if ctx.Command == "" {
		ctx.Command = "show clock\n"
	} else {
		ctx.Command = ctx.Command + " ; show clock\n"
	}
	io.WriteString(conn.stdin, ctx.Command)

	errorChan := make(chan error)
	scannerDone := make(chan struct{}, 1)
	abortSignal := false
	go conn.scanLines(scannerDone, &abortSignal, reader, ctx.Output, errorChan)

	select {
	case err := <-errorChan:
		ctx.Errors <- errors.Wrapf(err, "Error reading from stdout: %v", err)
		abortSignal = true
		conn.terminate()
		return
	case <-scannerDone:
		return
	case <-time.After(time.Duration(ctx.Timeout) * time.Second):
		ctx.Errors <- errors.New(fmt.Sprintf("Timeout reached for '%s' on %s", ctx.Command, conn.Device.Host))
		abortSignal = true
		conn.terminate()
		return
	}
}

func (conn *SSHConnection) scanLines(done chan struct{}, abortSignal *bool, reader *bufio.Reader, output chan<- string, errorChan chan error) {
	defer func() {
		done <- struct{}{}
		reader.Reset(nil)
	}()
	cmdLineRegex := regexp.MustCompile(`^\.?\d{2}:\d{2}:\d{2}.\d{3}`)
	authenticationRegexp := regexp.MustCompile(`[aA]uthentication [eE]xpired`)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				errorChan <- errors.Wrapf(err, "Scanner received err: %v", err)
			} else {
				errorChan <- errors.New("Scanner reached EOF")
			}
			return
		}
		line := scanner.Text()
		if cmdLineRegex.MatchString(line) {
			return
		} else if authenticationRegexp.MatchString(line) {
			errorChan <- errors.New("Authentication Expired")
			return
		}
		if *abortSignal {
			return
		}
		output <- line
	}
}

// Terminate terminates the SSHConnection
func (conn *SSHConnection) Terminate() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.terminate()
}

func (conn *SSHConnection) terminate() {
	if conn.transportConnection == nil {
		return
	}
	conn.sshClient.Close()
	conn.transportConnection.Close()
	conn.sshClient = nil
	conn.transportConnection = nil
}
