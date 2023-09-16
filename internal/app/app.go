package app

import (
	"fmt"

	"github.com/duke0x/ts-notifier/config"
	"github.com/duke0x/ts-notifier/tscalculator"
)

type Notifier interface {
	Notify(channel, message string) error
}

type App struct {
	args        config.Args
	params      config.Params
	dtFetcher   tscalculator.DayTypeFetcher
	logsFetcher tscalculator.WorkLogFetcher
	notifier    Notifier
}

func NewCliApp(
	args config.Args,
	params config.Params,
	dtFetcher tscalculator.DayTypeFetcher,
	logsFetcher tscalculator.WorkLogFetcher,
	notifier Notifier,
) *App {
	return &App{
		args:        args,
		params:      params,
		dtFetcher:   dtFetcher,
		logsFetcher: logsFetcher,
		notifier:    notifier,
	}
}

func (app *App) Run() (err error) {
	tsc := tscalculator.New(app.dtFetcher, app.logsFetcher)
	for _, team := range app.params.Teams {
		teamSpends, err := tsc.CalcDailyTimeSpends(app.args.Date, team)
		if err != nil {
			return fmt.Errorf("checking time spends: %w", err)
		}

		if teamSpends.RemainSpend() == 0 {
			fmt.Printf(
				"all members of team '%s' has written their timelogs\n",
				team.Name,
			)
		}

		if err := app.notifier.Notify(
			team.Channel,
			teamSpends.Report(app.args.Date),
		); err != nil {
			return fmt.Errorf(
				"notify about remaining team '%s' time spends: %w",
				team.Name,
				err,
			)
		}
		fmt.Printf("notification for team '%s' sent\n", team.Name)
	}

	return nil
}
