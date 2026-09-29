package domain

import "time"

const (
	HelpDocumentStatusDraft     = "draft"
	HelpDocumentStatusPublished = "published"
)

type HelpDocumentSnapshot struct {
	Title           string         `json:"title"`
	Category        string         `json:"category"`
	Summary         string         `json:"summary"`
	ContentMarkdown string         `json:"content_markdown"`
	SelectorSchema  map[string]any `json:"selector_schema"`
	Version         int            `json:"version"`
	PublishedAt     time.Time      `json:"published_at"`
}
