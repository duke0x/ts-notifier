package client

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
	ErrNonIntegerCode     = errors.New("service return non-integer code")
	ErrBadDayFormat       = errors.New("invalid day format. example: YYYYMMDD or YYYY-MM-DD")
	ErrDataNotFound       = errors.New("day data not found")
	ErrServiceUnavailable = errors.New("service not working")
	ErrUnknown            = errors.New("service return non-specified code")
)

// IsDayOff.ru service response codes
const (
	errBadDay      = 100
	errNoData      = 101
	errUnavailable = 199
)

const IsDayOffURL = "https://isdayoff.ru"

type IsDayOff struct {
	client *http.Client
	url    string
}

func NewIsDayOff(client *http.Client, srvURL string) IsDayOff {
	url := IsDayOffURL
	if srvURL != "" {
		url = srvURL
	}

	return IsDayOff{client: client, url: url}
}

// FetchDayType returns the type of day: working, non-working or shortened day.
// It goes to IsDayOffURL site via https REST API with day parameter
// and returns this day type described in model.DayType.
// If IsDayOffURL returns error this function returns model.DayError and error.
func (i IsDayOff) FetchDayType(ctx context.Context, dt time.Time) (model.DayType, error) {
	day := dt.Format(model.DayFormat)
	url := strings.Join([]string{i.url, "/", day}, "")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.DayError, fmt.Errorf("creating request to 'isdayof' service: %w", err)
	}

	qp := req.URL.Query()
	qp.Add("pre", "1") // check if day is short
	req.URL.RawQuery = qp.Encode()

	resp, err := i.client.Do(req)
	if err != nil {
		return model.DayError, fmt.Errorf("sending request to 'isdayof' service: %w", err)
	}
	if resp == nil {
		return model.DayError, errors.New("response is nil")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var bb bytes.Buffer
	_, err = bb.ReadFrom(resp.Body)
	if err != nil {
		return model.DayError, fmt.Errorf("reading 'isdayof' response: %w", err)
	}

	rc, err := strconv.Atoi(bb.String())
	if err != nil {
		return model.DayError, ErrNonIntegerCode
	}
	if resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusNotFound {
		switch rc {
		case errBadDay:
			return model.DayError, ErrBadDayFormat
		case errNoData:
			return model.DayError, ErrDataNotFound
		case errUnavailable:
			return model.DayError, ErrServiceUnavailable
		default:
			return model.DayError, ErrUnknown
		}
	} else if resp.StatusCode >= http.StatusInternalServerError {
		return model.DayError, ErrServiceUnavailable
	}

	return model.DayType(rc), nil
}
