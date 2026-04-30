package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CreateGroupResponse struct {
	GroupID int64  `json:"groupid"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GenericSuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *MoodleClient) CreateGroup(
	courseID int64,
	name string,
	description string,
	enrolmentkey string,
	visibility int64,
	participation int64,
	idnumber string,
) (int64, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_create_group")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("name", name)
	params.Add("description", description)
	params.Add("enrolmentkey", enrolmentkey)
	params.Add("visibility", fmt.Sprintf("%d", visibility))
	params.Add("participation", fmt.Sprintf("%d", participation))
	params.Add("idnumber", idnumber)

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
		return 0, fmt.Errorf("moodle API error creating group: %s", string(body))
	}

	var result CreateGroupResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	return result.GroupID, nil
}

func (c *MoodleClient) UpdateGroup(
	groupID int64,
	name string,
	description string,
	enrolmentkey string,
	visibility int64,
	participation int64,
	idnumber string,
) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_group")
	params.Add("moodlewsrestformat", "json")
	params.Add("groupid", fmt.Sprintf("%d", groupID))
	params.Add("name", name)
	params.Add("description", description)
	params.Add("enrolmentkey", enrolmentkey)
	params.Add("visibility", fmt.Sprintf("%d", visibility))
	params.Add("participation", fmt.Sprintf("%d", participation))
	params.Add("idnumber", idnumber)

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
		return fmt.Errorf("moodle API error updating group: %s", string(body))
	}

	var result GenericSuccessResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("moodle API returned success=false for group update")
	}

	return nil
}

func (c *MoodleClient) DeleteGroup(courseID, groupID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_delete_group")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("groupid", fmt.Sprintf("%d", groupID))

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
		return fmt.Errorf("moodle API error deleting group: %s", string(body))
	}

	var result GenericSuccessResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("moodle API returned success=false for group delete: %s", result.Message)
	}

	return nil
}

func (c *MoodleClient) AddMemberToGroup(groupID, userID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_add_member_to_group")
	params.Add("moodlewsrestformat", "json")
	params.Add("groupid", fmt.Sprintf("%d", groupID))
	params.Add("userid", fmt.Sprintf("%d", userID))

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
		return fmt.Errorf("moodle API error adding member to group: %s", string(body))
	}

	return nil
}

func (c *MoodleClient) RemoveMemberFromGroup(groupID, userID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_remove_member_from_group")
	params.Add("moodlewsrestformat", "json")
	params.Add("groupid", fmt.Sprintf("%d", groupID))
	params.Add("userid", fmt.Sprintf("%d", userID))

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
		return fmt.Errorf("moodle API error removing member from group: %s", string(body))
	}

	return nil
}
