package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/duke0x/ts-notifier/config"
	"github.com/duke0x/ts-notifier/model"
)

// Jira fetched work-logs data from jira worklogs service
type Jira struct {
	config.Jira
}

func NewJira(jiraParams config.Jira) *Jira {
	return &Jira{jiraParams}
}

func (wls *Jira) WorkLogsPerIssues(
	user model.User,
	startedAfter time.Time,
	startedBefore time.Time,
	issues []model.Issue,
) ([]model.WorkLog, error) {
	var wl []model.WorkLog
	for _, issue := range issues {
		fullURL := fmt.Sprintf("%s/rest/api/2/issue/%s/worklog", wls.URL, issue.Key)
		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		qp := req.URL.Query()
		qp.Add("startedAfter", strconv.FormatInt(startedAfter.UnixMilli(), 10))
		qp.Add("startedBefore", strconv.FormatInt(startedBefore.UnixMilli(), 10))
		req.URL.RawQuery = qp.Encode()
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.SetBasicAuth(wls.UserEmail, wls.AuthToken)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("sending 'get worklogs' request: %w", err)
		}

		rspData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reqding jira_worklogs 'get worklogs' response: %w", err)
		}
		_ = resp.Body.Close()

		var wlResp workLogResponse
		err = json.Unmarshal(rspData, &wlResp)
		if err != nil {
			return nil, fmt.Errorf("decoding jira_worklogs 'get worklogs' response: %w", err)
		}

		for _, worklog := range wlResp.Worklogs {
			if worklog.Author.AccountID != string(user) {
				continue
			}

			st, err := time.Parse("2006-01-02T15:04:05.999-0700", worklog.Started)
			if err != nil {
				continue
			}

			wl = append(wl, model.WorkLog{
				Key:              issue.Key,
				User:             user,
				TimeSpentSeconds: worklog.TimeSpentSeconds,
				Started:          st,
				Comment:          worklog.Comment,
			})
		}
	}

	return wl, nil
}

func (wls *Jira) UserWorkedIssuesByDate(
	user model.User,
	date time.Time,
) ([]model.Issue, error) {
	url := strings.Join([]string{
		wls.URL,
		"/rest/api/2/search",
	}, "")

	// Create a new HTTP request
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.SetBasicAuth(wls.UserEmail, wls.AuthToken)
	req.Header.Set("Accept", "application/json")

	jql := fmt.Sprintf(
		"worklogDate=%s AND worklogAuthor=%s",
		date.Format("2006-01-02"),
		user,
	)

	// Set query params
	qp := req.URL.Query()
	qp.Add("jql", jql)
	qp.Add("fields", "summary")
	req.URL.RawQuery = qp.Encode()

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending 'get worklogs' request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	rspData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reqding jira_worklogs 'get worklogs' response: %w", err)
	}

	var wlResp issuesResponse
	err = json.Unmarshal(rspData, &wlResp)
	if err != nil {
		return nil, fmt.Errorf("decoding jira_worklogs 'get worklogs' response: %w", err)
	}

	issues := make([]model.Issue, 0, len(wlResp.Issues))
	for _, issue := range wlResp.Issues {
		issues = append(issues, model.Issue{
			ID:  issue.ID,
			Key: issue.Key,
		})
	}

	return issues, nil
}

// issuesResponse stores issues work logs data as described in
// https://developer.atlassian.om/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-search-get
type issuesResponse struct {
	Expand     string `json:"expand"`
	StartAt    int    `json:"startAt"`
	MaxResults int    `json:"maxResults"`
	Total      int    `json:"total"`
	Issues     []struct {
		Expand string `json:"expand"`
		ID     string `json:"id"`
		Self   string `json:"self"`
		Key    string `json:"key"`
		Fields struct {
			Summary string          `json:"summary"`
			Worklog workLogResponse `json:"worklog"`
		} `json:"fields"`
	} `json:"issues"`
}

type workLogResponse struct {
	StartAt    int `json:"startAt"`
	MaxResults int `json:"maxResults"`
	Total      int `json:"total"`
	Worklogs   []struct {
		Author struct {
			AccountID    string `json:"accountId"`
			EmailAddress string `json:"emailAddress,omitempty"`
			DisplayName  string `json:"displayName"`
		} `json:"author"`
		UpdateAuthor struct {
			AccountID    string `json:"accountId"`
			EmailAddress string `json:"emailAddress,omitempty"`
			DisplayName  string `json:"displayName"`
		} `json:"updateAuthor"`
		Comment          string `json:"comment"`
		Created          string `json:"created"`
		Updated          string `json:"updated"`
		Started          string `json:"started"`
		TimeSpentSeconds int    `json:"timeSpentSeconds"`
		IssueID          string `json:"issueId"`
	} `json:"worklogs"`
}
