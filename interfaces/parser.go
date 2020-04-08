package interfaces

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

// Parse parses cli output and tries to find interfaces with related stats
func Parse(sshCtx *connector.SSHCommandContext, interfaces chan *Interface, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	newIfRegexp := regexp.MustCompile(`(?:^!?(?: |admin|show|.+#).*$|^$)`)
	macRegexp := regexp.MustCompile(`^\s+Hardware(?: is|:) .+, address(?: is|:) (.*) \(.*\)$`)
	deviceNameRegexp := regexp.MustCompile(`^([a-zA-Z0-9\/\.-]+) is.*$`)
	adminStatusRegexp := regexp.MustCompile(`^.+ is (administratively)?\s*(up|down).*, line protocol is.*$`)
	adminStatusNXOSRegexp := regexp.MustCompile(`^\S+ is (up|down)(?:\s|,)?(\(Administratively down\))?.*$`)
	descRegexp := regexp.MustCompile(`^\s+Description: (.*)$`)
	dropsRegexp := regexp.MustCompile(`^\s+Input queue: \d+\/\d+\/(\d+)\/\d+ .+ Total output drops: (\d+)$`)
	inputBytesRegexp := regexp.MustCompile(`^\s+\d+ (?:packets input,|input packets)\s+(\d+) bytes.*$`)
	outputBytesRegexp := regexp.MustCompile(`^\s+\d+ (?:packets output,|output packets)\s+(\d+) bytes.*$`)
	inputErrorsRegexp := regexp.MustCompile(`^\s+(\d+) input error(?:s,)? .*$`)
	outputErrorsRegexp := regexp.MustCompile(`^\s+(\d+) output error(?:s,)? .*$`)
	speedRegexp := regexp.MustCompile(`^\s+(.*)-duplex,\s(\d+) ((\wb)/s).*$`)

	current := &Interface{}

	for {
		select {
		case <-sshCtx.Done:
			if current.Name != "" {
				interfaces <- current
			}
			return
		case line := <-sshCtx.Output:
			if !newIfRegexp.MatchString(line) {
				if current.Name != "" {
					interfaces <- current
				}
				matches := deviceNameRegexp.FindStringSubmatch(line)
				if matches == nil {
					continue
				}
				current = &Interface{
					Name: matches[1],
				}
			}
			if current.Name == "" {
				continue
			}

			if matches := adminStatusRegexp.FindStringSubmatch(line); matches != nil {
				if matches[1] == "" {
					current.AdminStatus = "up"
				} else {
					current.AdminStatus = "down"
				}
				current.OperStatus = matches[2]
			} else if matches := adminStatusNXOSRegexp.FindStringSubmatch(line); matches != nil {
				if matches[2] == "" {
					current.AdminStatus = "up"
				} else {
					current.AdminStatus = "down"
				}
				current.OperStatus = matches[1]
			} else if matches := descRegexp.FindStringSubmatch(line); matches != nil {
				current.Description = matches[1]
			} else if matches := macRegexp.FindStringSubmatch(line); matches != nil {
				current.MacAddress = matches[1]
			} else if matches := dropsRegexp.FindStringSubmatch(line); matches != nil {
				current.InputDrops = util.Str2float64(matches[1])
				current.OutputDrops = util.Str2float64(matches[2])
			} else if matches := inputBytesRegexp.FindStringSubmatch(line); matches != nil {
				current.InputBytes = util.Str2float64(matches[1])
			} else if matches := outputBytesRegexp.FindStringSubmatch(line); matches != nil {
				current.OutputBytes = util.Str2float64(matches[1])
			} else if matches := inputErrorsRegexp.FindStringSubmatch(line); matches != nil {
				current.InputErrors = util.Str2float64(matches[1])
			} else if matches := outputErrorsRegexp.FindStringSubmatch(line); matches != nil {
				current.OutputErrors = util.Str2float64(matches[1])
			} else if matches := speedRegexp.FindStringSubmatch(line); matches != nil {
				current.Speed = matches[2] + " " + matches[3]
			}
		}
	}
}
