package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CreateQuizResponse struct {
	QuizID int64 `json:"quizId"`
}

type UpdateQuizResponse struct {
	Success   bool  `json:"success"`
	CMID      int64 `json:"cmid"`
	SectionID int64 `json:"sectionid"`
}

type DeleteQuizResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *MoodleClient) CreateQuiz(
	courseID int64,
	name string,
	intro string,
	password string,
	timeOpen int64,
	timeClose int64,
	timeLimit int64,
	attempts int64,
	gradeMethod int64,
	questionsPerPage int64,
	navMethod string,
	section int64,
	visible int64,
) (int64, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_create_quiz")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("name", name)
	params.Add("intro", intro)
	params.Add("password", password)
	params.Add("timeopen", fmt.Sprintf("%d", timeOpen))
	params.Add("timeclose", fmt.Sprintf("%d", timeClose))
	params.Add("timelimit", fmt.Sprintf("%d", timeLimit))
	params.Add("attempts", fmt.Sprintf("%d", attempts))
	params.Add("grademethod", fmt.Sprintf("%d", gradeMethod))
	params.Add("questionsperpage", fmt.Sprintf("%d", questionsPerPage))
	params.Add("navmethod", navMethod)
	params.Add("section", fmt.Sprintf("%d", section))
	params.Add("visible", fmt.Sprintf("%d", visible))

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
		return 0, fmt.Errorf("moodle API error creating quiz: %s", string(body))
	}

	var result CreateQuizResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	return result.QuizID, nil
}

func (c *MoodleClient) UpdateQuiz(
	quizID int64,
	name string,
	intro string,
	password string,
	timeOpen int64,
	timeClose int64,
	timeLimit int64,
	attempts int64,
	gradeMethod int64,
	questionsPerPage int64,
	navMethod string,
	section int64,
	visible int64,
) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_quiz")
	params.Add("moodlewsrestformat", "json")
	params.Add("quizid", fmt.Sprintf("%d", quizID))
	params.Add("name", name)
	params.Add("intro", intro)
	params.Add("password", password)
	params.Add("timeopen", fmt.Sprintf("%d", timeOpen))
	params.Add("timeclose", fmt.Sprintf("%d", timeClose))
	params.Add("timelimit", fmt.Sprintf("%d", timeLimit))
	params.Add("attempts", fmt.Sprintf("%d", attempts))
	params.Add("grademethod", fmt.Sprintf("%d", gradeMethod))
	params.Add("questionsperpage", fmt.Sprintf("%d", questionsPerPage))
	params.Add("navmethod", navMethod)
	params.Add("section", fmt.Sprintf("%d", section))
	params.Add("visible", fmt.Sprintf("%d", visible))

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
		return fmt.Errorf("moodle API error updating quiz: %s", string(body))
	}

	var result UpdateQuizResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("moodle API returned success=false for quiz update")
	}

	return nil
}

func (c *MoodleClient) DeleteQuiz(courseID int64, quizID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_delete_quiz")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("quizid", fmt.Sprintf("%d", quizID))

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
		return fmt.Errorf("moodle API error deleting quiz: %s", string(body))
	}

	return nil
}
