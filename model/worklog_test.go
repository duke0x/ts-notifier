package model

import (
	"reflect"
	"testing"
	"time"
)

func TestWorkLog_Date(t *testing.T) {
	type fields struct {
		Key              string
		User             User
		TimeSpentSeconds int
		Started          time.Time
		Comment          string
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Time
	}{
		{
			name: "regular day",
			fields: fields{
				Key:              "PRJ-1",
				User:             "user1",
				TimeSpentSeconds: 3600,
				Started:          time.Date(2023, 9, 01, 12, 35, 23, 0, time.UTC),
				Comment:          "some work",
			},
			want: time.Date(2023, 9, 01, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wl := WorkLog{
				Key:              tt.fields.Key,
				User:             tt.fields.User,
				TimeSpentSeconds: tt.fields.TimeSpentSeconds,
				Started:          tt.fields.Started,
				Comment:          tt.fields.Comment,
			}
			if got := wl.Date(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Date() = %v, want %v", got, tt.want)
			}
		})
	}
}
