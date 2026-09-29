package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrHelpDocumentNotFound = infraerrors.NotFound("HELP_DOCUMENT_NOT_FOUND", "help document not found")
	ErrHelpDocumentConflict = infraerrors.Conflict("HELP_DOCUMENT_CONFLICT", "document changed; reload before saving")
	ErrHelpDocumentInvalid  = infraerrors.BadRequest("HELP_DOCUMENT_INVALID", "invalid help document")
)

type HelpDocument struct {
	ID                 int64                         `json:"id"`
	Slug               string                        `json:"slug"`
	Title              string                        `json:"title"`
	Category           string                        `json:"category"`
	Summary            string                        `json:"summary"`
	ContentMarkdown    string                        `json:"content_markdown"`
	Status             string                        `json:"status"`
	SelectorSchema     map[string]any                `json:"selector_schema"`
	Version            int                           `json:"version"`
	PublishedAt        *time.Time                    `json:"published_at,omitempty"`
	UpdatedBy          *int64                        `json:"updated_by,omitempty"`
	CreatedAt          time.Time                     `json:"created_at"`
	UpdatedAt          time.Time                     `json:"updated_at"`
	PublishedSnapshot  *domain.HelpDocumentSnapshot  `json:"-"`
	PublicationHistory []domain.HelpDocumentSnapshot `json:"-"`
	CanRollback        bool                          `json:"can_rollback"`
}

type HelpDocumentRepository interface {
	Create(context.Context, *HelpDocument) error
	GetByID(context.Context, int64) (*HelpDocument, error)
	GetBySlug(context.Context, string, bool) (*HelpDocument, error)
	List(context.Context, pagination.PaginationParams, string, string) ([]HelpDocument, *pagination.PaginationResult, error)
	Mutate(context.Context, int64, func(*HelpDocument) error) (*HelpDocument, error)
	Delete(context.Context, int64) error
}

type HelpDocumentService struct{ repo HelpDocumentRepository }

func NewHelpDocumentService(repo HelpDocumentRepository) *HelpDocumentService {
	return &HelpDocumentService{repo: repo}
}
func (s *HelpDocumentService) GetByID(ctx context.Context, id int64) (*HelpDocument, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *HelpDocumentService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
func (s *HelpDocumentService) GetPublished(ctx context.Context, slug string) (*HelpDocument, error) {
	d, err := s.repo.GetBySlug(ctx, strings.Trim(slug, "/"), true)
	if err != nil {
		return nil, err
	}
	return publishedHelpDocument(d), nil
}
func publishedHelpDocument(d *HelpDocument) *HelpDocument {
	p := d.PublishedSnapshot
	if p == nil {
		return nil
	}
	return &HelpDocument{ID: d.ID, Slug: d.Slug, Title: p.Title, Category: p.Category, Summary: p.Summary, ContentMarkdown: p.ContentMarkdown, SelectorSchema: p.SelectorSchema, Version: p.Version, Status: domain.HelpDocumentStatusPublished, PublishedAt: &p.PublishedAt, CreatedAt: d.CreatedAt, UpdatedAt: p.PublishedAt}
}
func (s *HelpDocumentService) ListPublished(ctx context.Context, category string) ([]HelpDocument, error) {
	out := []HelpDocument{}
	for page := 1; ; page++ {
		items, result, err := s.repo.List(ctx, pagination.PaginationParams{Page: page, PageSize: 200}, "", domain.HelpDocumentStatusPublished)
		if err != nil {
			return nil, err
		}
		for i := range items {
			if d := publishedHelpDocument(&items[i]); d != nil && (category == "" || d.Category == category) {
				out = append(out, *d)
			}
		}
		if page >= result.Pages {
			break
		}
	}
	return out, nil
}
func (s *HelpDocumentService) List(ctx context.Context, p pagination.PaginationParams, category, status string) ([]HelpDocument, *pagination.PaginationResult, error) {
	return s.repo.List(ctx, p, category, status)
}

var helpSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:[-/][a-z0-9]+)*$`)

func ValidateHelpDocument(d *HelpDocument) error {
	d.Slug = strings.Trim(strings.TrimSpace(d.Slug), "/")
	d.Title = strings.TrimSpace(d.Title)
	if !helpSlugPattern.MatchString(d.Slug) || len(d.Slug) > 160 || d.Title == "" || len([]rune(d.Title)) > 240 || len([]rune(d.Summary)) > 1000 || strings.TrimSpace(d.ContentMarkdown) == "" || len(d.ContentMarkdown) > 200000 {
		return ErrHelpDocumentInvalid
	}
	switch d.Category {
	case "quick-start", "claude-code", "codex", "opencode", "gemini-cli", "grok-cli", "cherry-studio", "cursor", "cline", "roo-code", "dsh", "pi", "openclaw", "hermes", "workbuddy", "zcode", "trae", "read-frog", "kiss-translator", "immersive-translate", "sillytavern", "tavernai", "api", "faq":
	default:
		return ErrHelpDocumentInvalid
	}
	if d.SelectorSchema == nil {
		d.SelectorSchema = map[string]any{}
	}
	raw, err := json.Marshal(d.SelectorSchema)
	if err != nil || len(raw) > 16384 {
		return ErrHelpDocumentInvalid
	}
	allowed := map[string]bool{"platforms": true, "clients": true, "systems": true, "shells": true, "protocols": true, "versions": true}
	for key, value := range d.SelectorSchema {
		if !allowed[key] {
			return ErrHelpDocumentInvalid
		}
		encoded, _ := json.Marshal(value)
		var values []string
		if json.Unmarshal(encoded, &values) != nil || len(values) > 40 {
			return ErrHelpDocumentInvalid
		}
		for _, v := range values {
			if len(v) > 160 {
				return ErrHelpDocumentInvalid
			}
		}
	}
	return nil
}
func (s *HelpDocumentService) Preview(d *HelpDocument) error { return ValidateHelpDocument(d) }
func (s *HelpDocumentService) Save(ctx context.Context, d *HelpDocument) error {
	if err := ValidateHelpDocument(d); err != nil {
		return err
	}
	if d.ID == 0 {
		d.Status = domain.HelpDocumentStatusDraft
		d.Version = 1
		d.PublishedSnapshot = nil
		d.PublicationHistory = nil
		d.PublishedAt = nil
		return s.repo.Create(ctx, d)
	}
	saved, err := s.repo.Mutate(ctx, d.ID, func(current *HelpDocument) error {
		if current.Version != d.Version {
			return ErrHelpDocumentConflict
		}
		// Stable slugs keep published links independent from draft edits.
		if current.Slug != d.Slug {
			return ErrHelpDocumentInvalid
		}
		current.Title = d.Title
		current.Category = d.Category
		current.Summary = d.Summary
		current.ContentMarkdown = d.ContentMarkdown
		current.SelectorSchema = d.SelectorSchema
		current.UpdatedBy = d.UpdatedBy
		current.Version++
		return nil
	})
	if err == nil {
		*d = *saved
	}
	return err
}
func (s *HelpDocumentService) Transition(ctx context.Context, id int64, action string, version int, actor int64) (*HelpDocument, error) {
	return s.repo.Mutate(ctx, id, func(d *HelpDocument) error {
		if d.Version != version {
			return ErrHelpDocumentConflict
		}
		now := time.Now().UTC()
		switch action {
		case "publish":
			if err := ValidateHelpDocument(d); err != nil {
				return err
			}
			if d.PublishedSnapshot != nil {
				d.PublicationHistory = append(d.PublicationHistory, *d.PublishedSnapshot)
			}
			d.PublishedSnapshot = &domain.HelpDocumentSnapshot{Title: d.Title, Category: d.Category, Summary: d.Summary, ContentMarkdown: d.ContentMarkdown, SelectorSchema: d.SelectorSchema, Version: d.Version + 1, PublishedAt: now}
			d.Status = domain.HelpDocumentStatusPublished
			d.PublishedAt = &now
		case "unpublish":
			d.Status = domain.HelpDocumentStatusDraft
			d.PublishedAt = nil
		case "rollback":
			if len(d.PublicationHistory) == 0 {
				return infraerrors.BadRequest("HELP_DOCUMENT_NO_HISTORY", "no previous publication")
			}
			previous := d.PublicationHistory[len(d.PublicationHistory)-1]
			if d.PublishedSnapshot != nil {
				d.PublicationHistory = append(d.PublicationHistory, *d.PublishedSnapshot)
			}
			previous.Version = d.Version + 1
			previous.PublishedAt = now
			d.PublishedSnapshot = &previous
			d.Status = domain.HelpDocumentStatusPublished
			d.PublishedAt = &now
		default:
			return ErrHelpDocumentInvalid
		}
		d.Version++
		d.UpdatedBy = &actor
		return nil
	})
}
