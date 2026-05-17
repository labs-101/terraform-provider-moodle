package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Quiz struct {
	QuizID           int64  `json:"quizid"`
	CMID             int64  `json:"cmid"`
	Name             string `json:"name"`
	Intro            string `json:"intro"`
	QuizPassword     string `json:"quizpassword"`
	TimeOpen         int64  `json:"timeopen"`
	TimeClose        int64  `json:"timeclose"`
	TimeLimit        int64  `json:"timelimit"`
	Attempts         int64  `json:"attempts"`
	GradeMethod      int64  `json:"grademethod"`
	QuestionsPerPage int64  `json:"questionsperpage"`
	NavMethod        string `json:"navmethod"`
	Section          int64  `json:"section"`
	Visible          bool   `json:"visible"`
}

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

func (c *MoodleClient) GetQuiz(courseID, quizID int64) (*Quiz, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_quiz")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("quizid", fmt.Sprintf("%d", quizID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetQuiz quiz=%d: %w", quizID, err)
	}

	var raw struct {
		QuizID           int64  `json:"quizid"`
		CMID             int64  `json:"cmid"`
		Name             string `json:"name"`
		Intro            string `json:"intro"`
		QuizPassword     string `json:"quizpassword"`
		TimeOpen         int64  `json:"timeopen"`
		TimeClose        int64  `json:"timeclose"`
		TimeLimit        int64  `json:"timelimit"`
		Attempts         int64  `json:"attempts"`
		GradeMethod      int64  `json:"grademethod"`
		QuestionsPerPage int64  `json:"questionsperpage"`
		NavMethod        string `json:"navmethod"`
		Section          int64  `json:"section"`
		Visible          int64  `json:"visible"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing quiz: %w — body: %s", err, string(body))
	}

	return &Quiz{
		QuizID:           raw.QuizID,
		CMID:             raw.CMID,
		Name:             raw.Name,
		Intro:            raw.Intro,
		QuizPassword:     raw.QuizPassword,
		TimeOpen:         raw.TimeOpen,
		TimeClose:        raw.TimeClose,
		TimeLimit:        raw.TimeLimit,
		Attempts:         raw.Attempts,
		GradeMethod:      raw.GradeMethod,
		QuestionsPerPage: raw.QuestionsPerPage,
		NavMethod:        raw.NavMethod,
		Section:          raw.Section,
		Visible:          raw.Visible == 1,
	}, nil
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
	visible bool,
) (int64, error) {
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

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
	params.Add("visible", fmt.Sprintf("%d", visibleInt))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return 0, fmt.Errorf("CreateQuiz: %w", err)
	}

	var result CreateQuizResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
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
	visible bool,
) error {
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

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
	params.Add("visible", fmt.Sprintf("%d", visibleInt))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("UpdateQuiz %d: %w", quizID, err)
	}

	var result UpdateQuizResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("UpdateQuiz %d: moodle returned success=false", quizID)
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("DeleteQuiz %d: %w", quizID, err)
	}

	var result DeleteQuizResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if !result.Success {
		return fmt.Errorf("DeleteQuiz %d: moodle returned success=false: %s", quizID, result.Message)
	}

	return nil
}
