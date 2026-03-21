package moodle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CourseModule struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	ModType string `json:"modname"`
}

type uploadFileResponse struct {
	ItemID   int64  `json:"itemid"`
	Filename string `json:"filename"`
}

func (c *MoodleClient) UploadFile(filePath string) (int64, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, "", fmt.Errorf("error opening file %q: %w", filePath, err)
	}
	defer f.Close()

	filename := filepath.Base(filePath)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return 0, "", fmt.Errorf("error creating multipart field: %w", err)
	}
	if _, err = io.Copy(part, f); err != nil {
		return 0, "", fmt.Errorf("error reading file: %w", err)
	}
	mw.Close()

	uploadURL := fmt.Sprintf("%s/webservice/upload.php?token=%s&moodlewsrestformat=json", c.Host, c.Token)

	req, err := http.NewRequest("POST", uploadURL, &buf)
	if err != nil {
		return 0, "", fmt.Errorf("error creating upload request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("error sending upload request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, "", fmt.Errorf("error reading upload response: %w", err)
	}

	if strings.Contains(string(body), "exception") {
		return 0, "", fmt.Errorf("moodle upload error: %s", string(body))
	}

	var uploads []uploadFileResponse
	if err := json.Unmarshal(body, &uploads); err != nil {
		return 0, "", fmt.Errorf("error parsing upload response: %w\nBody: %s", err, string(body))
	}

	if len(uploads) == 0 {
		return 0, "", fmt.Errorf("moodle returned no upload response")
	}

	return uploads[0].ItemID, uploads[0].Filename, nil
}

func (c *MoodleClient) AddFileToSection(courseID int64, sectionNum int64, itemID int64, displayName string, visible int64) (int64, error) {
	params := url.Values{}
	params.Add("itemid", fmt.Sprintf("%d", itemID))
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("sectionnum", fmt.Sprintf("%d", sectionNum))
	params.Add("displayname", displayName)
	params.Add("visible", fmt.Sprintf("%d", visible))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?wstoken=%s&wsfunction=local_course_add_new_course_module_resource&moodlewsrestformat=json",
		c.Host, c.Token)

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
		return 0, fmt.Errorf("moodle API error adding file: %s", string(body))
	}

	var result struct {
		Message string `json:"message"`
		Id      string `json:"id"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return -1, fmt.Errorf("error parsing course contents: %w\nBody: %s", err, string(body))
	}

	id, err := strconv.ParseInt(result.Id, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("could not parse ID as int64: %w", err)
	}

	return id, nil
}
