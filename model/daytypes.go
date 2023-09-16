package model

const DayFormat = "20060102"

type DayType int

const (
	DayError DayType = -1
	// WorkDay is a regular working day.
	// For example, in Russia it lasts 8 hours for 40h/week.
	WorkDay DayType = 0
	// NoWorkDay is a holiday or weekend day.
	// For example, in Russia it could be a Saturday or Sunday.
	NoWorkDay DayType = 1
	// ShortWorkDay is a day before the holidays.
	// For example, in Russia this day is shorter than regular on 1 hour.
	ShortWorkDay DayType = 2
)

// String converts DayType to string
func (t DayType) String() string {
	switch t {
	case WorkDay:
		return "Working day"
	case NoWorkDay:
		return "Non-working day"
	case ShortWorkDay:
		return "Short working day"
	case DayError:
		fallthrough
	default:
		return "Day error"
	}
}
