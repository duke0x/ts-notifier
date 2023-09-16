package tscalculator

import (
	"context"
	"errors"
	"github.com/duke0x/ts-notifier/client"
	"github.com/duke0x/ts-notifier/config"
	mock_tscalculator "github.com/duke0x/ts-notifier/mock"
	"github.com/duke0x/ts-notifier/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func Test_calculateTimeSpent(t *testing.T) {
	type args struct {
		user model.User
		wls  []model.WorkLog
		day  time.Time
	}
	tests := []struct {
		name           string
		args           args
		wantTotalSpent time.Duration
	}{
		{"regular day with one record", args{
			user: "user1",
			wls: []model.WorkLog{
				{Key: "PRJ-1", User: "user1", TimeSpentSeconds: 8 * 3600, Started: time.Now(), Comment: "some work"},
			},
			day: time.Now(),
		}, 8 * time.Hour},
		{"regular day with two records", args{
			user: "user1",
			wls: []model.WorkLog{
				{Key: "PRJ-1", User: "user1", TimeSpentSeconds: 3600, Started: time.Now(), Comment: "some work"},
				{Key: "PRJ-2", User: "user1", TimeSpentSeconds: 1800, Started: time.Now(), Comment: "daily"},
			},
			day: time.Now(),
		}, 1*time.Hour + 30*time.Minute},
		{"no working logs", args{
			user: "user1",
			wls:  []model.WorkLog{},
			day:  time.Now(),
		}, 0},
		{"regular day with two records for 2 users", args{
			user: "user1",
			wls: []model.WorkLog{
				{Key: "PRJ-1", User: "user1", TimeSpentSeconds: 3600, Started: time.Now(), Comment: "some work"},
				{Key: "PRJ-2", User: "user2", TimeSpentSeconds: 1800, Started: time.Now(), Comment: "daily"},
			},
			day: time.Now(),
		}, 1 * time.Hour},
		{"regular day with two records in 2 days", args{
			user: "user1",
			wls: []model.WorkLog{
				{Key: "PRJ-1", User: "user1", TimeSpentSeconds: 3600, Started: time.Now(), Comment: "some work"},
				{Key: "PRJ-2", User: "user1", TimeSpentSeconds: 1800, Started: time.Now().Add(-24 * time.Hour), Comment: "daily"},
			},
			day: time.Now(),
		}, 1 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTotalSpent := calculateTimeSpent(tt.args.user, tt.args.wls, tt.args.day); gotTotalSpent != tt.wantTotalSpent {
				t.Errorf("calculateTimeSpent() = %v, want %v", gotTotalSpent, tt.wantTotalSpent)
			}
		})
	}
}

func Test_remainTimeSpend(t *testing.T) {
	type args struct {
		ts      time.Duration
		dayType model.DayType
	}
	tests := []struct {
		name     string
		args     args
		wantDiff time.Duration
	}{
		{
			name: "spends complete",
			args: args{
				ts:      8 * time.Hour,
				dayType: model.WorkDay,
			},
			wantDiff: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDiff := remainTimeSpend(tt.args.ts, tt.args.dayType); gotDiff != tt.wantDiff {
				t.Errorf("remainTimeSpend() = %v, want %v", gotDiff, tt.wantDiff)
			}
		})
	}
}

func TestTSCalc_CalcDailyTimeSpends(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	dc := mock_tscalculator.NewMockDayTypeFetcher(ctrl)
	wlf := mock_tscalculator.NewMockWorkLogFetcher(ctrl)

	// input params
	var (
		day  = time.Now()
		team = config.Team{
			Name:    "team1",
			Channel: "chan1",
			Members: []config.Member{{
				Name:               "user1",
				JiraAccID:          "user1_Jira_ID",
				MattermostUsername: "user1_MM_ID",
				Email:              "user1@example.com",
			}},
		}
	)

	// test specs
	var dayType = model.WorkDay

	dc.EXPECT().FetchDayType(ctx, day).Return(dayType, nil)

	for _, member := range team.Members {
		issues := []model.Issue{{ID: "asdni12312h31jg1h23", Key: "PRJ-1"}}
		wlf.EXPECT().UserWorkedIssuesByDate(
			ctx,
			model.User(member.JiraAccID),
			day,
		).Return(issues, nil)

		dayStart := day.Truncate(time.Hour * hoursPerDay).UTC()
		dayEnd := dayStart.Add(time.Hour*23 + time.Minute*59 + time.Second*59)

		wls := []model.WorkLog{{
			Key:              "PRJ-1",
			User:             model.User(member.JiraAccID),
			TimeSpentSeconds: 7200,
			Started:          time.Now(),
			Comment:          "some work",
		}}
		wlf.EXPECT().WorkLogsPerIssues(
			ctx,
			model.User(member.JiraAccID),
			dayStart,
			dayEnd,
			issues,
		).Return(wls, nil)
	}

	tsc := TSCalc{
		dc:  dc,
		wlf: wlf,
	}

	got, err := tsc.CalcDailyTimeSpends(day, team)
	if err != nil {
		t.Errorf("CalcDailyTimeSpends() got error = %v, but error should be nil", err)
		return
	}

	var want TeamRemainSpends
	want = []MemberRemainSpend{{
		Member: config.Member{
			Name:               "user1",
			JiraAccID:          "user1_Jira_ID",
			MattermostUsername: "user1_MM_ID",
			Email:              "user1@example.com",
		},
		RemainSpend: 8*time.Hour - 2*time.Hour,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CalcDailyTimeSpends() got = %+v, want %+v", got, want)
	}
}

func TestTSCalc_CalcDailyTimeSpendsErrServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	dc := mock_tscalculator.NewMockDayTypeFetcher(ctrl)
	wlf := mock_tscalculator.NewMockWorkLogFetcher(ctrl)

	// input params
	var (
		day  = time.Now()
		team = config.Team{
			Name:    "team1",
			Channel: "chan1",
			Members: []config.Member{{
				Name:               "user1",
				JiraAccID:          "user1_Jira_ID",
				MattermostUsername: "user1_MM_ID",
				Email:              "user1@example.com",
			}},
		}
	)

	dc.EXPECT().FetchDayType(ctx, day).Return(model.DayError, client.ErrServiceUnavailable)

	tsc := TSCalc{
		dc:  dc,
		wlf: wlf,
	}

	wantErr := client.ErrServiceUnavailable
	dayType, err := tsc.CalcDailyTimeSpends(day, team)
	require.Equal(t, true, errors.Is(err, wantErr))
	var trs TeamRemainSpends
	require.Equal(t, trs, dayType)
}

func TestTeamRemainSpends_RemainSpend(t *testing.T) {
	tests := []struct {
		name string
		trs  TeamRemainSpends
		want time.Duration
	}{
		{
			name: "one user, one work record",
			trs: []MemberRemainSpend{{
				Member: config.Member{
					Name:               "user1",
					JiraAccID:          "user1_Jira_ID",
					MattermostUsername: "user1_MM_ID",
					Email:              "user1@example.com",
				},
				RemainSpend: 1 * time.Hour,
			}},
			want: 1 * time.Hour,
		},
		{
			name: "two users",
			trs: []MemberRemainSpend{{
				Member: config.Member{
					Name:               "user1",
					JiraAccID:          "user1_Jira_ID",
					MattermostUsername: "user1_MM_ID",
					Email:              "user1@example.com",
				},
				RemainSpend: 1 * time.Hour,
			}, {
				Member: config.Member{
					Name:               "user1",
					JiraAccID:          "user1_Jira_ID",
					MattermostUsername: "user1_MM_ID",
					Email:              "user1@example.com",
				},
				RemainSpend: 2 * time.Hour,
			}},
			want: 3 * time.Hour,
		},
		{
			name: "no users",
			trs:  []MemberRemainSpend{},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.trs.RemainSpend(); got != tt.want {
				t.Errorf("RemainSpend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTeamRemainSpends_Report(t *testing.T) {
	type args struct {
		day time.Time
	}
	tests := []struct {
		name string
		trs  TeamRemainSpends
		args args
		want string
	}{
		{
			name: "no spends",
			trs:  TeamRemainSpends{},
			args: args{day: time.Now().Truncate(24 * time.Hour).UTC()},
			want: "Отчет по списанию времени за " + time.Now().Format("2006.01.02") + ":\nВсе молодцы, все списания произведены! :)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.trs.Report(tt.args.day); got != tt.want {
				t.Errorf("Report() = %v, want %v", got, tt.want)
			}
		})
	}
}
