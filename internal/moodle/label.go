package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CreateLabel erstellt ein Label (Textfeld) in einem Kursabschnitt und gibt die
// Course Module ID (cmid) zurück. Mit previousElementId > 0 wird das Label
// direkt vor dem angegebenen Modul positioniert.
func (c *MoodleClient) CreateLabel(courseID int64, sectionNum int64, intro string, name string, previousElementId int64) (int64, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_create_label")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("sectionnum", fmt.Sprintf("%d", sectionNum))
	params.Add("intro", intro)
	params.Add("name", name)
	params.Add("previousElementId", fmt.Sprintf("%d", previousElementId))

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
		return 0, fmt.Errorf("moodle API error creating label: %s", string(body))
	}

	var result struct {
		CMID    int64 `json:"cmid"`
		Visible bool  `json:"visible"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("error parsing API response: %w\nBody: %s", err, string(body))
	}

	return result.CMID, nil
}

// UpdateLabel aktualisiert den Inhalt und optionale Positionierungsparameter
// eines bestehenden Labels.
func (c *MoodleClient) UpdateLabel(courseID int64, cmID int64, intro string, name string, sectionNum int64, previousElementId int64, visible int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_update_label")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))
	params.Add("intro", intro)
	params.Add("name", name)
	params.Add("section", fmt.Sprintf("%d", sectionNum))
	params.Add("previousElementId", fmt.Sprintf("%d", previousElementId))
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
		return fmt.Errorf("moodle API error updating label: %s", string(body))
	}

	return nil
}

// DeleteLabel löscht ein Label aus einem Kurs.
func (c *MoodleClient) DeleteLabel(courseID int64, cmID int64) error {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_course_api_delete_label")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

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
		return fmt.Errorf("moodle API error deleting label: %s", string(body))
	}

	return nil
}
