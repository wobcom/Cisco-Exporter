package mpls

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

func parseForwardingTable(sshCtx *connector.SSHCommandContext, labelStatistics chan *LabelStatistic, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	newLabelSingleLineRegexp := regexp.MustCompile(`^(\d+)\s+(No Label|Pop Label|\d+)\s+(\S+)\s+(\d+)\s+(\S+)\s+(\S+)`)
	nextEntryRegexp := regexp.MustCompile(`^\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)`)
	newLabelMultilineRegexp := regexp.MustCompile(`^(\d+)\s+(No Label|Pop Label|\d+)\s+(\S+)\s+\\`)
	multilineLabelRegexp1 := regexp.MustCompile(`^\s+(\d+)\s+(\S+)\s+(\S+)`)
	multilineLabelRegexp2 := regexp.MustCompile(`^\s+(\d+)\s+(\S+)`)

	currentLocalLabel := "0"
	current := &LabelStatistic{}

	for {
		select {
		case <-sshCtx.Done:
			return
		case line := <-sshCtx.Output:
			if matches := newLabelSingleLineRegexp.FindStringSubmatch(line); matches != nil {
				currentLocalLabel = matches[1]
				labelStatistics <- &LabelStatistic{
					LocalLabel:         matches[1],
					OutgoingLabel:      matches[2],
					PrefixOrTunnelID:   matches[3],
					BytesLabelSwitched: util.Str2float64(matches[4]),
					OutgoingInterface:  matches[5],
					NextHop:            matches[6],
				}
			} else if nextEntryRegexp.FindStringSubmatch(line); matches != nil {
				labelStatistics <- &LabelStatistic{
					LocalLabel:         currentLocalLabel,
					OutgoingLabel:      matches[1],
					PrefixOrTunnelID:   matches[2],
					BytesLabelSwitched: util.Str2float64(matches[3]),
					OutgoingInterface:  matches[4],
					NextHop:            matches[5],
				}
			} else if newLabelMultilineRegexp.FindStringSubmatch(line); matches != nil {
				current = &LabelStatistic{
					LocalLabel:       matches[1],
					OutgoingLabel:    matches[2],
					PrefixOrTunnelID: matches[3],
				}
			} else if multilineLabelRegexp1.FindStringSubmatch(line); matches != nil {
				current.BytesLabelSwitched = util.Str2float64(matches[1])
				current.OutgoingInterface = matches[2]
				current.NextHop = matches[3]
				labelStatistics <- current
			} else if multilineLabelRegexp2.FindStringSubmatch(line); matches != nil {
				current.BytesLabelSwitched = util.Str2float64(matches[1])
				current.OutgoingInterface = matches[2]
				labelStatistics <- current
			}
		}
	}
}

func parseMemory(sshCtx *connector.SSHCommandContext, memoryStatistics chan *MemoryStatistic, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	memoryRegexp := regexp.MustCompile(`\s+(.*?)\s+:\s+(\d+)\/(\d+)`)

	for {
		select {
		case <-sshCtx.Done:
			return
		case line := <-sshCtx.Output:
			if matches := memoryRegexp.FindStringSubmatch(line); matches != nil {
				memoryStatistics <- &MemoryStatistic{
					AllocatorName: matches[1],
					InUse:         util.Str2float64(matches[2]),
					Allocated:     util.Str2float64(matches[3]),
				}
			}
		}
	}
}
