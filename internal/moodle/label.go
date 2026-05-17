package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Label struct {
	CMID    int64  `json:"cmid"`
	Name    string `json:"name"`
	Intro   string `json:"intro"`
	Section int64  `json:"section"`
	Visible bool   `json:"visible"`
}

// GetLabel reads the current state of a label activity by course module ID.
func (c *MoodleClient) GetLabel(courseID, cmID int64) (*Label, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_label")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetLabel cm=%d: %w", cmID, err)
	}

	var raw struct {
		CMID    int64  `json:"cmid"`
		Name    string `json:"name"`
		Intro   string `json:"intro"`
		Section int64  `json:"section"`
		Visible int64  `json:"visible"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing label: %w — body: %s", err, string(body))
	}

	return &Label{
		CMID:    raw.CMID,
		Name:    raw.Name,
		Intro:   raw.Intro,
		Section: raw.Section,
		Visible: raw.Visible == 1,
	}, nil
}

// CreateLabel creates a label activity in a course section and returns the cmID.
func (c *MoodleClient) CreateLabel(courseID int64, sectionNum int64, intro string, name string, previousElementId int64) (int64, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_create_label")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("sectionnum", fmt.Sprintf("%d", sectionNum))
	params.Add("intro", intro)
	params.Add("name", name)
	params.Add("previousElementId", fmt.Sprintf("%d", previousElementId))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return 0, fmt.Errorf("CreateLabel: %w", err)
	}

	var result struct {
		CMID    int64 `json:"cmid"`
		Visible bool  `json:"visible"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	return result.CMID, nil
}

// UpdateLabel updates an existing label activity.
func (c *MoodleClient) UpdateLabel(courseID int64, cmID int64, intro string, name string, sectionNum int64, previousElementId int64, visible bool) error {
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_label")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))
	params.Add("intro", intro)
	params.Add("name", name)
	params.Add("section", fmt.Sprintf("%d", sectionNum))
	params.Add("previousElementId", fmt.Sprintf("%d", previousElementId))
	params.Add("visible", fmt.Sprintf("%d", visibleInt))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("UpdateLabel cm=%d: %w", cmID, err)
	}

	return nil
}

// DeleteLabel deletes a label activity from a course.
func (c *MoodleClient) DeleteLabel(courseID int64, cmID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_delete_label")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("DeleteLabel cm=%d: %w", cmID, err)
	}

	return nil
}
