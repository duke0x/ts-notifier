package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var ErrBadDayFormat = errors.New("bad day format")

// Params stores Jira, Notifier (Mattermost) and Teams parameters
type Params struct {
	Jira     `yaml:"jira"`
	Notifier `yaml:"notifier"`
	Teams    `yaml:"teams"`
}

// Jira stores Jira URL and access credentials
type Jira struct {
	URL string `yaml:"url"`

	UserEmail string `yaml:"user_email"`

	// AuthToken is a Jira authentication token
	// https://confluence.atlassian.com/enterprise/using-personal-access-tokens-1026032365.html
	AuthToken string `yaml:"auth_token"`
}

type Team struct {
	// Name is a team name
	Name string `yaml:"name"`
	// Channel stores identifier in mattermost for this team
	Channel string `yaml:"channel"`
	// Members is a list of members in team
	Members []Member `yaml:"members"`
}

type Member struct {
	// Name is a member name
	Name string `yaml:"name"`
	// JiraAccID is a member Jira account identifier.
	// It is needed for fetching work logs.
	JiraAccID string `yaml:"jira_account_id"`
	// MattermostUsername is a member mattermost username
	// It is needed for tagging user in mattermost notification message
	MattermostUsername string `yaml:"mattermost_username"`
	// Email is a user email, can be omitted
	Email string `yaml:"email"`
}

// Teams stores list of teams and members of this teams
type Teams []Team

// Notifier stores notifiers (Mattermost) settings
type Notifier struct {
	Mattermost `yaml:"mattermost"`
}

// Mattermost stores Mattermost server URL and authentication token
type Mattermost struct {
	// URL is a Mattermost server URL
	URL string `yaml:"url"`
	// AuthToken ia a Mattermost authentication token
	// https://docs.mattermost.com/integrations/cloud-personal-access-tokens.html#:~:text=Sign%20in%20to%20the%20user,Select%20Save.
	AuthToken string `yaml:"auth_token"`
}

// Args command-line parameters
type Args struct {
	// ConfigPath is a path to config file 'config.yml'
	ConfigPath string
	// Date is a day for which time spends will be checked
	Date time.Time
}

// ProcessArgs processes command arguments and fills the Args structure
func ProcessArgs(args []string) (Args, error) {
	var a Args

	f := flag.NewFlagSet("time spends notifier", 1)
	f.StringVar(
		&a.ConfigPath,
		"c",
		"./config.yml",
		"Path to configuration file",
	)

	const dayFormat = "2006-01-02"
	var date string
	f.StringVar(
		&date,
		"d",
		time.Now().UTC().Format(dayFormat),
		"What day is to be reported, format: "+dayFormat+".",
	)

	if err := f.Parse(args); err != nil {
		_, _ = fmt.Fprintln(f.Output())
		return Args{}, err
	}

	t, err := time.Parse(dayFormat, date)
	if err != nil {
		return Args{}, ErrBadDayFormat
	}
	a.Date = t

	return a, nil
}

// ReadConfig reads config file and fills Params structure
func ReadConfig(path string) (Params, error) {
	f, err := os.Open(path)
	if err != nil {
		return Params{}, err
	}

	cfgBytes, err := io.ReadAll(f)
	if err != nil {
		return Params{}, fmt.Errorf("reading config file '%s': %w", path, err)
	}

	var params Params
	if err = yaml.Unmarshal(cfgBytes, &params); err != nil {
		return Params{}, fmt.Errorf("unmarshal config file data: %w", err)
	}

	return params, nil
}
