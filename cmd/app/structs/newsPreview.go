package structs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type NewsPreview struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func NewNewsPreview(title string, link string) NewsPreview {
	return NewsPreview{
		Title: title,
		Link:  link,
	}
}

func (preview NewsPreview) Format() string {
	return fmt.Sprintf("Title: %s\n Link: %s", preview.Title, preview.Link)
}

func (preview NewsPreview) Json() ([]byte, error) {
	// Serialize the struct to JSON
	previewJson, err := json.Marshal(preview)
	if err != nil {
		return nil, err
	}
	return previewJson, nil
}

func (preview NewsPreview) GetHash() [32]byte {
	previewJson, _ := preview.Json()
	hash := sha256.Sum256(previewJson)

	return hash
}
