package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("EnrolUser user=%d course=%d: %w", userID, courseID, err)
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("UnenrolUser user=%d course=%d: %w", userID, courseID, err)
	}

	return nil
}

func (c *MoodleClient) GetEnrolledUsers(courseID int64) ([]EnrolledUser, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_enrol_get_enrolled_users")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetEnrolledUsers course=%d: %w", courseID, err)
	}

	var users []EnrolledUser
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("parsing enrolled users: %w — body: %s", err, string(body))
	}

	return users, nil
}
