package moodle

import (
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
		return 0, "", fmt.Errorf("opening file %q: %w", filePath, err)
	}
	defer f.Close()

	filename := filepath.Base(filePath)

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		part, err := mw.CreateFormFile("file", filename)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, f); err != nil {
			pw.CloseWithError(err)
			return
		}
		mw.Close()
	}()

	uploadURL := fmt.Sprintf("%s/webservice/upload.php?token=%s&moodlewsrestformat=json", c.Host, c.Token)

	req, err := http.NewRequest("POST", uploadURL, pr)
	if err != nil {
		return 0, "", fmt.Errorf("creating upload request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	body, err := c.doRequest(req)
	if err != nil {
		return 0, "", fmt.Errorf("UploadFile %q: %w", filename, err)
	}

	var uploads []uploadFileResponse
	if err := json.Unmarshal(body, &uploads); err != nil {
		return 0, "", fmt.Errorf("parsing upload response: %w — body: %s", err, string(body))
	}

	if len(uploads) == 0 {
		return 0, "", fmt.Errorf("UploadFile: moodle returned no upload response")
	}

	return uploads[0].ItemID, uploads[0].Filename, nil
}

func (c *MoodleClient) AddFileToSection(courseID int64, sectionNum int64, itemID int64, displayName string, visible bool) (int64, error) {
	visibleInt := 0
	if visible {
		visibleInt = 1
	}

	params := url.Values{}
	params.Add("itemid", fmt.Sprintf("%d", itemID))
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("sectionnum", fmt.Sprintf("%d", sectionNum))
	params.Add("displayname", displayName)
	params.Add("visible", fmt.Sprintf("%d", visibleInt))

	reqURL := fmt.Sprintf("%s/webservice/rest/server.php?wstoken=%s&wsfunction=local_course_add_new_course_module_resource&moodlewsrestformat=json",
		c.Host, c.Token)

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return 0, fmt.Errorf("AddFileToSection: %w", err)
	}

	var result struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return -1, fmt.Errorf("parsing response: %w — body: %s", err, string(body))
	}

	id, err := strconv.ParseInt(result.ID, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("parsing ID as int64: %w", err)
	}

	return id, nil
}
