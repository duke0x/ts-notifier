package model

import "time"

// Issue stores issue identifiers in different formats.
type Issue struct {
	// ID is a Jira internal issue identifier.
	ID string `json:"id"`
	// Key is a Jira issue identifier, written in user-friendly format.
	// Format: <project>-<task number>. Example: PRJ-1.
	Key string `json:"key"`
}

// WorkLog stores work log specific data.
type WorkLog struct {
	// Key is a Jira user issue identifier.
	Key string `json:"issue_key"`
	// User is a Jira user account identifier.
	User User `json:"user"`
	// TimeSpentSeconds time spend is seconds.
	TimeSpentSeconds int `json:"time_spent_seconds"`
	// Started is a time of record.
	Started time.Time `json:"started"`
	// Comment is a short message 'what was done'.
	Comment string `json:"comment"`
}

// Date returns work log date only
func (wl WorkLog) Date() time.Time {
	const hoursPerDay = 24

	return wl.Started.Truncate(time.Hour * hoursPerDay)
}
