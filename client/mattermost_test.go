package client

import (
	"encoding/json"
	"errors"
	"github.com/duke0x/ts-notifier/config"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMattermost_Notify(t *testing.T) {
	type args struct {
		channel string
		message string
	}
	tests := []struct {
		name    string
		handler http.HandlerFunc
		args    args
		wantErr error
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/api/v4/posts", r.RequestURI)
				require.Equal(t, http.MethodPost, r.Method)
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				var cpr CreatePostRequest
				err = json.Unmarshal(bodyBytes, &cpr)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
				}
				require.NotEmpty(t, cpr.ChannelID)
				require.NotEmpty(t, cpr.Message)

				w.WriteHeader(http.StatusCreated)
				resp := CreatePostResponse{}
				respBytes, _ := json.Marshal(resp)
				_, _ = w.Write(respBytes)
			},
			args: args{
				channel: "ch1",
				message: "notification for team",
			},
			wantErr: nil,
		},
		{
			name: "http not 201",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/api/v4/posts", r.RequestURI)
				require.Equal(t, http.MethodPost, r.Method)
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				var cpr CreatePostRequest
				err = json.Unmarshal(bodyBytes, &cpr)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
				}
				require.NotEmpty(t, cpr.ChannelID)
				require.NotEmpty(t, cpr.Message)

				w.WriteHeader(http.StatusServiceUnavailable)
				//resp := CreatePostResponse{}
				//respBytes, _ := json.Marshal(resp)
				//_, _ = w.Write(respBytes)
			},
			args: args{
				channel: "ch1",
				message: "notification for team",
			},
			wantErr: errors.New("mattermost return 503 rsp code, expected 201 response"),
		}, {
			name: "srv not exist",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				time.Sleep(4 * time.Second)
			},
			args: args{
				channel: "ch1",
				message: "message",
			},
			wantErr: errors.New("sending 'post message to channel' request: "),
		},
		{
			name: "srv bad response data",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("bad response data"))
			},
			args: args{
				channel: "ch1",
				message: "message",
			},
			wantErr: errors.New("parsing 'post message to channel' response: "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			nm := NewNotifier(srv.Client(), config.Mattermost{
				URL: srv.URL,
			})

			err := nm.Notify(tt.args.channel, tt.args.message)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr.Error())
			}
		})
	}
}
