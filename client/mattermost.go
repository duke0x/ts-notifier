package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/duke0x/ts-notifier/config"
)

type Mattermost struct {
	client *http.Client
	config config.Mattermost
}

func NewNotifier(client *http.Client, cfg config.Mattermost) *Mattermost {
	return &Mattermost{client: client, config: cfg}
}

type CreatePostRequest struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
}

type CreatePostResponse struct {
	ID            string      `json:"id"`
	CreateAt      int64       `json:"create_at"`
	UpdateAt      int64       `json:"update_at"`
	EditAt        int         `json:"edit_at"`
	DeleteAt      int         `json:"delete_at"`
	IsPinned      bool        `json:"is_pinned"`
	UserID        string      `json:"user_id"`
	ChannelID     string      `json:"channel_id"`
	RootID        string      `json:"root_id"`
	OriginalID    string      `json:"original_id"`
	Message       string      `json:"message"`
	Type          string      `json:"type"`
	Hashtags      string      `json:"hashtags"`
	PendingPostID string      `json:"pending_post_id"`
	ReplyCount    int         `json:"reply_count"`
	LastReplyAt   int         `json:"last_reply_at"`
	Participants  interface{} `json:"participants"`
}

func (c *Mattermost) Notify(channel, message string) error {
	url := strings.Join([]string{c.config.URL, "/api/v4/posts"}, "")

	cr := CreatePostRequest{
		ChannelID: channel,
		Message:   message,
	}
	crData, _ := json.Marshal(cr)

	// Create a new HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(crData),
	)
	if err != nil {
		return fmt.Errorf("creating 'create post' request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending 'post message to channel' request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Parse the response
	rspData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("sending 'post message to channel' response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf(
			"mattermost return %d rsp code, expected %d response",
			resp.StatusCode,
			http.StatusCreated,
		)
	}

	var rsp CreatePostResponse
	err = json.Unmarshal(rspData, &rsp)
	if err != nil {
		return fmt.Errorf("parsing 'post message to channel' response: %w", err)
	}

	return nil
}
