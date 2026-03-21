package moodle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AddChoiceToSection erstellt eine Choice-Aktivität in einem Kursabschnitt
// und gibt die Course Module ID (cmid) zurück.
func (c *MoodleClient) AddChoiceToSection(courseID int64, sectionNum int64, name string, intro string, options []string, allowMultiple bool) (int64, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_add_choice_to_section")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("sectionnum", fmt.Sprintf("%d", sectionNum))
	params.Add("name", name)
	params.Add("intro", intro)
	if allowMultiple {
		params.Add("allowmultiple", "1")
	} else {
		params.Add("allowmultiple", "0")
	}
	for i, opt := range options {
		params.Add(fmt.Sprintf("options[%d]", i), opt)
	}

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
		return 0, fmt.Errorf("moodle API error creating choice: %s", string(body))
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
