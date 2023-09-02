// Package isdayoff is used to define the type of day:
// working, non-working or shortened day.
// It uses http service isdayoff.ru.
// IsDayOff.ru REST API:
//   - description, https://www.isdayoff.ru/desc/
//   - extended, https://www.isdayoff.ru/extapi/
//
// Important: currently this client implementation works only with russian work calendar
package isdayoff

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/duke0x/ts-notifier/model"
)

var (
	ErrBadDayFormat = errors.New("invalid day format. example: YYYYMMDD or YYYY-MM-DD")
	ErrDataNotFound = errors.New("day data not found")
	ErrServiceError = errors.New("service not working")
	ErrUnknown      = errors.New("service return non-specified code")
)

// isdayoff service response codes
const (
	errBadDay      = 100
	errNoData      = 101
	errUnavailable = 199
)

const URL = "https://isdayoff.ru"

type Service struct{}

// CheckDay goes to isdayoff.ru site via https REST API with day parameter
// and returns this day type. Day types described in model.DayType.
// If service return error model.Error and error.
func (s Service) CheckDay(dt time.Time) (model.DayType, error) {
	day := dt.Format(model.DayFormat)
	url := strings.Join([]string{URL, "/", day}, "")
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.Error, fmt.Errorf("creating request to 'isdayof' service: %w", err)
	}

	qp := req.URL.Query()
	qp.Add("pre", "1") // check if day is short
	req.URL.RawQuery = qp.Encode()

	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return model.Error, fmt.Errorf("sending request to 'isdayof' service: %w", err)
	}
	if resp == nil {
		return model.Error, errors.New("response is nil")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var bb bytes.Buffer
	_, err = bb.ReadFrom(resp.Body)
	if err != nil {
		return model.Error, fmt.Errorf("reading 'isdayof' response: %w", err)
	}

	const (
		err4XX = 4
		err5XX = 5
	)

	rc, err := strconv.Atoi(bb.String())
	if err != nil {
		return model.Error, fmt.Errorf("service return non-integer code")
	}
	if resp.StatusCode/100 == err4XX {
		switch rc {
		case errBadDay:
			return model.Error, ErrBadDayFormat
		case errNoData:
			return model.Error, ErrDataNotFound
		case errUnavailable:
			return model.Error, ErrServiceError
		default:
			return model.Error, ErrUnknown
		}
	} else if resp.StatusCode/100 == err5XX {
		return model.Error, ErrServiceError
	}

	return model.DayType(rc), nil
}
