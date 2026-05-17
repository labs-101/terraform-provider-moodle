package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Choicegroup struct {
	CMID                    int64   `json:"cmid"`
	ChoicegroupID           int64   `json:"choicegroupid"`
	Name                    string  `json:"name"`
	Intro                   string  `json:"intro"`
	GroupIDs                []int64 `json:"groupids"`
	MultipleEnrollments     bool    `json:"multipleenrollmentspossible"`
	ShowResults             int64   `json:"showresults"`
	AllowUpdate             bool    `json:"allowupdate"`
	TimeOpen                int64   `json:"timeopen"`
	TimeClose               int64   `json:"timeclose"`
	Section                 int64   `json:"section"`
	Visible                 bool    `json:"visible"`
}

type CreateChoicegroupResponse struct {
	CMID          int64  `json:"cmid"`
	ChoicegroupID int64  `json:"choicegroupid"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}

// GetChoicegroup reads the current state of a choicegroup activity by course module ID.
func (c *MoodleClient) GetChoicegroup(courseID, cmID int64) (*Choicegroup, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_choicegroup")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetChoicegroup cm=%d: %w", cmID, err)
	}

	var raw struct {
		CMID                    int64   `json:"cmid"`
		ChoicegroupID           int64   `json:"choicegroupid"`
		Name                    string  `json:"name"`
		Intro                   string  `json:"intro"`
		GroupIDs                []int64 `json:"groupids"`
		MultipleEnrollments     int64   `json:"multipleenrollmentspossible"`
		ShowResults             int64   `json:"showresults"`
		AllowUpdate             int64   `json:"allowupdate"`
		TimeOpen                int64   `json:"timeopen"`
		TimeClose               int64   `json:"timeclose"`
		Section                 int64   `json:"section"`
		Visible                 int64   `json:"visible"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing choicegroup: %w — body: %s", err, string(body))
	}

	return &Choicegroup{
		CMID:                raw.CMID,
		ChoicegroupID:       raw.ChoicegroupID,
		Name:                raw.Name,
		Intro:               raw.Intro,
		GroupIDs:            raw.GroupIDs,
		MultipleEnrollments: raw.MultipleEnrollments == 1,
		ShowResults:         raw.ShowResults,
		AllowUpdate:         raw.AllowUpdate == 1,
		TimeOpen:            raw.TimeOpen,
		TimeClose:           raw.TimeClose,
		Section:             raw.Section,
		Visible:             raw.Visible == 1,
	}, nil
}

func (c *MoodleClient) CreateChoicegroup(
	courseID int64,
	sectionNum int64,
	name string,
	intro string,
	groupIDs []int64,
	multipleEnrollments bool,
	showResults int64,
	allowUpdate bool,
	timeOpen int64,
	timeClose int64,
	visible bool,
	previousElementID int64,
) (int64, error) {
	multipleInt := 0
	if multipleEnrollments {
		multipleInt = 1
	}
	allowUpdateInt := 0
	if allowUpdate {
		allowUpdateInt = 1
	}
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_create_choicegroup")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("sectionnum", fmt.Sprintf("%d", sectionNum))
	params.Add("name", name)
	params.Add("intro", intro)
	for i, gid := range groupIDs {
		params.Add(fmt.Sprintf("groupids[%d]", i), fmt.Sprintf("%d", gid))
	}
	params.Add("multipleenrollmentspossible", fmt.Sprintf("%d", multipleInt))
	params.Add("showresults", fmt.Sprintf("%d", showResults))
	params.Add("allowupdate", fmt.Sprintf("%d", allowUpdateInt))
	params.Add("timeopen", fmt.Sprintf("%d", timeOpen))
	params.Add("timeclose", fmt.Sprintf("%d", timeClose))
	params.Add("visible", fmt.Sprintf("%d", visibleInt))
	params.Add("previousElementId", fmt.Sprintf("%d", previousElementID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return 0, fmt.Errorf("CreateChoicegroup: %w", err)
	}

	var result CreateChoicegroupResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	return result.CMID, nil
}

func (c *MoodleClient) UpdateChoicegroup(
	courseID int64,
	cmID int64,
	name string,
	intro string,
	groupIDs []int64,
	multipleEnrollments bool,
	showResults int64,
	allowUpdate bool,
	timeOpen int64,
	timeClose int64,
	visible bool,
	section int64,
	previousElementID int64,
) error {
	multipleInt := 0
	if multipleEnrollments {
		multipleInt = 1
	}
	allowUpdateInt := 0
	if allowUpdate {
		allowUpdateInt = 1
	}
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_choicegroup")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))
	params.Add("name", name)
	params.Add("intro", intro)
	for i, gid := range groupIDs {
		params.Add(fmt.Sprintf("groupids[%d]", i), fmt.Sprintf("%d", gid))
	}
	params.Add("multipleenrollmentspossible", fmt.Sprintf("%d", multipleInt))
	params.Add("showresults", fmt.Sprintf("%d", showResults))
	params.Add("allowupdate", fmt.Sprintf("%d", allowUpdateInt))
	params.Add("timeopen", fmt.Sprintf("%d", timeOpen))
	params.Add("timeclose", fmt.Sprintf("%d", timeClose))
	params.Add("visible", fmt.Sprintf("%d", visibleInt))
	params.Add("section", fmt.Sprintf("%d", section))
	params.Add("previousElementId", fmt.Sprintf("%d", previousElementID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("UpdateChoicegroup cm=%d: %w", cmID, err)
	}

	var result GenericSuccessResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("UpdateChoicegroup cm=%d: moodle returned success=false", cmID)
	}

	return nil
}

// DeleteModule deletes any course module via the universal delete endpoint.
func (c *MoodleClient) DeleteModule(courseID, cmID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_delete_module")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("DeleteModule cm=%d: %w", cmID, err)
	}

	return nil
}
