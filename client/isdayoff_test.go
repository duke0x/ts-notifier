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
