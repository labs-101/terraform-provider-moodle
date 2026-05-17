package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Auth      string `json:"auth"`
}

func (c *MoodleClient) CreateUser(username, password, firstname, lastname, email, auth string) (*User, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_user_create_users")
	params.Add("moodlewsrestformat", "json")
	params.Add("users[0][username]", username)
	params.Add("users[0][password]", password)
	params.Add("users[0][firstname]", firstname)
	params.Add("users[0][lastname]", lastname)
	params.Add("users[0][email]", email)
	if auth != "" {
		params.Add("users[0][auth]", auth)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}

	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("CreateUser: moodle returned no user")
	}

	return &users[0], nil
}

func (c *MoodleClient) GetUser(userID int64) (*User, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_user_get_users_by_field")
	params.Add("moodlewsrestformat", "json")
	params.Add("field", "id")
	params.Add("values[0]", fmt.Sprintf("%d", userID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetUser %d: %w", userID, err)
	}

	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user with ID %d not found", userID)
	}

	return &users[0], nil
}

func (c *MoodleClient) GetUserByEmail(email string) (*User, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_user_get_users_by_field")
	params.Add("moodlewsrestformat", "json")
	params.Add("field", "email")
	params.Add("values[0]", email)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetUserByEmail %q: %w", email, err)
	}

	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user with email %q not found", email)
	}

	return &users[0], nil
}

func (c *MoodleClient) UpdateUser(userID int64, firstname, lastname, email string) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_user_update_users")
	params.Add("moodlewsrestformat", "json")
	params.Add("users[0][id]", fmt.Sprintf("%d", userID))
	params.Add("users[0][firstname]", firstname)
	params.Add("users[0][lastname]", lastname)
	params.Add("users[0][email]", email)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("UpdateUser %d: %w", userID, err)
	}

	return nil
}

func (c *MoodleClient) DeleteUser(userID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_user_delete_users")
	params.Add("moodlewsrestformat", "json")
	params.Add("userids[0]", fmt.Sprintf("%d", userID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("DeleteUser %d: %w", userID, err)
	}

	return nil
}
