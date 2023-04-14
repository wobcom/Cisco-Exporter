package local_pools_test

import (
	"gitlab.com/wobcom/cisco-exporter/local_pools"
	"strings"
	"testing"

	"gitlab.com/wobcom/cisco-exporter/connector"
)

func inputContext() connector.SSHCommandContext {
	const input = `

 PoolGroup                     Begin           End             Free  In use
 IP-JAIL                  172.16.0.60     172.16.0.119      56       4
                          172.16.0.30     172.16.0.55       22       4
 IP-CGNAT-CISCO           100.65.128.0    100.65.191.255  6809    9575
 IP-PUBLIC                5.159.24.0      5.159.24.255     153     103
                          5.159.25.0      5.159.25.255     140     116
                          5.159.26.0      5.159.26.255     130     126

`

	outputChan := make(chan string)
	errChan := make(chan error)
	doneChan := make(chan struct{})

	go func() {
		for _, line := range strings.Split(strings.TrimSuffix(input, "\n"), "\n") {
			outputChan <- line
		}
		doneChan <- struct{}{}
	}()

	return connector.SSHCommandContext{
		Command: "",
		Output:  outputChan,
		Errors:  errChan,
		Done:    doneChan,
	}
}

func TestParse(t *testing.T) {
	reference_pools := []local_pools.PoolGroup{
		{
			Name: "IP-JAIL",
			Pools: []local_pools.Pool{
				{
					StartIP:           "172.16.0.60",
					EndIP:             "172.16.0.119",
					AddressesTotal:    60,
					AddressesAvail:    56,
					AddressesAssigned: 4,
				},
				{
					StartIP:           "172.16.0.30",
					EndIP:             "172.16.0.55",
					AddressesTotal:    26,
					AddressesAvail:    22,
					AddressesAssigned: 4,
				},
			},
		},
		{
			Name: "IP-CGNAT-CISCO",
			Pools: []local_pools.Pool{

				{
					StartIP:           "100.65.128.0",
					EndIP:             "100.65.191.255",
					AddressesTotal:    16384,
					AddressesAvail:    6809,
					AddressesAssigned: 9575,
				},
			},
		},
		{
			Name: "IP-PUBLIC",
			Pools: []local_pools.Pool{

				{
					StartIP:           "5.159.24.0",
					EndIP:             "5.159.24.255",
					AddressesTotal:    256,
					AddressesAvail:    153,
					AddressesAssigned: 103,
				},
				{
					StartIP:           "5.159.25.0",
					EndIP:             "5.159.25.255",
					AddressesTotal:    256,
					AddressesAvail:    140,
					AddressesAssigned: 116,
				},
				{
					StartIP:           "5.159.26.0",
					EndIP:             "5.159.26.255",
					AddressesTotal:    256,
					AddressesAvail:    130,
					AddressesAssigned: 126,
				},
			},
		},
	}

	ctx := inputContext()
	poolGroupChan := make(chan *local_pools.PoolGroup)
	done := make(chan struct{})

	go local_pools.ParsePool(&ctx, poolGroupChan, done)

	at := 0
	for {
		select {
		case poolGroup := <-poolGroupChan:
			for i, pool := range poolGroup.Pools {
				if pool != reference_pools[at].Pools[i] {
					t.Errorf("Got an unexpected interface output, expected %v, got %v", reference_pools[at], poolGroup)
				}
			}

			at++
		case <-done:
			if at != len(reference_pools) {
				t.Errorf("Got %d interfaces, expected %d", at+1, len(reference_pools))
			}
			return
		}
	}
}
