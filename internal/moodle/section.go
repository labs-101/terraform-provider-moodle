package moodle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Section represents a course section in Moodle.
// Visible is int64 because core_course_get_contents returns 0/1, not JSON booleans.
type Section struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Section int64  `json:"section"`
	Summary string `json:"summary"`
	Visible int64  `json:"visible"`
}

// CreateSection appends a new empty section to the course and returns the newly created section.
// A per-course mutex serializes concurrent calls so parallel Terraform creates never race on
// the "newest section" detection logic.
func (c *MoodleClient) CreateSection(courseID int64) (*Section, error) {
	mu := c.courseSectionLock(courseID)
	mu.Lock()
	defer mu.Unlock()

	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_courseformat_update_course")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("action", "section_add")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("CreateSection course=%d: %w", courseID, err)
	}

	// The Moodle API returns the response double-encoded: a JSON string that itself contains a JSON array.
	var rawJSON string
	if err := json.Unmarshal(body, &rawJSON); err != nil {
		rawJSON = string(body)
	}

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
		return nil, fmt.Errorf("parsing CreateSection response: %w — body: %s", err, string(body))
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
		visibleInt := int64(0)
		if u.Fields.Visible {
			visibleInt = 1
		}
		s := &Section{
			ID:      id,
			Name:    u.Fields.Title,
			Section: u.Fields.Section,
			Visible: visibleInt,
		}
		if newest == nil || s.Section > newest.Section {
			newest = s
		}
	}

	if newest == nil {
		return nil, fmt.Errorf("CreateSection: no new section found in API response for course %d", courseID)
	}

	return newest, nil
}

// GetCourseSections returns all sections of a course.
func (c *MoodleClient) GetCourseSections(courseID int64) ([]Section, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_sections")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetCourseSections course=%d: %w", courseID, err)
	}

	var raw []struct {
		ID      int64  `json:"id"`
		Section int64  `json:"section"`
		Name    string `json:"name"`
		Summary string `json:"summary"`
		Visible int64  `json:"visible"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing sections: %w — body: %s", err, string(body))
	}

	sections := make([]Section, len(raw))
	for i, r := range raw {
		sections[i] = Section{
			ID:      r.ID,
			Section: r.Section,
			Name:    r.Name,
			Summary: r.Summary,
			Visible: r.Visible,
		}
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

// EditSection updates the name, summary, and visibility of a section.
func (c *MoodleClient) EditSection(sectionID int64, name string, summary string, visible bool) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_update_inplace_editable")
	params.Add("moodlewsrestformat", "json")
	params.Add("component", "format_topics")
	params.Add("itemtype", "sectionname")
	params.Add("itemid", fmt.Sprintf("%d", sectionID))
	params.Add("value", name)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("EditSection %d: %w", sectionID, err)
	}

	return nil
}

// DeleteSection deletes a section from a course.
func (c *MoodleClient) DeleteSection(courseID int64, sectionID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "core_courseformat_update_course")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("action", "section_delete")
	params.Add("ids[0]", fmt.Sprintf("%d", sectionID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webservice/rest/server.php", c.Host), strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if _, err := c.doRequest(req); err != nil {
		return fmt.Errorf("DeleteSection %d: %w", sectionID, err)
	}

	return nil
}
