package onedrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents a OneDrive API client
type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

// Folder represents a OneDrive folder
type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// FileInfo represents a OneDrive file
type FileInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	CreatedTime  time.Time `json:"createdDateTime"`
	ModifiedTime time.Time `json:"lastModifiedDateTime"`
}

// DriveResponse represents the response from the OneDrive API
type DriveResponse struct {
	Value []interface{} `json:"value"`
}

// NewClient creates a new OneDrive client
func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		baseURL:     "https://graph.microsoft.com/v1.0",
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GetTopLevelFolders retrieves top-level folders from OneDrive
func (c *Client) GetTopLevelFolders() ([]Folder, error) {
	url := fmt.Sprintf("%s/me/drive/root/children", c.baseURL)
	
	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var driveResp DriveResponse
	if err := json.NewDecoder(resp.Body).Decode(&driveResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var folders []Folder
	for _, item := range driveResp.Value {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if it's a folder
		if folder, exists := itemMap["folder"]; exists && folder != nil {
			folder := Folder{
				ID:   itemMap["id"].(string),
				Name: itemMap["name"].(string),
			}
			if size, ok := itemMap["size"].(float64); ok {
				folder.Size = int64(size)
			}
			folders = append(folders, folder)
		}
	}

	return folders, nil
}

// GetFolderContents retrieves contents of a specific folder
func (c *Client) GetFolderContents(folderID string) ([]FileInfo, error) {
	url := fmt.Sprintf("%s/me/drive/items/%s/children", c.baseURL, folderID)
	
	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var driveResp DriveResponse
	if err := json.NewDecoder(resp.Body).Decode(&driveResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var files []FileInfo
	for _, item := range driveResp.Value {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if it's a file
		if _, exists := itemMap["file"]; exists {
			file := FileInfo{
				ID:   itemMap["id"].(string),
				Name: itemMap["name"].(string),
			}
			
			if size, ok := itemMap["size"].(float64); ok {
				file.Size = int64(size)
			}

			if createdTime, ok := itemMap["createdDateTime"].(string); ok {
				if t, err := time.Parse(time.RFC3339, createdTime); err == nil {
					file.CreatedTime = t
				}
			}

			if modifiedTime, ok := itemMap["lastModifiedDateTime"].(string); ok {
				if t, err := time.Parse(time.RFC3339, modifiedTime); err == nil {
					file.ModifiedTime = t
				}
			}

			files = append(files, file)
		}
	}

	return files, nil
}

// GetSubfolders retrieves subfolders from a specific folder
func (c *Client) GetSubfolders(folderID string) ([]Folder, error) {
	url := fmt.Sprintf("%s/me/drive/items/%s/children", c.baseURL, folderID)
	
	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var driveResp DriveResponse
	if err := json.NewDecoder(resp.Body).Decode(&driveResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var folders []Folder
	for _, item := range driveResp.Value {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if it's a folder
		if folder, exists := itemMap["folder"]; exists && folder != nil {
			folder := Folder{
				ID:   itemMap["id"].(string),
				Name: itemMap["name"].(string),
			}
			if size, ok := itemMap["size"].(float64); ok {
				folder.Size = int64(size)
			}
			folders = append(folders, folder)
		}
	}

	return folders, nil
}

// GetAllSnapshots retrieves all files from the snapshots folder
func (c *Client) GetAllSnapshots(folderID string) ([]FileInfo, error) {
	// Look for snapshots subfolder
	subfolders, err := c.GetSubfolders(folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subfolders for folder %s: %w", folderID, err)
	}

	// List available subfolders for debugging
	var folderNames []string
	for _, folder := range subfolders {
		folderNames = append(folderNames, folder.Name)
	}

	var snapshotsFolderID string
	for _, folder := range subfolders {
		if folder.Name == "snapshots" {
			snapshotsFolderID = folder.ID
			break
		}
	}

	if snapshotsFolderID == "" {
		return nil, fmt.Errorf("snapshots folder not found in folder %s. Available subfolders: %v", folderID, folderNames)
	}

	// Get all files in snapshots folder
	files, err := c.GetFolderContents(snapshotsFolderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot files from folder %s: %w", snapshotsFolderID, err)
	}

	return files, nil
}

// CheckTodayBackups checks if there are files created today in the snapshots folder
func (c *Client) CheckTodayBackups(folderID string) (bool, []FileInfo, error) {
	// Get all snapshot files
	allFiles, err := c.GetAllSnapshots(folderID)
	if err != nil {
		return false, nil, err
	}

	// Check if any files were created today
	today := time.Now().UTC().Truncate(24 * time.Hour)
	var todayFiles []FileInfo
	
	for _, file := range allFiles {
		fileDate := file.CreatedTime.UTC().Truncate(24 * time.Hour)
		if fileDate.Equal(today) {
			todayFiles = append(todayFiles, file)
		}
	}

	return len(todayFiles) > 0, todayFiles, nil
}

// makeRequest makes an HTTP request to the OneDrive API
func (c *Client) makeRequest(method, url string, body interface{}) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return resp, nil
} 