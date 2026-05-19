package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Choice struct {
	CMID          int64    `json:"cmid"`
	Name          string   `json:"name"`
	Intro         string   `json:"intro"`
	Options       []string `json:"options"`
	AllowMultiple bool     `json:"allowmultiple"`
	Section       int64    `json:"section"`
	Visible       bool     `json:"visible"`
}

// GetChoice reads the current state of a choice activity by course module ID.
func (c *MoodleClient) GetChoice(courseID, cmID int64) (*Choice, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_choice")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetChoice cm=%d: %w", cmID, err)
	}

	var raw struct {
		CMID          int64    `json:"cmid"`
		Name          string   `json:"name"`
		Intro         string   `json:"intro"`
		Options       []string `json:"options"`
		AllowMultiple int64    `json:"allowmultiple"`
		Section       int64    `json:"section"`
		Visible       int64    `json:"visible"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing choice: %w — body: %s", err, string(body))
	}

	return &Choice{
		CMID:          raw.CMID,
		Name:          raw.Name,
		Intro:         raw.Intro,
		Options:       raw.Options,
		AllowMultiple: raw.AllowMultiple == 1,
		Section:       raw.Section,
		Visible:       raw.Visible == 1,
	}, nil
}

// AddChoiceToSection creates a choice activity and returns the cmID.
func (c *MoodleClient) AddChoiceToSection(courseID int64, sectionNum int64, name string, intro string, options []string, allowMultiple bool) (int64, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_add_choice_to_section")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("sectionnum", fmt.Sprintf("%d", sectionNum))
	params.Add("name", name)
	params.Add("intro", intro)
	if allowMultiple {
		params.Add("allowmultiple", "1")
	} else {
		params.Add("allowmultiple", "0")
	}
	for i, opt := range options {
		params.Add(fmt.Sprintf("options[%d]", i), opt)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return 0, fmt.Errorf("AddChoiceToSection: %w", err)
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

// UpdateChoice updates an existing choice activity.
func (c *MoodleClient) UpdateChoice(courseID, cmID int64, name, intro string, options []string, allowMultiple bool) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_choice")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))
	params.Add("name", name)
	params.Add("intro", intro)
	if allowMultiple {
		params.Add("allowmultiple", "1")
	} else {
		params.Add("allowmultiple", "0")
	}
	for i, opt := range options {
		params.Add(fmt.Sprintf("options[%d]", i), opt)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("UpdateChoice cm=%d: %w", cmID, err)
	}
	return nil
}
