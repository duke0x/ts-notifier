package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/duke0x/ts-notifier/client"
	"github.com/duke0x/ts-notifier/config"
	"github.com/duke0x/ts-notifier/internal/app"
	"github.com/duke0x/ts-notifier/internal/stdoutnotifier"
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
	args, err := config.ProcessArgs(os.Args[1:])
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
		exit(fmt.Sprintf("reading config file: %s", err.Error()), readConfig)
	}

	// initialize dependencies
	do := client.NewIsDayOff(&http.Client{}, client.IsDayOffURL)
	jira := client.NewJiraCli(&http.Client{}, cfg.Jira)
	var n app.Notifier
	n = client.NewNotifier(&http.Client{}, cfg.Mattermost)
	if cfg.Mattermost.URL == "" {
		n = &stdoutnotifier.StdOut{}
	}

	a := app.NewCliApp(args, cfg, do, jira, n)
	// a := app.NewCliApp(args, cfg, do, jira, tn)
	if err := a.Run(); err != nil {
		exit(
			fmt.Sprintf("check remaining time spends & notify: %s", err.Error()),
			checkTS,
		)
	}
}
