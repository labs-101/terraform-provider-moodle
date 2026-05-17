package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Group struct {
	GroupID       int64  `json:"groupid"`
	CourseID      int64  `json:"courseid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	EnrolmentKey  string `json:"enrolmentkey"`
	Visibility    int64  `json:"visibility"`
	Participation bool   `json:"participation"`
	IDNumber      string `json:"idnumber"`
}

type CreateGroupResponse struct {
	GroupID int64  `json:"groupid"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GenericSuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetGroup reads the current state of a group by its ID.
func (c *MoodleClient) GetGroup(courseID, groupID int64) (*Group, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_group")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("groupid", fmt.Sprintf("%d", groupID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetGroup group=%d: %w", groupID, err)
	}

	var raw struct {
		GroupID       int64  `json:"groupid"`
		CourseID      int64  `json:"courseid"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		EnrolmentKey  string `json:"enrolmentkey"`
		Visibility    int64  `json:"visibility"`
		Participation int64  `json:"participation"`
		IDNumber      string `json:"idnumber"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing group: %w — body: %s", err, string(body))
	}

	return &Group{
		GroupID:       raw.GroupID,
		CourseID:      raw.CourseID,
		Name:          raw.Name,
		Description:   raw.Description,
		EnrolmentKey:  raw.EnrolmentKey,
		Visibility:    raw.Visibility,
		Participation: raw.Participation == 1,
		IDNumber:      raw.IDNumber,
	}, nil
}

// GetGroupMembers returns the user IDs of all members of a group.
func (c *MoodleClient) GetGroupMembers(courseID, groupID int64) ([]int64, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_group_members")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("groupid", fmt.Sprintf("%d", groupID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetGroupMembers group=%d: %w", groupID, err)
	}

	var userIDs []int64
	if err := json.Unmarshal(body, &userIDs); err != nil {
		return nil, fmt.Errorf("parsing group members: %w — body: %s", err, string(body))
	}

	return userIDs, nil
}

func (c *MoodleClient) CreateGroup(
	courseID int64,
	name string,
	description string,
	enrolmentkey string,
	visibility int64,
	participation bool,
	idnumber string,
) (int64, error) {
	participationInt := 0
	if participation {
		participationInt = 1
	}

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_create_group")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("name", name)
	params.Add("description", description)
	params.Add("enrolmentkey", enrolmentkey)
	params.Add("visibility", fmt.Sprintf("%d", visibility))
	params.Add("participation", fmt.Sprintf("%d", participationInt))
	params.Add("idnumber", idnumber)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return 0, fmt.Errorf("CreateGroup: %w", err)
	}

	var result CreateGroupResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	return result.GroupID, nil
}

func (c *MoodleClient) UpdateGroup(
	groupID int64,
	name string,
	description string,
	enrolmentkey string,
	visibility int64,
	participation bool,
	idnumber string,
) error {
	participationInt := 0
	if participation {
		participationInt = 1
	}

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_group")
	params.Add("moodlewsrestformat", "json")
	params.Add("groupid", fmt.Sprintf("%d", groupID))
	params.Add("name", name)
	params.Add("description", description)
	params.Add("enrolmentkey", enrolmentkey)
	params.Add("visibility", fmt.Sprintf("%d", visibility))
	params.Add("participation", fmt.Sprintf("%d", participationInt))
	params.Add("idnumber", idnumber)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("UpdateGroup %d: %w", groupID, err)
	}

	var result GenericSuccessResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("UpdateGroup %d: moodle returned success=false", groupID)
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("DeleteGroup %d: %w", groupID, err)
	}

	var result GenericSuccessResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("DeleteGroup %d: moodle returned success=false: %s", groupID, result.Message)
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("AddMemberToGroup group=%d user=%d: %w", groupID, userID, err)
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("RemoveMemberFromGroup group=%d user=%d: %w", groupID, userID, err)
	}

	return nil
}
