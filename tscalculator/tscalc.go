// Package tscalculator fetches information about current day model.DayType.
// If day is non-working day it returns ErrNonWorkingDay.
package tscalculator

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

const (
	hoursPerDay        = 24
	hoursPerWorkingDay = 8
)

type TSCalc struct {
	dc  DayTypeFetcher
	wlf WorkLogFetcher
}

func New(dc DayTypeFetcher, wlf WorkLogFetcher) *TSCalc {
	return &TSCalc{
		dc:  dc,
		wlf: wlf,
	}
}

//go:generate mockgen -source=tscalc.go -destination=../mock/mock_tscalc.go
type DayTypeFetcher interface {
	FetchDayType(ctx context.Context, dt time.Time) (model.DayType, error)
}

type WorkLogFetcher interface {
	UserWorkedIssuesByDate(
		ctx context.Context,
		user model.User,
		date time.Time,
	) ([]model.Issue, error)
	WorkLogsPerIssues(
		ctx context.Context,
		user model.User,
		startedAfter time.Time,
		startedBefore time.Time,
		issues []model.Issue,
	) ([]model.WorkLog, error)
}

// MemberRemainSpend stores team member name and his remain time spend
type MemberRemainSpend struct {
	Member      config.Member
	RemainSpend time.Duration
}

// TeamRemainSpends stores all team member time remain spends
type TeamRemainSpends []MemberRemainSpend

// RemainSpend returns remained time to spend for all members in team
func (trs TeamRemainSpends) RemainSpend() time.Duration {
	total := time.Duration(0)
	for _, ms := range trs {
		total += ms.RemainSpend
	}

	return total
}

func (trs TeamRemainSpends) Report(day time.Time) string {
	// TODO: try to replace message building with template text/template
	var report strings.Builder
	report.WriteString("Отчет по списанию времени за " + day.Format("2006.01.02") + ":\n")

	emptyReport := true
	for _, urs := range trs {
		if urs.RemainSpend > 0 {
			emptyReport = false
			report.WriteString("  - @" + urs.Member.MattermostUsername +
				" нужно списать еще " + urs.RemainSpend.String() + ".\n")
		}
	}

	if emptyReport {
		report.WriteString("Все молодцы, все списания произведены! :)")
	}

	return report.String()
}

// CalcDailyTimeSpends returns remaining time spends for a team per day.
// It determines the model.DayType of the day and fetches all team members work logs.
// Then it calculates remaining time spent depends on model.DayType.
func (tsc TSCalc) CalcDailyTimeSpends(
	day time.Time,
	team config.Team,
) (TeamRemainSpends, error) {
	ds := day.Format(model.DayFormat)
	dt, err := tsc.dc.FetchDayType(context.Background(), day)
	if err != nil {
		return nil, fmt.Errorf("checking day '%s': %w", ds, err)
	}

	if dt == model.NoWorkDay {
		return nil, fmt.Errorf("%w; day: %s", ErrNonWorkingDay, ds)
	}

	dayStart := day.Truncate(time.Hour * hoursPerDay).UTC()
	dayEnd := dayStart.Add(time.Hour*23 + time.Minute*59 + time.Second*59)

	ctx := context.Background()
	trs := TeamRemainSpends{}
	for _, member := range team.Members {
		user := model.User(member.JiraAccID)
		issues, err := tsc.wlf.UserWorkedIssuesByDate(ctx, user, day)
		if err != nil {
			return nil, fmt.Errorf("fetching user worked issies: %w", err)
		}

		wl, err := tsc.wlf.WorkLogsPerIssues(ctx, user, dayStart, dayEnd, issues)
		if err != nil {
			return nil, fmt.Errorf("fetching working issues: %w", err)
		}

		tsWorked := calculateTimeSpent(user, wl, day)
		tsRemain := remainTimeSpend(tsWorked, dt)

		trs = append(trs, MemberRemainSpend{
			Member:      member,
			RemainSpend: tsRemain,
		})
	}

	return trs, nil
}

// calculateTimeSpent returns total amount of all user work logs per day
func calculateTimeSpent(
	user model.User,
	wls []model.WorkLog,
	day time.Time,
) (totalSpent time.Duration) {
	day = day.Truncate(time.Hour * hoursPerDay)

	for _, wl := range wls {
		if wl.User != user {
			continue
		}

		if !wl.Date().Equal(day) {
			continue
		}

		totalSpent += time.Duration(wl.TimeSpentSeconds) * time.Second
	}

	return
}

func remainTimeSpend(ts time.Duration, dayType model.DayType) (diff time.Duration) {
	workDayTime := hoursPerWorkingDay * time.Hour // regular work daytime

	if dayType == model.ShortWorkDay {
		workDayTime -= time.Hour // remove 1h if day is short day
	}

	if ts < workDayTime {
		diff = workDayTime - ts
	}

	return diff
}
