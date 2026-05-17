package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MoodleClient wraps HTTP access to a Moodle web-service endpoint.
type MoodleClient struct {
	Host          string
	Token         string
	MoodleVersion string
	HTTPClient    *http.Client
	// sectionMu serializes section_add calls per course so concurrent Terraform
	// resource creates never race on the "newest section" response logic.
	sectionMu sync.Map // map[int64]*sync.Mutex
}

// moodleException is the standard error envelope returned by the Moodle REST API.
type moodleException struct {
	Exception string `json:"exception"`
	ErrorCode string `json:"errorcode"`
	Message   string `json:"message"`
	DebugInfo string `json:"debuginfo,omitempty"`
}

func NewMoodleClient(host string, token string, moodleVersion string) (*MoodleClient, error) {
	cleanHost := strings.TrimRight(host, "/")

	return &MoodleClient{
		Host:          cleanHost,
		Token:         token,
		MoodleVersion: moodleVersion,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// courseSectionLock returns the per-course mutex for section_add serialization.
func (c *MoodleClient) courseSectionLock(courseID int64) *sync.Mutex {
	mu, _ := c.sectionMu.LoadOrStore(courseID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// quizQuestionLock returns the per-quiz mutex for question add/update/delete serialization.
// Moodle renumbers mdl_quiz_slots on every mutation, causing deadlocks under concurrency.
func (c *MoodleClient) quizQuestionLock(quizID int64) *sync.Mutex {
	mu, _ := c.sectionMu.LoadOrStore(-quizID, &sync.Mutex{}) // negative key avoids collision with course IDs
	return mu.(*sync.Mutex)
}

// checkMoodleError inspects a raw API response body for a Moodle exception
// and returns a descriptive error when one is found, or nil otherwise.
func checkMoodleError(body []byte) error {
	if !strings.Contains(string(body), `"exception"`) {
		return nil
	}
	var exc moodleException
	if err := json.Unmarshal(body, &exc); err == nil && exc.Exception != "" {
		if exc.DebugInfo != "" {
			return fmt.Errorf("moodle API error [%s]: %s — %s", exc.ErrorCode, exc.Message, exc.DebugInfo)
		}
		return fmt.Errorf("moodle API error [%s]: %s", exc.ErrorCode, exc.Message)
	}
	// Fallback: return the raw body as an error message.
	return fmt.Errorf("moodle API error: %s", string(body))
}

// doRequest executes req, reads the full body, checks for Moodle exceptions,
// and returns the raw bytes. The caller is responsible for JSON unmarshalling.
func (c *MoodleClient) doRequest(req *http.Request) ([]byte, error) {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if err := checkMoodleError(body); err != nil {
		return nil, err
	}

	return body, nil
}
