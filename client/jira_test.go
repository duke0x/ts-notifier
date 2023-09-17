package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duke0x/ts-notifier/config"
	"github.com/duke0x/ts-notifier/model"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestUserWorkedIssuesByDate(t *testing.T) {
	var (
		user model.User = "user1"
		date            = time.Now()
	)
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr error
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/rest/api/2/search", r.URL.Path)
				require.Equal(t, http.MethodGet, r.Method)
				rp := r.URL.Query()
				require.Equal(t, "summary", rp.Get("fields"))
				jql := rp.Get("jql")
				require.NotEmpty(t, jql)
				_, _ = url.ParseQuery(jql)

				w.WriteHeader(http.StatusCreated)
				resp := issuesResponse{}
				respBytes, _ := json.Marshal(resp)
				_, _ = w.Write(respBytes)
			},

			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			jc := NewJiraCli(srv.Client(), config.Jira{
				URL: srv.URL,
			})

			_, err := jc.UserWorkedIssuesByDate(context.Background(), user, date)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr.Error())
			}
		})
	}
}

func TestWorkLogsPerIssues(t *testing.T) {
	var (
		ctx                      = context.Background()
		user          model.User = "user1"
		date                     = time.Now()
		startedAfter             = date.Truncate(24 * time.Hour)
		startedBefore            = startedAfter.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		issues                   = []model.Issue{{ID: "a1asd1a1asd1", Key: "PRJ-1"}}
	)
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr error
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, fmt.Sprintf(
					"/rest/api/2/issue/%s/worklog",
					issues[0].Key,
				), r.URL.Path)
				require.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(http.StatusCreated)
				resp := workLogResponse{}
				respBytes, _ := json.Marshal(resp)
				_, _ = w.Write(respBytes)
			},

			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			jc := NewJiraCli(srv.Client(), config.Jira{
				URL: srv.URL,
			})

			_, err := jc.WorkLogsPerIssues(ctx, user, startedAfter, startedBefore, issues)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr.Error())
			}
		})
	}
}
