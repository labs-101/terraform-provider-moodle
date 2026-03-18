package moodle

import (
	"net/http"
	"strings"
	"time"
)

type MoodleClient struct {
	Host          string
	Token         string
	MoodleVersion string
	HTTPClient    *http.Client
}

func NewMoodleClient(host string, token string, moodleVersion string) (*MoodleClient, error) {
	cleanHost := strings.TrimRight(host, "/")

	return &MoodleClient{
		Host:          cleanHost,
		Token:         token,
		MoodleVersion: moodleVersion,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}
