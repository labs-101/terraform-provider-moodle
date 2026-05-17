package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Assignment struct {
	CMID                     int64  `json:"cmid"`
	Name                     string `json:"name"`
	Intro                    string `json:"intro"`
	DueDate                  int64  `json:"duedate"`
	AllowSubmissionsFromDate int64  `json:"allowsubmissionsfromdate"`
	MaxBytes                 int64  `json:"maxbytes"`
	MaxFileSubmissions       int64  `json:"maxfilesubmissions"`
	SubmissionTypes          string `json:"submissiontypes"`
	Section                  int64  `json:"section"`
	Visible                  bool   `json:"visible"`
}

// GetAssignment reads the current state of an assignment by its course module ID.
func (c *MoodleClient) GetAssignment(courseID, cmID int64) (*Assignment, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_assignment")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetAssignment cm=%d: %w", cmID, err)
	}

	var raw struct {
		CMID                     int64  `json:"cmid"`
		Name                     string `json:"name"`
		Intro                    string `json:"intro"`
		DueDate                  int64  `json:"duedate"`
		AllowSubmissionsFromDate int64  `json:"allowsubmissionsfromdate"`
		MaxBytes                 int64  `json:"maxbytes"`
		MaxFileSubmissions       int64  `json:"maxfilesubmissions"`
		SubmissionTypes          string `json:"submissiontypes"`
		Section                  int64  `json:"section"`
		Visible                  int64  `json:"visible"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing assignment: %w — body: %s", err, string(body))
	}

	return &Assignment{
		CMID:                     raw.CMID,
		Name:                     raw.Name,
		Intro:                    raw.Intro,
		DueDate:                  raw.DueDate,
		AllowSubmissionsFromDate: raw.AllowSubmissionsFromDate,
		MaxBytes:                 raw.MaxBytes,
		MaxFileSubmissions:       raw.MaxFileSubmissions,
		SubmissionTypes:          raw.SubmissionTypes,
		Section:                  raw.Section,
		Visible:                  raw.Visible == 1,
	}, nil
}

// AddAssignmentToSection creates an assignment activity in a course section and returns the cmID.
func (c *MoodleClient) AddAssignmentToSection(
	courseID int64,
	sectionNum int64,
	name string,
	intro string,
	dueDate int64,
	allowSubmissionsFromDate int64,
	maxBytes int64,
	maxFileSubmissions int64,
	submissionTypes string,
) (int64, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_add_assignment_to_section")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("sectionnum", fmt.Sprintf("%d", sectionNum))
	params.Add("name", name)
	params.Add("intro", intro)
	params.Add("duedate", fmt.Sprintf("%d", dueDate))
	params.Add("allowsubmissionsfromdate", fmt.Sprintf("%d", allowSubmissionsFromDate))
	params.Add("maxbytes", fmt.Sprintf("%d", maxBytes))
	params.Add("maxfilesubmissions", fmt.Sprintf("%d", maxFileSubmissions))
	params.Add("submissiontypes", submissionTypes)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return 0, fmt.Errorf("AddAssignmentToSection: %w", err)
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

// UpdateAssignment updates an existing assignment activity.
func (c *MoodleClient) UpdateAssignment(
	courseID int64,
	cmID int64,
	name string,
	intro string,
	dueDate int64,
	allowSubmissionsFromDate int64,
	maxBytes int64,
	maxFileSubmissions int64,
	submissionTypes string,
	visible bool,
) error {
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_assignment")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))
	params.Add("name", name)
	params.Add("intro", intro)
	params.Add("duedate", fmt.Sprintf("%d", dueDate))
	params.Add("allowsubmissionsfromdate", fmt.Sprintf("%d", allowSubmissionsFromDate))
	params.Add("maxbytes", fmt.Sprintf("%d", maxBytes))
	params.Add("maxfilesubmissions", fmt.Sprintf("%d", maxFileSubmissions))
	params.Add("submissiontypes", submissionTypes)
	params.Add("visible", fmt.Sprintf("%d", visibleInt))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("UpdateAssignment cm=%d: %w", cmID, err)
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("UpdateAssignment cm=%d: moodle returned success=false", cmID)
	}

	return nil
}
