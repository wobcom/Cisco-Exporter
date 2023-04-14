package local_pools

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

func ParsePool(sshCtx *connector.SSHCommandContext, poolOut chan *PoolGroup, poolParsingDone chan struct{}) {

	poolRegexp := regexp.MustCompile(` Pool`)
	listPoolRegexp := regexp.MustCompile(`^ (\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)$`)
	appendPoolRegexp := regexp.MustCompile(`^ \s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)$`)

	inList := false
	var openPoolGroup *PoolGroup

	for {
		select {
		case <-sshCtx.Done:
			if openPoolGroup != nil {
				poolOut <- openPoolGroup
				openPoolGroup = nil
			}
			poolParsingDone <- struct{}{}
			return
		case line := <-sshCtx.Output:
			if poolRegexp.MatchString(line) {
				inList = true
			} else if matches := appendPoolRegexp.FindStringSubmatch(line); inList && matches != nil {
				// openPoolGroup must be not nil at this point.
				freeNum := util.Str2float64(matches[3])
				assignedNum := util.Str2float64(matches[4])
				totalNum := freeNum + assignedNum

				newPool := Pool{
					StartIP:           matches[1],
					EndIP:             matches[2],
					AddressesTotal:    totalNum,
					AddressesAvail:    freeNum,
					AddressesAssigned: assignedNum,
				}
				openPoolGroup.Pools = append(openPoolGroup.Pools, newPool)

			} else if matches := listPoolRegexp.FindStringSubmatch(line); inList && matches != nil {

				if openPoolGroup != nil {
					poolOut <- openPoolGroup
					openPoolGroup = nil
				}

				freeNum := util.Str2float64(matches[4])
				assignedNum := util.Str2float64(matches[5])
				totalNum := freeNum + assignedNum

				openPoolGroup = &PoolGroup{
					Name: matches[1],

					Pools: []Pool{
						{
							StartIP:           matches[2],
							EndIP:             matches[3],
							AddressesTotal:    totalNum,
							AddressesAvail:    freeNum,
							AddressesAssigned: assignedNum,
						},
					},
				}

			}
		}
	}
}
