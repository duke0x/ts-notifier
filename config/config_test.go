package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestReadConfig(t *testing.T) {
	want := Params{
		Jira: Jira{
			URL:       "https://myorg.atlassian.net",
			UserEmail: "user@emample.com",
			AuthToken: "<user-token>",
		},
		Notifier: Notifier{Mattermost{
			URL:       "https://chat.myorg.com",
			AuthToken: "<service-user-token>",
		}},
		Teams: Teams{Team{
			Name:    "my-jira-team-name",
			Channel: "<my-mattermost-team-channel-ID>",
			Members: []Member{{
				Name:               "<my team member 1>",
				JiraAccID:          "<team member 1 jira account ID>",
				MattermostUsername: "<team member 1 mattermost name>",
				Email:              "member1@myorg.com",
			}, {
				Name:               "<my team member 2>",
				JiraAccID:          "<team member 2 jira account ID>",
				MattermostUsername: "<team member 2 mattermost name>",
				Email:              "member2@myorg.com",
			}},
		}},
	}

	path := "config-example.yml"
	got, err := ReadConfig(path)
	if err != nil {
		t.Errorf("ReadConfig() error = %v, wantErr: nil", err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadConfig() got = %v, want %v", got, want)
	}

}

func TestReadConfigFileNotExist(t *testing.T) {
	path := "config-file-not-exists.yml"
	params, err := ReadConfig(path)
	require.Equal(t, true, os.IsNotExist(err))
	require.Equal(t, Jira{}, params.Jira)
	require.Equal(t, Mattermost{}, params.Mattermost)
}

func TestProcessArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Args
		wantErr bool
	}{
		{
			name: "no args",
			args: []string{},
			want: Args{
				ConfigPath: "./config.yml",
				Date:       time.Now().UTC().Truncate(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "set custom date",
			args: []string{"-d=2023-09-09"},
			want: Args{
				ConfigPath: "./config.yml",
				Date:       time.Date(2023, 9, 9, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "set custom config",
			args: []string{"-c=custom-config.yml"},
			want: Args{
				ConfigPath: "custom-config.yml",
				Date:       time.Now().Truncate(24 * time.Hour).UTC(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessArgs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
