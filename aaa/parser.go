package aaa

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

type context struct {
	subsystem    string
	radiusServer *RadiusServer
}

type handleMatch func(*context, []string)

type matcher struct {
	regexp       *regexp.Regexp
	handleMatch  handleMatch
	needscontext bool
}

// Parse parses cli output and tries to find interfaces with related stats
func (c *Collector) parse(sshCtx *connector.SSHCommandContext, radiusServers chan *RadiusServer, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	newServerRegexp := regexp.MustCompile(`RADIUS: id (\d+), priority (\d+), host ([^,]*), auth-port (\d+), acct-port (\d+)`)

	matchers := makematchers()
	context := &context{}

	for {
		select {
		case <-sshCtx.Done:
			if context.radiusServer != nil {
				radiusServers <- context.radiusServer
			}
			return
		case line := <-sshCtx.Output:
			if matches := newServerRegexp.FindStringSubmatch(line); matches != nil {
				if context.radiusServer != nil {
					radiusServers <- context.radiusServer
				}
				context.radiusServer = NewRadiusServer()
				context.radiusServer.ID = matches[1]
				context.radiusServer.Priority = matches[2]
				context.radiusServer.Host = matches[3]
				context.radiusServer.AuthPort = matches[4]
				context.radiusServer.AccountingPort = matches[5]
			}
			if context.radiusServer == nil {
				continue
			}

			for _, matcher := range matchers {
				if matcher.needscontext && context.subsystem == "" {
					continue
				}
				matches := matcher.regexp.FindStringSubmatch(line)
				if matches != nil {
					matcher.handleMatch(context, matches)
				}
			}
		}
	}
}

func makematchers() []*matcher {
	return []*matcher{
		&matcher{
			regexp:      regexp.MustCompile(`State: current ([^\,]*), duration (\d+)`),
			handleMatch: setUptime,
		}, &matcher{
			regexp:      regexp.MustCompile(`Dead: total time (\d+)s, count (\d+)`),
			handleMatch: setDead,
		}, &matcher{
			regexp:      regexp.MustCompile(`Quarantined: (.*)$`),
			handleMatch: setQuarantined,
		}, &matcher{
			regexp:      regexp.MustCompile(`(Authen|Author|Account): request (\d+), timeouts (\d+), failover (\d+), retransmission (\d+)`),
			handleMatch: setRequests,
		}, &matcher{
			regexp:       regexp.MustCompile(`Response: accept (\d+), reject (\d+), challenge (\d+)`),
			handleMatch:  handleResponses1,
			needscontext: true,
		}, &matcher{
			regexp:       regexp.MustCompile(`Response: unexpected (\d+), server error (\d+), incorrect (\d+), time (\d+)ms`),
			handleMatch:  handleResponses2,
			needscontext: true,
		}, &matcher{
			regexp:       regexp.MustCompile(`Transaction: success (\d+), failure (\d+)`),
			handleMatch:  setTransmissions,
			needscontext: true,
		}, &matcher{
			regexp:       regexp.MustCompile(`Throttled: transaction (\d+), timeout (\d+), failure (\d+)`),
			handleMatch:  setThrottles,
			needscontext: true,
		}, &matcher{
			regexp: regexp.MustCompile(`Malformed responses: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.MalformedResponses[context.subsystem] = util.Str2float64(matches[1])
			},
			needscontext: true,
		}, &matcher{
			regexp: regexp.MustCompile(`Bad authenticators: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.BadAuthenticators[context.subsystem] = util.Str2float64(matches[1])
			},
			needscontext: true,
		}, &matcher{
			regexp: regexp.MustCompile(`Estimated Outstanding Access Transactions: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.EstimatedOutstandingAccessTransactions = util.Str2float64(matches[1])
			},
		}, &matcher{
			regexp: regexp.MustCompile(`Estimated Outstanding Accounting Transactions: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.EstimatedOutstandingAccountingTransactions = util.Str2float64(matches[1])
			},
		}, &matcher{
			regexp: regexp.MustCompile(`Estimated Throttled Access Transactions: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.EstimatedThrottledAccessTransactions = util.Str2float64(matches[1])
			},
		}, &matcher{
			regexp: regexp.MustCompile(`Estimated Throttled Accounting Transactions: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.EstimatedThrottledAccountingTransactions = util.Str2float64(matches[1])
			},
		}, &matcher{
			regexp: regexp.MustCompile(`high.*ago: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.RequestsPerMinuteHigh = util.Str2float64(matches[1])
			},
		}, &matcher{
			regexp: regexp.MustCompile(`low.*ago: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.RequestsPerMinuteLow = util.Str2float64(matches[1])
			},
		}, &matcher{
			regexp: regexp.MustCompile(`average: (\d+)`),
			handleMatch: func(context *context, matches []string) {
				context.radiusServer.RequestsPerMinuteAverage = util.Str2float64(matches[1])
			},
		},
	}
}

func setUptime(context *context, matches []string) {
	context.radiusServer.Up = 0
	if matches[1] == "UP" {
		context.radiusServer.Up = 1
	}
	context.radiusServer.UpDuration = util.Str2float64(matches[2])
}

func setDead(context *context, matches []string) {
	context.radiusServer.DeadTotalTime = util.Str2float64(matches[1])
	context.radiusServer.DeadCount = util.Str2float64(matches[2])
}

func setQuarantined(context *context, matches []string) {
	context.radiusServer.Quarantined = 1
	if matches[1] == "No" {
		context.radiusServer.Quarantined = 0
	}
}

func setRequests(context *context, matches []string) {
	context.subsystem = matches[1]
	context.radiusServer.Requests[context.subsystem] = util.Str2float64(matches[2])
	context.radiusServer.Timeouts[context.subsystem] = util.Str2float64(matches[3])
	context.radiusServer.Failovers[context.subsystem] = util.Str2float64(matches[4])
	context.radiusServer.Retransmissions[context.subsystem] = util.Str2float64(matches[5])
}

func handleResponses1(context *context, matches []string) {
	if context.radiusServer.Responses[context.subsystem] == nil {
		context.radiusServer.Responses[context.subsystem] = make(map[string]float64)
	}
	context.radiusServer.Responses[context.subsystem]["accept"] = util.Str2float64(matches[1])
	context.radiusServer.Responses[context.subsystem]["reject"] = util.Str2float64(matches[2])
	context.radiusServer.Responses[context.subsystem]["challenge"] = util.Str2float64(matches[3])
}

func handleResponses2(context *context, matches []string) {
	if context.radiusServer.Responses[context.subsystem] == nil {
		context.radiusServer.Responses[context.subsystem] = make(map[string]float64)
	}
	context.radiusServer.Responses[context.subsystem]["unexpected"] = util.Str2float64(matches[1])
	context.radiusServer.Responses[context.subsystem]["server error"] = util.Str2float64(matches[2])
	context.radiusServer.Responses[context.subsystem]["incorrect"] = util.Str2float64(matches[3])
	context.radiusServer.ResponseTime[context.subsystem] = util.Str2float64(matches[4])
}

func setTransmissions(context *context, matches []string) {
	context.radiusServer.SuccessfullTransactions[context.subsystem] = util.Str2float64(matches[1])
	context.radiusServer.FailedTransactions[context.subsystem] = util.Str2float64(matches[2])
}

func setThrottles(context *context, matches []string) {
	context.radiusServer.ThrottledTransactions[context.subsystem] = util.Str2float64(matches[1])
	context.radiusServer.ThrottledTimeouts[context.subsystem] = util.Str2float64(matches[2])
	context.radiusServer.ThrottledFailures[context.subsystem] = util.Str2float64(matches[3])
}
