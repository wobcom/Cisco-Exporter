package vlans

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

func (c *Collector) parse(sshCtx *connector.SSHCommandContext, vlans chan *VLANInterface, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	deviceNameRegexp, _ := regexp.Compile(`([a-zA-Z0-9\/-]+\.\d+) \(\d+\)`)
	inputBytesRegexp, _ := regexp.Compile(`Total \d+ packets, (\d+) bytes input`)
	outputBytesRegexp, _ := regexp.Compile(`Total \d+ packets, (\d+) bytes output`)

	current := &VLANInterface{}
	for {
		select {
		case <-sshCtx.Done:
			if current.Name != "" {
				vlans <- current
			}
			return
		case line := <-sshCtx.Output:
			if matches := deviceNameRegexp.FindStringSubmatch(line); matches != nil {
				if current.Name != "" {
					vlans <- current
				}
				current = &VLANInterface{
					Name: matches[1],
				}
			}
			if current.Name == "" {
				continue
			}
			if matches := inputBytesRegexp.FindStringSubmatch(line); matches != nil {
				current.InputBytes = util.Str2float64(matches[1])
			} else if matches := outputBytesRegexp.FindStringSubmatch(line); matches != nil {
				current.OutputBytes = util.Str2float64(matches[1])
			}
		}
	}
}
