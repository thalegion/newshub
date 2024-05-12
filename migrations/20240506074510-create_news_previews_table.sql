-- +migrate Up
CREATE TABLE news_previews
(
    uuid       UUID PRIMARY KEY,
    title      VARCHAR(255) NOT NULL,
    link       VARCHAR(255) NOT NULL,
    content    TEXT,
    status     VARCHAR(20) NOT NULL ,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_news_previews_title ON news_previews (title);

-- +migrate Down
DROP TABLE news_previews;