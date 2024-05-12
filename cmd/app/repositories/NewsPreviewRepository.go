package repositories

import (
	"database/sql"
	"log"
	"newshub/cmd/app/entities"
	"newshub/cmd/app/enums"
	"newshub/cmd/app/models"
	"time"
)

type NewsPreviewRepository struct {
	db *sql.DB
}

func NewNewsPreviewRepository(db *sql.DB) *NewsPreviewRepository {
	return &NewsPreviewRepository{db: db}
}

func (repository *NewsPreviewRepository) FindByTitle(title string) *entities.NewsPreview {
	query := "SELECT uuid, title, link, content, status, created_at, updated_at FROM news_previews WHERE title = $1;"

	return repository.persistFromQuery(query, title)
}

func (repository *NewsPreviewRepository) FindByUuid(uuid string) *entities.NewsPreview {
	query := "SELECT uuid, title, link, content, status, created_at, updated_at FROM news_previews WHERE uuid = $1;"

	return repository.persistFromQuery(query, uuid)
}

func (repository *NewsPreviewRepository) Add(newsPreview *entities.NewsPreview) {
	query := `
INSERT INTO news_previews (uuid, title, link, content, status, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);`

	_, err := repository.db.Exec(query, newsPreview.Uuid, newsPreview.Title, newsPreview.Link, nil, newsPreview.Status, time.Now())
	if err != nil {
		log.Fatalf("Error inserting news preview: %v", err)
	}
}

func (repository *NewsPreviewRepository) Update(newsPreview *entities.NewsPreview) {
	query := `
UPDATE news_previews SET title = $2, link = $3, content = $4, status = $5, updated_at = $6
WHERE uuid = $1;`

	_, err := repository.db.Exec(query, newsPreview.Uuid, newsPreview.Title, newsPreview.Link, newsPreview.Content, newsPreview.Status, time.Now())
	if err != nil {
		log.Fatalf("Error updating news preview: %v", err)
	}
}

func (repository *NewsPreviewRepository) persistFromQuery(query string, params ...any) *entities.NewsPreview {
	rows, err := repository.db.Query(query, params...)
	if err != nil {
		log.Fatalf("Issue during persist on query: %v", err)
	}
	defer rows.Close()

	var newsPreview models.NewsPreview
	for rows.Next() {
		err := rows.Scan(
			&newsPreview.Uuid,
			&newsPreview.Title,
			&newsPreview.Link,
			&newsPreview.Content,
			&newsPreview.Status,
			&newsPreview.CreatedAt,
			&newsPreview.UpdatedAt,
		)
		if err != nil {
			log.Fatalf("Issue during findByTitle row scan: %v", err)
		}
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Issue during findByTitle after row scan: %v", err)
	}

	if newsPreview == (models.NewsPreview{}) {
		return nil
	}

	return entities.NewsPreviewFromStorage(newsPreview.Uuid, newsPreview.Title, newsPreview.Link, newsPreview.Content, enums.NewsPreviewStatus(newsPreview.Status))
}
