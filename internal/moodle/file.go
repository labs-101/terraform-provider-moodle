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
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ModType  string `json:"modname"`
	Visible  bool   `json:"-"`
	FileSize int64  `json:"-"` // total size in bytes of the module's files (0 if it has none)
}

type uploadFileResponse struct {
	ItemID   int64  `json:"itemid"`
	Filename string `json:"filename"`
}

// ResourceFile holds the details of a single-file resource activity, including
// the SHA-1 content hash Moodle stores for the underlying file.
type ResourceFile struct {
	CMID        int64
	Name        string
	Visible     bool
	FileName    string
	FileSize    int64
	ContentHash string
}

// GetResourceFile returns the stored file of a resource activity. It returns
// (nil, nil) when the activity no longer exists, so callers can drop it from state.
func (c *MoodleClient) GetResourceFile(courseID, cmID int64) (*ResourceFile, error) {
	params := url.Values{}
	params.Add("wstoken", c.Token)
	params.Add("wsfunction", "local_courseapi_get_resource_file")
	params.Add("moodlewsrestformat", "json")
	params.Add("courseid", fmt.Sprintf("%d", courseID))
	params.Add("cmid", fmt.Sprintf("%d", cmID))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/webservice/rest/server.php?%s", c.Host, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("GetResourceFile course=%d cm=%d: %w", courseID, cmID, err)
	}

	var raw struct {
		Found       bool   `json:"found"`
		CMID        int64  `json:"cmid"`
		Name        string `json:"name"`
		Visible     int64  `json:"visible"`
		FileName    string `json:"filename"`
		FileSize    int64  `json:"filesize"`
		ContentHash string `json:"contenthash"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing resource file: %w — body: %s", err, string(body))
	}

	if !raw.Found {
		return nil, nil
	}

	return &ResourceFile{
		CMID:        raw.CMID,
		Name:        raw.Name,
		Visible:     raw.Visible == 1,
		FileName:    raw.FileName,
		FileSize:    raw.FileSize,
		ContentHash: raw.ContentHash,
	}, nil
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
