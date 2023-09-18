package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/duke0x/ts-notifier/model"
	"github.com/stretchr/testify/require"
)

func TestIsDayOff_FetchDayType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		day := time.Now()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, fmt.Sprintf(
				"/%s?pre=1",
				day.Format("20060102"),
			), r.RequestURI)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("0"))
		}))

		isDayOffClient := NewIsDayOff(srv.Client(), srv.URL)
		got, err := isDayOffClient.FetchDayType(context.Background(), time.Now())
		require.NoError(t, err)
		require.Equal(t, model.WorkDay, got)
	})
}

func TestIsDayOff_FetchDayTypeRespStatusCode(t *testing.T) {
	day := time.Now()

	type args struct {
		srv *httptest.Server
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "non-integer code",
			args: args{srv: func() *httptest.Server {
				srv := httptest.NewServer(http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					require.Equal(t, fmt.Sprintf("/%s?pre=1", day.Format("20060102")), r.RequestURI)

					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("abc"))
				}))

				return srv
			}()},
			wantErr: ErrNonIntegerCode,
		},
		{
			name: "invalid day format",
			args: args{srv: func() *httptest.Server {
				srv := httptest.NewServer(http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					require.Equal(t, fmt.Sprintf("/%s?pre=1", day.Format("20060102")), r.RequestURI)

					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("100"))
				}))

				return srv
			}()},
			wantErr: ErrBadDayFormat,
		},
		{
			name: "day data not found",
			args: args{srv: func() *httptest.Server {
				srv := httptest.NewServer(http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					require.Equal(t, fmt.Sprintf("/%s?pre=1", day.Format("20060102")), r.RequestURI)

					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("101"))
				}))

				return srv
			}()},
			wantErr: ErrDataNotFound,
		},
		{
			name: "service unavailable #1",
			args: args{srv: func() *httptest.Server {
				srv := httptest.NewServer(http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					require.Equal(t, fmt.Sprintf("/%s?pre=1", day.Format("20060102")), r.RequestURI)

					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("199"))
				}))

				return srv
			}()},
			wantErr: ErrServiceUnavailable,
		},
		{
			name: "service unavailable #2",
			args: args{srv: func() *httptest.Server {
				srv := httptest.NewServer(http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					require.Equal(t, fmt.Sprintf("/%s?pre=1", day.Format("20060102")), r.RequestURI)

					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte("200"))
				}))

				return srv
			}()},
			wantErr: ErrServiceUnavailable,
		},
		{
			name: "unknown",
			args: args{srv: func() *httptest.Server {
				srv := httptest.NewServer(http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					require.Equal(t, fmt.Sprintf("/%s?pre=1", day.Format("20060102")), r.RequestURI)

					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("777"))
				}))

				return srv
			}()},
			wantErr: ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isDayOffClient := NewIsDayOff(tt.args.srv.Client(), tt.args.srv.URL)
			dt, err := isDayOffClient.FetchDayType(context.Background(), time.Now())
			require.Equal(t, model.DayError, dt)
			require.EqualError(t, err, tt.wantErr.Error())
		})
	}
}
