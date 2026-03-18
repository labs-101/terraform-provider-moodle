package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type EnrolledUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Roles    []struct {
		RoleID int64 `json:"roleid"`
	} `json:"roles"`
}

func (c *MoodleClient) EnrolUser(userID, courseID, roleID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "enrol_manual_enrol_users")
	params.Add("moodlewsrestformat", "json")
	params.Add("enrolments[0][roleid]", fmt.Sprintf("%d", roleID))
	params.Add("enrolments[0][userid]", fmt.Sprintf("%d", userID))
	params.Add("enrolments[0][courseid]", fmt.Sprintf("%d", courseID))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return fmt.Errorf("moodle API error: %s", string(body))
	}

	return nil
}

func (c *MoodleClient) UnenrolUser(userID, courseID, roleID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "enrol_manual_unenrol_users")
	params.Add("moodlewsrestformat", "json")
	params.Add("enrolments[0][userid]", fmt.Sprintf("%d", userID))
	params.Add("enrolments[0][courseid]", fmt.Sprintf("%d", courseID))
	if roleID != 0 {
		params.Add("enrolments[0][roleid]", fmt.Sprintf("%d", roleID))
	}

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return fmt.Errorf("moodle API error: %s", string(body))
	}

	return nil
}

func (c *MoodleClient) GetEnrolledUsers(courseID int64) ([]EnrolledUser, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_enrol_get_enrolled_users")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return nil, fmt.Errorf("moodle API error: %s", string(body))
	}

	var users []EnrolledUser
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return users, nil
}
