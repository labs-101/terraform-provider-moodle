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
		return nil, fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return nil, fmt.Errorf("moodle API Fehler beim Hinzufügen der Sektion: %s", string(body))
	}

	// Die Moodle-API liefert die Antwort doppelt kodiert: ein JSON-String, der selbst ein JSON-Array enthält.
	// Erster Schritt: den äußeren String entpacken.
	var rawJSON string
	if err := json.Unmarshal(body, &rawJSON); err != nil {
		// Fallback: vielleicht ist es doch direkt ein Array
		rawJSON = string(body)
	}

	// Zweiter Schritt: das eigentliche Array parsen.
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
		return nil, fmt.Errorf("fehler beim Parsen der API-Antwort: %w\nBody: %s", err, string(body))
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
		return nil, fmt.Errorf("keine neue Sektion in der API-Antwort des Kurses %d gefunden", courseID)
	}

	return newest, nil
}

// GetCourseSections gibt alle Sektionen eines Kurses zurück (core_course_get_contents).
func (c *MoodleClient) GetCourseSections(courseID int64) ([]Section, error) {
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
		return nil, fmt.Errorf("moodle API Fehler beim Lesen der Sektionen: %s", string(body))
	}

	var sections []Section
	if err := json.Unmarshal(body, &sections); err != nil {
		return nil, fmt.Errorf("fehler beim Parsen der Sektionen: %w\nBody: %s", err, string(body))
	}

	return sections, nil
}

// GetSection gibt eine bestimmte Sektion anhand ihrer Datenbank-ID zurück.
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

	return nil, fmt.Errorf("sektion mit ID %d wurde im Kurs %d nicht gefunden", sectionID, courseID)
}

// EditSection aktualisiert Name, Zusammenfassung und Sichtbarkeit einer Sektion
// über die Moodle-Funktion core_course_edit_section.
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
		return fmt.Errorf("moodle API Fehler beim Bearbeiten der Sektion: %s", string(body))
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
