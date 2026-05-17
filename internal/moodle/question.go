package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type AddQuestionResponse struct {
	QuestionID int64 `json:"questionid"`
	Slot       int64 `json:"slot"`
	Page       int64 `json:"page"`
}

type UpdateQuestionResponse struct {
	QuestionID int64 `json:"questionid"`
	Slot       int64 `json:"slot"`
	Page       int64 `json:"page"`
	Success    bool  `json:"success"`
}

type DeleteQuestionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *MoodleClient) AddQuestionToQuiz(
	courseID int64,
	name string,
	quizID int64,
	questionType string,
	choices string,
	page int64,
) (*AddQuestionResponse, error) {
	mu := c.quizQuestionLock(quizID)
	mu.Lock()
	defer mu.Unlock()

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_add_question_to_quiz")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("name", name)
	params.Add("quizid", fmt.Sprintf("%d", quizID))
	params.Add("type", questionType)
	params.Add("choices", choices)
	params.Add("page", fmt.Sprintf("%d", page))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("AddQuestionToQuiz quiz=%d: %w", quizID, err)
	}

	var result AddQuestionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	return &result, nil
}

func (c *MoodleClient) UpdateQuestion(
	courseID int64,
	quizID int64,
	questionID int64,
	name string,
	questionType string,
	choices string,
	slot int64,
	page int64,
) (*UpdateQuestionResponse, error) {
	mu := c.quizQuestionLock(quizID)
	mu.Lock()
	defer mu.Unlock()

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_question")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("quizid", fmt.Sprintf("%d", quizID))
	params.Add("questionid", fmt.Sprintf("%d", questionID))
	params.Add("name", name)
	params.Add("type", questionType)
	params.Add("choices", choices)
	params.Add("slot", fmt.Sprintf("%d", slot))
	params.Add("page", fmt.Sprintf("%d", page))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("UpdateQuestion %d: %w", questionID, err)
	}

	var result UpdateQuestionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if !result.Success {
		return nil, fmt.Errorf("UpdateQuestion %d: moodle returned success=false", questionID)
	}

	return &result, nil
}

func (c *MoodleClient) DeleteQuestion(courseID int64, quizID int64, questionID int64) error {
	mu := c.quizQuestionLock(quizID)
	mu.Lock()
	defer mu.Unlock()

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_delete_question")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("quizid", fmt.Sprintf("%d", quizID))
	params.Add("questionid", fmt.Sprintf("%d", questionID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("DeleteQuestion %d: %w", questionID, err)
	}

	return nil
}
