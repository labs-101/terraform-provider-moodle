package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Section repräsentiert einen Kursabschnitt in Moodle.
type Section struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Section int64  `json:"section"`
	Summary string `json:"summary"`
	Visible int64  `json:"visible"`
}

// CreateSection fügt eine neue leere Sektion am Ende des Kurses hinzu
// und gibt die neu erstellte Sektion mit ihrer Datenbank-ID zurück.
func (c *MoodleClient) CreateSection(courseID int64) (*Section, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_courseformat_update_course")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("action", "section_add")

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php", c.Host)

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return nil, fmt.Errorf("moodle API error adding section: %s", string(body))
	}

	// The Moodle API returns the response double encoded: a JSON string that itself contains a JSON array.
	// First step: unpack the outer string.
	var rawJSON string
	if err := json.Unmarshal(body, &rawJSON); err != nil {
		// Fallback: maybe it's directly an array
		rawJSON = string(body)
	}

	// Second step: parse the actual array.
	var updates []struct {
		Name   string `json:"name"`
		Action string `json:"action"`
		Fields struct {
			ID      string `json:"id"`
			Section int64  `json:"section"`
			Title   string `json:"title"`
			Visible bool   `json:"visible"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &updates); err != nil {
		return nil, fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	var newest *Section
	for _, u := range updates {
		if u.Name != "section" || u.Action != "put" {
			continue
		}
		id, err := strconv.ParseInt(u.Fields.ID, 10, 64)
		if err != nil {
			continue
		}
		visible := int64(0)
		if u.Fields.Visible {
			visible = 1
		}
		s := &Section{
			ID:      id,
			Name:    u.Fields.Title,
			Section: u.Fields.Section,
			Visible: visible,
		}
		if newest == nil || s.Section > newest.Section {
			newest = s
		}
	}

	if newest == nil {
		return nil, fmt.Errorf("no new section found in API response for course %d", courseID)
	}

	return newest, nil
}

// GetCourseSections GetCourseSections returns all sections of a course (core_course_get_contents).
func (c *MoodleClient) GetCourseSections(courseID int64) ([]Section, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_course_get_contents")
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
		return nil, fmt.Errorf("error reading API response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return nil, fmt.Errorf("moodle API error reading sections: %s", string(body))
	}

	var sections []Section
	if err := json.Unmarshal(body, &sections); err != nil {
		return nil, fmt.Errorf("error parsing sections: %w\nBody: %s", err, string(body))
	}

	return sections, nil
}

// GetSection returns a specific section by its database ID.
func (c *MoodleClient) GetSection(courseID int64, sectionID int64) (*Section, error) {
	sections, err := c.GetCourseSections(courseID)
	if err != nil {
		return nil, err
	}

	for i := range sections {
		if sections[i].ID == sectionID {
			return &sections[i], nil
		}
	}

	return nil, fmt.Errorf("section with ID %d not found in course %d", sectionID, courseID)
}

// EditSection updates name, summary and visibility of a section
// via the Moodle function core_course_edit_section.
func (c *MoodleClient) EditSection(sectionID int64, name string, summary string, visible int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_update_inplace_editable")
	params.Add("moodlewsrestformat", "json")
	params.Add("component", "format_topics")
	params.Add("itemtype", "sectionname")
	params.Add("itemid", fmt.Sprintf("%d", sectionID))
	params.Add("value", name)

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
		return fmt.Errorf("moodle API error editing section: %s", string(body))
	}

	return nil
}

// DeleteSection löscht eine Sektion aus dem Kurs über core_courseformat_update_course.
func (c *MoodleClient) DeleteSection(courseID int64, sectionID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_courseformat_update_course")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("action", "section_delete")
	params.Add("ids[0]", fmt.Sprintf("%d", sectionID))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php", c.Host)

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return fmt.Errorf("moodle API Fehler beim Löschen der Sektion: %s", string(body))
	}

	return nil
}
