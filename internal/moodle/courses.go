package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Course represents a Moodle course as returned by core_course_get_courses_by_field.
// Visible is int64 because the core API returns 0/1, not JSON booleans.
type Course struct {
	Id          int64  `json:"id"`
	Shortname   string `json:"shortname"`
	Fullname    string `json:"fullname"`
	Idnumber    string `json:"idnumber"`
	Description string `json:"summary"`
	Visible     int64  `json:"visible"`
	StartDate   int64  `json:"startdate"`
	EndDate     int64  `json:"enddate"`
}

func (c *MoodleClient) GetAllCourses() ([]Course, error) {
	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?wstoken=%s&wsfunction=core_course_get_courses&moodlewsrestformat=json", c.Host, c.Token)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetAllCourses: %w", err)
	}

	var courses []Course
	if err := json.Unmarshal(body, &courses); err != nil {
		return nil, fmt.Errorf("parsing courses: %w — body: %s", err, string(body))
	}

	return courses, nil
}

func (c *MoodleClient) CreateCourse(fullname, shortname string, categoryID int64, idnumber string, summary string, visible bool, startdate int64, enddate int64) (*Course, error) {
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_create_courses")
	params.Add("moodlewsrestformat", "json")
	params.Add("courses[0][fullname]", fullname)
	params.Add("courses[0][shortname]", shortname)
	params.Add("courses[0][categoryid]", fmt.Sprintf("%d", categoryID))
	params.Add("courses[0][idnumber]", idnumber)
	params.Add("courses[0][summary]", summary)
	params.Add("courses[0][visible]", fmt.Sprintf("%d", visibleInt))
	params.Add("courses[0][startdate]", fmt.Sprintf("%d", startdate))
	params.Add("courses[0][enddate]", fmt.Sprintf("%d", enddate))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("CreateCourse: %w", err)
	}

	var courses []Course
	if err := json.Unmarshal(body, &courses); err != nil {
		return nil, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if len(courses) == 0 {
		return nil, fmt.Errorf("CreateCourse: moodle returned no course")
	}

	return &courses[0], nil
}

func (c *MoodleClient) GetCourse(id int64) (*Course, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_get_courses_by_field")
	params.Add("moodlewsrestformat", "json")
	params.Add("field", "id")
	params.Add("value", fmt.Sprintf("%d", id))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetCourse %d: %w", id, err)
	}

	var result struct {
		Courses []Course `json:"courses"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	if len(result.Courses) == 0 {
		return nil, fmt.Errorf("course with ID %d not found", id)
	}

	return &result.Courses[0], nil
}

func (c *MoodleClient) DeleteCourse(id int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_delete_courses")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseids[0]", fmt.Sprintf("%d", id))

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("DeleteCourse %d: %w", id, err)
	}

	return nil
}

func (c *MoodleClient) UpdateCourse(id int64, fullname, shortname string, categoryID int64, idnumber string, summary string, visible bool, startdate int64, enddate int64) error {
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_update_courses")
	params.Add("moodlewsrestformat", "json")
	params.Add("courses[0][id]", fmt.Sprintf("%d", id))
	params.Add("courses[0][fullname]", fullname)
	params.Add("courses[0][shortname]", shortname)
	params.Add("courses[0][categoryid]", fmt.Sprintf("%d", categoryID))
	params.Add("courses[0][idnumber]", idnumber)
	params.Add("courses[0][summary]", summary)
	params.Add("courses[0][visible]", fmt.Sprintf("%d", visibleInt))
	params.Add("courses[0][startdate]", fmt.Sprintf("%d", startdate))
	params.Add("courses[0][enddate]", fmt.Sprintf("%d", enddate))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("UpdateCourse %d: %w", id, err)
	}

	return nil
}

func (c *MoodleClient) GetCourseModule(courseID int64, cmID int64) (*CourseModule, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_get_contents")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetCourseModule course=%d cm=%d: %w", courseID, cmID, err)
	}

	var sections []struct {
		Modules []struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			ModName string `json:"modname"`
		} `json:"modules"`
	}
	if err := json.Unmarshal(body, &sections); err != nil {
		return nil, fmt.Errorf("parsing course contents: %w — body: %s", err, string(body))
	}

	for _, section := range sections {
		for _, mod := range section.Modules {
			if mod.ID == cmID {
				return &CourseModule{
					ID:      mod.ID,
					Name:    mod.Name,
					ModType: mod.ModName,
				}, nil
			}
		}
	}

	return nil, nil
}

func (c *MoodleClient) DeleteCourseModule(cmID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_delete_modules")
	params.Add("moodlewsrestformat", "json")
	params.Add("cmids[0]", fmt.Sprintf("%d", cmID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("DeleteCourseModule cm=%d: %w", cmID, err)
	}

	return nil
}
