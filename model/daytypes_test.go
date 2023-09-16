package model

import (
	"testing"
)

func TestDayType_String(t *testing.T) {
	tests := []struct {
		name string
		t    DayType
		want string
	}{
		{"WorkDay", WorkDay, "Working day"},
		{"NoWorkDay", NoWorkDay, "Non-working day"},
		{"ShortWorkDay", ShortWorkDay, "Short working day"},
		{"DayError", DayError, "Day error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
