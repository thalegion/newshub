package messages

type NewNewsPreview struct {
	Uuid  string `json:"uuid"`
	Title string `json:"title"`
	Link  string `json:"link"`
}
