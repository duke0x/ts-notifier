// Package tscalculator fetches information about current day model.DayType.
// If day is non-working day it returns ErrNonWorkingDay.
// Next it fetches all working logs for each team member in all teams.
// Sends message to mattermost team-channel for each team
// if almost one member has not spent all working hours per day.
package tscalculator

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/duke0x/ts-notifier/config"
	"github.com/duke0x/ts-notifier/model"
)

// проверяем тип дня, если не рабочий, скипаем
// для каждой команды
// достаем пользователей из команды
// получаем задачи пользователя за день
// получаем списания по задачам за день
// проверяем, все ли списано за день, если не все добавляем в сообщение
// отправляем сообщение в группу команды, тегаем тех кто не списался
// завершаем работу

var ErrNonWorkingDay = errors.New("non working day")

type TSCalc struct {
	dc  DayChecker
	wlf WorkLogFetcher
	n   Notifier
}

func New(dc DayChecker, wlf WorkLogFetcher, n Notifier) *TSCalc {
	return &TSCalc{
		dc:  dc,
		wlf: wlf,
		n:   n,
	}
}

type DayChecker interface {
	CheckDay(dt time.Time) (model.DayType, error)
}

type WorkLogFetcher interface {
	UserWorkedIssuesByDate(user model.User, date time.Time) ([]model.Issue, error)
	WorkLogsPerIssues(
		user model.User,
		startedAfter time.Time,
		startedBefore time.Time,
		issues []model.Issue,
	) ([]model.WorkLog, error)
}

type Notifier interface {
	Notify(channel, message string) error
}

func (tsc TSCalc) CheckDailyTimeSpends(day time.Time, teams config.Teams) error {
	ds := day.Format(model.DayFormat)
	dt, err := tsc.dc.CheckDay(day)
	if err != nil {
		return fmt.Errorf("checking day '%s': %w", ds, err)
	}

	if dt == model.NoWorkDay {
		return fmt.Errorf("%w; day: %s", ErrNonWorkingDay, ds)
	}

	for _, team := range teams {
		var teamMessage string

		for _, member := range team.Members {
			user := model.User(member.JiraAccID)
			issues, err := tsc.wlf.UserWorkedIssuesByDate(user, day)
			if err != nil {
				return fmt.Errorf("fetching user worked issies: %w", err)
			}

			dayBefore := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, time.Local)
			wl, err := tsc.wlf.WorkLogsPerIssues(user, day, dayBefore, issues)
			if err != nil {
				return fmt.Errorf("fetching working issues: %w", err)
			}

			tsWorked := aggregateTimeSpent(wl, user, day)

			tsRemain := remainTimeSpend(tsWorked, dt)

			if tsRemain.Seconds() != 0 {
				if teamMessage == "" {
					teamMessage = "Отчет по списанию времени за " + day.Format("2006.01.02") + ":\n"
				}

				teamMessage += "  - @" + member.MattermostUsername + " нужно списать еще " + tsRemain.String() + ".\n"
			}
		}

		if teamMessage != "" {
			if err = tsc.n.Notify(team.Channel, teamMessage); err != nil {
				return fmt.Errorf("notify about remaining time spends: %w", err)
			}
			fmt.Printf("notification for team '%s' sent\n", team.Name)
		} else {
			fmt.Println("all members of team", team.Name, "has written their timelogs correctly")
		}
	}

	return nil
}

func aggregateTimeSpent(
	wls []model.WorkLog,
	user model.User,
	day time.Time,
) time.Duration {
	totalSecondsSpent := 0
	for _, wl := range wls {
		if wl.User != user {
			continue
		}

		wly, wlm, wld := wl.Started.Date()
		dayy, daym, dayd := day.Date()
		if wly != dayy || wlm != daym || wld != dayd {
			continue
		}

		totalSecondsSpent += wl.TimeSpentSeconds
	}

	ts, _ := time.ParseDuration(strconv.Itoa(totalSecondsSpent) + "s")

	return ts
}

func remainTimeSpend(ts time.Duration, dayType model.DayType) time.Duration {
	workDayTime := 28800 // 8h *3600s - regular work daytime

	if dayType == model.ShortWorkDay {
		workDayTime -= 3600 // remove 1h if day is short day
	}

	diffSec := 0
	if int(ts.Seconds()) < workDayTime {
		diffSec = workDayTime - int(ts.Seconds())
	}

	diffTS, _ := time.ParseDuration(strconv.Itoa(diffSec) + "s")

	return diffTS
}
