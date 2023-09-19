package app

import (
	"context"
	"fmt"
	"github.com/duke0x/ts-notifier/client"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/duke0x/ts-notifier/config"
	mock_day_type_fetcher "github.com/duke0x/ts-notifier/mock/day_type_fetcher"
	mock_notifier "github.com/duke0x/ts-notifier/mock/notifier"
	mock_worklog_fetcher "github.com/duke0x/ts-notifier/mock/work_log_fetcher"
	"github.com/duke0x/ts-notifier/model"
	"github.com/duke0x/ts-notifier/tscalculator"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestApp_RunSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		day := time.Now().Truncate(60 * time.Second)

		app := NewCliApp(
			config.Args{
				ConfigPath: "config.yml",
				Date:       time.Now().Truncate(60 * time.Second),
			},
			config.Params{
				Jira:     config.Jira{},
				Notifier: config.Notifier{},
				Teams: []config.Team{{
					Name:    "team1",
					Channel: "channel-team1",
					Members: []config.Member{{
						Name:               "Ivan Ivanov",
						JiraAccID:          "18gdasid123123jas",
						MattermostUsername: "ivanov.i",
						Email:              "ivanov.i@my.org",
					}},
				}},
			},
			mock_day_type_fetcher.NewMockDayTypeFetcher(ctrl),
			mock_worklog_fetcher.NewMockWorkLogFetcher(ctrl),
			mock_notifier.NewMockNotifier(ctrl),
		)

		dtf, ok := app.dtFetcher.(*mock_day_type_fetcher.MockDayTypeFetcher)
		assert.Equal(t, true, ok)

		dtf.EXPECT().FetchDayType(ctx, day).Return(model.WorkDay, nil)

		wlf, ok := app.logsFetcher.(*mock_worklog_fetcher.MockWorkLogFetcher)
		assert.Equal(t, true, ok)

		for _, team := range app.params.Teams {
			var trs tscalculator.TeamRemainSpends
			for _, member := range team.Members {
				issues := []model.Issue{{ID: "asdwelqwkmcsl12edsa", Key: "PRJ-1"}}

				wlf.EXPECT().UserWorkedIssuesByDate(
					ctx,
					model.User(member.JiraAccID),
					day,
				).Return(issues, nil)

				dayStart := day.Truncate(time.Hour * 24).UTC()
				dayEnd := dayStart.Add(time.Hour*23 + time.Minute*59 + time.Second*59)
				wl := []model.WorkLog{{
					Key:              "PRJ-1",
					User:             model.User(member.JiraAccID),
					TimeSpentSeconds: 7 * 3600,
					Started:          dayStart,
					Comment:          "some work",
				}}

				wlf.EXPECT().WorkLogsPerIssues(ctx, model.User(member.JiraAccID), dayStart, dayEnd, issues).Return(wl, nil)
				trs = append(trs, tscalculator.MemberRemainSpend{
					Member:      member,
					RemainSpend: 1 * time.Hour,
				})
			}

			n, ok := app.notifier.(*mock_notifier.MockNotifier)
			require.Equal(t, true, ok)

			n.EXPECT().Notify(team.Channel, trs.Report(app.args.Date)).Return(nil)
		}

		err := app.Run()
		require.NoError(t, err)
	})
}

func TestApp_RunNotifierError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	day := time.Now().Truncate(60 * time.Second)

	app := NewCliApp(
		config.Args{
			ConfigPath: "config.yml",
			Date:       time.Now().Truncate(60 * time.Second),
		},
		config.Params{
			Jira:     config.Jira{},
			Notifier: config.Notifier{},
			Teams: []config.Team{{
				Name:    "team1",
				Channel: "channel-team1",
				Members: []config.Member{{
					Name:               "Ivan Ivanov",
					JiraAccID:          "18gdasid123123jas",
					MattermostUsername: "ivanov.i",
					Email:              "ivanov.i@my.org",
				}},
			}},
		},
		mock_day_type_fetcher.NewMockDayTypeFetcher(ctrl),
		mock_worklog_fetcher.NewMockWorkLogFetcher(ctrl),
		mock_notifier.NewMockNotifier(ctrl),
	)

	dtf, ok := app.dtFetcher.(*mock_day_type_fetcher.MockDayTypeFetcher)
	assert.Equal(t, true, ok)

	dtf.EXPECT().FetchDayType(ctx, day).Return(model.WorkDay, nil)

	wlf, ok := app.logsFetcher.(*mock_worklog_fetcher.MockWorkLogFetcher)
	assert.Equal(t, true, ok)

	for _, team := range app.params.Teams {
		var trs tscalculator.TeamRemainSpends
		for _, member := range team.Members {
			issues := []model.Issue{{ID: "asdwelqwkmcsl12edsa", Key: "PRJ-1"}}

			wlf.EXPECT().UserWorkedIssuesByDate(
				ctx,
				model.User(member.JiraAccID),
				day,
			).Return(issues, nil)

			dayStart := day.Truncate(time.Hour * 24).UTC()
			dayEnd := dayStart.Add(time.Hour*23 + time.Minute*59 + time.Second*59)
			wl := []model.WorkLog{{
				Key:              "PRJ-1",
				User:             model.User(member.JiraAccID),
				TimeSpentSeconds: 7 * 3600,
				Started:          dayStart,
				Comment:          "some work",
			}}

			wlf.EXPECT().WorkLogsPerIssues(ctx, model.User(member.JiraAccID), dayStart, dayEnd, issues).Return(wl, nil)
			trs = append(trs, tscalculator.MemberRemainSpend{
				Member:      member,
				RemainSpend: 1 * time.Hour,
			})
		}

		n, ok := app.notifier.(*mock_notifier.MockNotifier)
		require.Equal(t, true, ok)

		n.EXPECT().Notify(team.Channel, trs.Report(app.args.Date)).Return(fmt.Errorf("service unavailable"))
	}

	err := app.Run()
	require.EqualError(t, err, "notify about remaining team '"+
		app.params.Teams[0].Name+"' time spends: "+"service unavailable")
}

func TestApp_RunCalcDailyTimeSpendsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	day := time.Now().Truncate(60 * time.Second)

	app := NewCliApp(
		config.Args{
			ConfigPath: "config.yml",
			Date:       time.Now().Truncate(60 * time.Second),
		},
		config.Params{
			Jira:     config.Jira{},
			Notifier: config.Notifier{},
			Teams: []config.Team{{
				Name:    "team1",
				Channel: "channel-team1",
				Members: []config.Member{{
					Name:               "Ivan Ivanov",
					JiraAccID:          "18gdasid123123jas",
					MattermostUsername: "ivanov.i",
					Email:              "ivanov.i@my.org",
				}},
			}},
		},
		mock_day_type_fetcher.NewMockDayTypeFetcher(ctrl),
		mock_worklog_fetcher.NewMockWorkLogFetcher(ctrl),
		mock_notifier.NewMockNotifier(ctrl),
	)

	dtf, ok := app.dtFetcher.(*mock_day_type_fetcher.MockDayTypeFetcher)
	assert.Equal(t, true, ok)

	dtf.EXPECT().FetchDayType(ctx, day).Return(model.DayError, client.ErrServiceUnavailable)

	fetchDayTypeError := fmt.Sprintf(
		"checking time spends: checking day '%s': %s",
		app.args.Date.Format("20060102"),
		client.ErrServiceUnavailable.Error(),
	)
	err := app.Run()
	require.EqualError(t, err, fetchDayTypeError)
}
