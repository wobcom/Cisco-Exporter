package local_pools

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

func ParsePool(sshCtx *connector.SSHCommandContext, poolOut chan *Pool, poolParsingDone chan struct{}) {

	poolRegexp := regexp.MustCompile(` Pool`)
	listPoolRegexp := regexp.MustCompile(` (\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)`)

	inList := false

	for {
		select {
		case <-sshCtx.Done:
			poolParsingDone <- struct{}{}
			return
		case line := <-sshCtx.Output:
			if poolRegexp.MatchString(line) {
				inList = true
			} else if matches := listPoolRegexp.FindStringSubmatch(line); inList && matches != nil {
				freeNum := util.Str2float64(matches[4])
				assignedNum := util.Str2float64(matches[5])
				totalNum := freeNum + assignedNum

				pool := Pool{
					Name:              matches[1],
					AddressesTotal:    totalNum,
					AddressesAvail:    freeNum,
					AddressesAssigned: assignedNum,
				}

				poolOut <- &pool
			}
		}
	}
}
