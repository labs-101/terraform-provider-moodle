package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CreateChoicegroupResponse struct {
	CMID          int64  `json:"cmid"`
	ChoicegroupID int64  `json:"choicegroupid"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}

func (c *MoodleClient) CreateChoicegroup(
	courseID int64,
	sectionNum int64,
	name string,
	intro string,
	groupIDs []int64,
	multipleEnrollments int64,
	showResults int64,
	allowUpdate int64,
	timeOpen int64,
	timeClose int64,
	visible int64,
	previousElementID int64,
) (int64, error) {
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
	params.Add("multipleenrollmentspossible", fmt.Sprintf("%d", multipleEnrollments))
	params.Add("showresults", fmt.Sprintf("%d", showResults))
	params.Add("allowupdate", fmt.Sprintf("%d", allowUpdate))
	params.Add("timeopen", fmt.Sprintf("%d", timeOpen))
	params.Add("timeclose", fmt.Sprintf("%d", timeClose))
	params.Add("visible", fmt.Sprintf("%d", visible))
	params.Add("previousElementId", fmt.Sprintf("%d", previousElementID))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php", c.Host)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading API response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return 0, fmt.Errorf("moodle API error creating choicegroup: %s", string(body))
	}

	var result CreateChoicegroupResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	return result.CMID, nil
}

func (c *MoodleClient) UpdateChoicegroup(
	courseID int64,
	cmID int64,
	name string,
	intro string,
	groupIDs []int64,
	multipleEnrollments int64,
	showResults int64,
	allowUpdate int64,
	timeOpen int64,
	timeClose int64,
	visible int64,
	section int64,
	previousElementID int64,
) error {
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
	params.Add("multipleenrollmentspossible", fmt.Sprintf("%d", multipleEnrollments))
	params.Add("showresults", fmt.Sprintf("%d", showResults))
	params.Add("allowupdate", fmt.Sprintf("%d", allowUpdate))
	params.Add("timeopen", fmt.Sprintf("%d", timeOpen))
	params.Add("timeclose", fmt.Sprintf("%d", timeClose))
	params.Add("visible", fmt.Sprintf("%d", visible))
	params.Add("section", fmt.Sprintf("%d", section))
	params.Add("previousElementId", fmt.Sprintf("%d", previousElementID))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php", c.Host)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading API response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return fmt.Errorf("moodle API error updating choicegroup: %s", string(body))
	}

	var result GenericSuccessResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("moodle API returned success=false for choicegroup update")
	}

	return nil
}

// DeleteModule deletes any course module using the universal delete endpoint.
func (c *MoodleClient) DeleteModule(courseID, cmID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_delete_module")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php", c.Host)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading API response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return fmt.Errorf("moodle API error deleting module (cmid=%d): %s", cmID, string(body))
	}

	return nil
}
