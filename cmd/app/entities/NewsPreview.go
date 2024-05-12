package entities

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"newshub/cmd/app/enums"
)

type NewsPreview struct {
	Uuid    uuid.UUID `json:"uuid"`
	Title   string    `json:"title"`
	Link    string    `json:"link"`
	Content *string   `json:"content"`
	Status  enums.NewsPreviewStatus
}

func NewNewsPreview(uuid uuid.UUID, title string, link string) *NewsPreview {
	return &NewsPreview{
		Uuid:    uuid,
		Title:   title,
		Link:    link,
		Content: nil,
		Status:  enums.NewsPreviewStatusNew,
	}
}

func NewsPreviewFromStorage(uuid uuid.UUID, title string, link string, content *string, status enums.NewsPreviewStatus) *NewsPreview {
	return &NewsPreview{
		Uuid:    uuid,
		Title:   title,
		Link:    link,
		Content: content,
		Status:  status,
	}
}

func (preview *NewsPreview) SetContent(content string) {
	if preview.Status != enums.NewsPreviewStatusNew {
		log.Fatalf("Set content allowed only for new status. Current status %v", preview.Status)
	}

	preview.Content = &content
	preview.Status = enums.NewsPreviewStatusFullContent
}

func (preview NewsPreview) Format() string {
	return fmt.Sprintf("Title: %s\n Link: %s", preview.Title, preview.Link)
}

func (preview NewsPreview) Json() []byte {
	// Serialize the struct to JSON
	previewJson, err := json.Marshal(preview)
	if err != nil {
		log.Fatalf("Issue during marshalling json: %s", err)
	}
	return previewJson
}
