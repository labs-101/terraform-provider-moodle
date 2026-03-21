package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Course struct {
	Id         int64  `json:"id"`
	Shortname  string `json:"shortname"`
	Fullname   string `json:"fullname"`
	Idnumber   string `json:"idnumber"`
	Summary    string `json:"summary"`
	Visibility int64  `json:"visible"`
	StartDate  int64  `json:"startdate"`
	EndDate    int64  `json:"enddate"`
}

func (c *MoodleClient) GetAllCourses() ([]Course, error) {
	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?wstoken=%s&wsfunction=core_course_get_courses&moodlewsrestformat=json", c.Host, c.Token)

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

	var courses []Course
	err = json.Unmarshal(body, &courses)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON (invalid format?): %w\nBody was: %s", err, string(body))
	}

	return courses, nil
}

func (c *MoodleClient) CreateCourse(fullname, shortname string, categoryID int64, idnumber string, summary string, visibility int64, startdate int64, enddate int64) (*Course, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_create_courses")
	params.Add("moodlewsrestformat", "json")
	params.Add("courses[0][fullname]", fullname)
	params.Add("courses[0][shortname]", shortname)
	params.Add("courses[0][categoryid]", fmt.Sprintf("%d", categoryID))
	params.Add("courses[0][idnumber]", idnumber)
	params.Add("courses[0][summary]", summary)
	params.Add("courses[0][visible]", fmt.Sprintf("%d", visibility))
	params.Add("courses[0][startdate]", fmt.Sprintf("%d", startdate))
	params.Add("courses[0][enddate]", fmt.Sprintf("%d", enddate))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Senden des Requests: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der API-Antwort: %w", err)
	}

	// TODO create utils function for default moodle errors
	if strings.Contains(string(body), "exception") {
		return nil, fmt.Errorf("moodle API Fehler: %s", string(body))
	}

	var courses []Course
	err = json.Unmarshal(body, &courses)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Parsen des JSON: %w\nBody: %s", err, string(body))
	}

	if len(courses) == 0 {
		return nil, fmt.Errorf("moodle hat keinen Kurs zurückgegeben")
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

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Senden des Requests: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der API-Antwort: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return nil, fmt.Errorf("moodle API Fehler: %s", string(body))
	}

	var result struct {
		Courses []Course `json:"courses"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Parsen des JSON: %w\nBody: %s", err, string(body))
	}

	if len(result.Courses) == 0 {
		return nil, fmt.Errorf("kurs mit ID %d nicht gefunden", id)
	}

	return &result.Courses[0], nil
}

func (c *MoodleClient) DeleteCourse(id int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_delete_courses")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseids[0]", fmt.Sprintf("%d", id))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("fehler beim Senden des Requests: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("fehler beim Lesen der API-Antwort: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return fmt.Errorf("moodle API Fehler: %s", string(body))
	}

	type MoodleWarning struct {
		Item        string `json:"item,omitempty"`
		ItemID      int    `json:"itemid,omitempty"`
		WarningCode string `json:"warningcode,omitempty"`
		Message     string `json:"message,omitempty"`
	}

	var result struct {
		Warnings []MoodleWarning `json:"warnings"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("fehler beim Parsen des JSON: %w\nBody: %s", err, string(body))
	}

	return nil
}

func (c *MoodleClient) UpdateCourse(id int64, fullname, shortname string, categoryID int64, idnumber string, summary string, visibility int64, startdate int64, enddate int64) error {
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
	params.Add("courses[0][visible]", fmt.Sprintf("%d", visibility))
	params.Add("courses[0][startdate]", fmt.Sprintf("%d", startdate))
	params.Add("courses[0][enddate]", fmt.Sprintf("%d", enddate))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php", c.Host)

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Update-Requests: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("fehler beim Senden des Update-Requests: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("fehler beim Lesen der API-Antwort: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return fmt.Errorf("moodle API Fehler beim Update: %s", string(body))
	}

	return nil
}

func (c *MoodleClient) GetCourseModule(courseID int64, cmID int64) (*CourseModule, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_get_contents")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Senden des Requests: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der API-Antwort: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return nil, fmt.Errorf("moodle API Fehler beim Lesen der Kursinhalte: %s", string(body))
	}

	var sections []struct {
		Modules []struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			ModName string `json:"modname"`
		} `json:"modules"`
	}
	if err := json.Unmarshal(body, &sections); err != nil {
		return nil, fmt.Errorf("fehler beim Parsen der Kursinhalte: %w\nBody: %s", err, string(body))
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

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php", c.Host)

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Delete-Requests: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("fehler beim Senden des Delete-Requests: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("fehler beim Lesen der Delete-Antwort: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return fmt.Errorf("moodle API Fehler beim Löschen des Moduls: %s", string(body))
	}

	return nil
}
