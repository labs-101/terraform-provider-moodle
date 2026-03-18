package moodle

import (
	"encoding/json"
	"fmt"
	"io"
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

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("POST", reqURL, nil)
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
		return nil, fmt.Errorf("error reading API response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return nil, fmt.Errorf("moodle API error: %s", string(body))
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w\nBody: %s", err, string(body))
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("moodle returned no user")
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
		return nil, fmt.Errorf("error reading API response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return nil, fmt.Errorf("moodle API error: %s", string(body))
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w\nBody: %s", err, string(body))
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
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
		return nil, fmt.Errorf("error reading API response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return nil, fmt.Errorf("moodle API error: %s", string(body))
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w\nBody: %s", err, string(body))
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

func (c *MoodleClient) DeleteUser(userID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_user_delete_users")
	params.Add("moodlewsrestformat", "json")
	params.Add("userids[0]", fmt.Sprintf("%d", userID))

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
		return fmt.Errorf("error reading API response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return fmt.Errorf("moodle API error: %s", string(body))
	}

	return nil
}
