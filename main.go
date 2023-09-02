package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/duke0x/ts-notifier/config"
	"github.com/duke0x/ts-notifier/isdayoff"
	"github.com/duke0x/ts-notifier/jiraworklogs"
	"github.com/duke0x/ts-notifier/notifier/mattermost"
	"github.com/duke0x/ts-notifier/tscalculator"
)

type errCode int

const (
	parseArgs  errCode = 1
	readConfig errCode = 2
	checkTS    errCode = 3
)

func exit(message string, code errCode) {
	fmt.Println(message)
	os.Exit(int(code))
}

func main() {
	args, err := config.ProcessArgs()
	if err != nil {
		if errors.Is(err, config.ErrBadDayFormat) {
			exit(fmt.Sprintf(
				"parsing config: %s, try 'YYYY-MM-DD",
				err.Error(),
			), parseArgs)
		}
		exit("parsing config:"+err.Error(), parseArgs)
	}

	// read configuration from the file
	cfg, err := config.ReadConfig(args.ConfigPath)
	if err != nil {
		exit(fmt.Sprintf(
			"reading config file: %s",
			err.Error(),
		), readConfig)
	}

	// TODO: validate config data

	// initialize dependencies
	dc := isdayoff.Service{}
	ws := jiraworklogs.NewWorkLogService(cfg.Jira)
	mmn := mattermost.NewNotifier(cfg.Mattermost)

	tsc := tscalculator.New(dc, ws, mmn)
	err = tsc.CheckDailyTimeSpends(args.Date, cfg.Teams)
	if err != nil {
		exit("checking time spends: "+err.Error(), checkTS)
	}
}
