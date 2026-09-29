package service

import (
	"context"
	"encoding/json"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
	"testing"
)

type memoryHelpRepo struct{ document *HelpDocument }

func cloneHelp(d *HelpDocument) *HelpDocument {
	if d == nil {
		return nil
	}
	out := *d
	if d.PublishedSnapshot != nil {
		p := *d.PublishedSnapshot
		out.PublishedSnapshot = &p
	}
	out.PublicationHistory = append(out.PublicationHistory[:0:0], d.PublicationHistory...)
	return &out
}
func (r *memoryHelpRepo) Create(_ context.Context, d *HelpDocument) error {
	d.ID = 1
	r.document = cloneHelp(d)
	return nil
}
func (r *memoryHelpRepo) GetByID(_ context.Context, _ int64) (*HelpDocument, error) {
	if r.document == nil {
		return nil, ErrHelpDocumentNotFound
	}
	return cloneHelp(r.document), nil
}
func (r *memoryHelpRepo) GetBySlug(ctx context.Context, slug string, published bool) (*HelpDocument, error) {
	d, err := r.GetByID(ctx, 1)
	if err != nil {
		return nil, err
	}
	if d.Slug != slug || (published && d.Status != "published") {
		return nil, ErrHelpDocumentNotFound
	}
	return d, nil
}
func (r *memoryHelpRepo) List(_ context.Context, p pagination.PaginationParams, category, status string) ([]HelpDocument, *pagination.PaginationResult, error) {
	out := []HelpDocument{}
	if d := r.document; d != nil && (status == "" || d.Status == status) {
		out = append(out, *cloneHelp(d))
	}
	return out, &pagination.PaginationResult{Total: int64(len(out)), Pages: 1}, nil
}
func (r *memoryHelpRepo) Mutate(_ context.Context, _ int64, f func(*HelpDocument) error) (*HelpDocument, error) {
	d := cloneHelp(r.document)
	if err := f(d); err != nil {
		return nil, err
	}
	r.document = cloneHelp(d)
	return d, nil
}
func (r *memoryHelpRepo) Delete(_ context.Context, _ int64) error { r.document = nil; return nil }

func TestHelpDocumentPublicationLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := &memoryHelpRepo{}
	s := NewHelpDocumentService(repo)
	d := &HelpDocument{Slug: "clients/codex", Title: "Version one", Category: "codex", ContentMarkdown: "published one", Status: "published"}
	require.NoError(t, s.Save(ctx, d))
	require.Equal(t, "draft", d.Status)
	_, err := s.GetPublished(ctx, d.Slug)
	require.ErrorIs(t, err, ErrHelpDocumentNotFound)
	d, err = s.Transition(ctx, d.ID, "publish", d.Version, 7)
	require.NoError(t, err)
	d.Title = "Draft two"
	d.Category = "faq"
	d.ContentMarkdown = "private draft"
	require.NoError(t, s.Save(ctx, d))
	live, err := s.GetPublished(ctx, d.Slug)
	require.NoError(t, err)
	require.Equal(t, "Version one", live.Title)
	list, err := s.ListPublished(ctx, "codex")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "published one", list[0].ContentMarkdown)
	encoded, err := json.Marshal(list)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "private draft")
	require.Contains(t, string(encoded), "content_markdown")
	d, err = s.Transition(ctx, d.ID, "publish", d.Version, 7)
	require.NoError(t, err)
	live, err = s.GetPublished(ctx, d.Slug)
	require.NoError(t, err)
	require.Equal(t, "Draft two", live.Title)
	d, err = s.Transition(ctx, d.ID, "rollback", d.Version, 7)
	require.NoError(t, err)
	live, err = s.GetPublished(ctx, d.Slug)
	require.NoError(t, err)
	require.Equal(t, "Version one", live.Title)
	require.Len(t, repo.document.PublicationHistory, 2)
	// A rollback changes the publication, without discarding the editor's draft.
	require.Equal(t, "Draft two", d.Title)
	d, err = s.Transition(ctx, d.ID, "unpublish", d.Version, 7)
	require.NoError(t, err)
	require.Nil(t, d.PublishedAt)
	require.Nil(t, repo.document.PublishedAt)
	_, err = s.GetPublished(ctx, d.Slug)
	require.ErrorIs(t, err, ErrHelpDocumentNotFound)
	_, err = s.Transition(ctx, d.ID, "publish", d.Version-1, 7)
	require.ErrorIs(t, err, ErrHelpDocumentConflict)
}

func TestHelpDocumentSeededTavernCategoriesCanBeEdited(t *testing.T) {
	for _, category := range []string{"sillytavern", "tavernai"} {
		require.NoError(t, ValidateHelpDocument(&HelpDocument{Slug: "clients/" + category, Title: "Guide", Category: category, ContentMarkdown: "Guide content"}))
	}
}

func TestHelpDocumentValidationAndStaleSave(t *testing.T) {
	ctx := context.Background()
	s := NewHelpDocumentService(&memoryHelpRepo{})
	d := &HelpDocument{Slug: "clients/claude-code", Title: "Claude", Category: "claude-code", ContentMarkdown: "hello"}
	require.NoError(t, s.Save(ctx, d))
	stale := *d
	d.ContentMarkdown = "new draft"
	require.NoError(t, s.Save(ctx, d))
	require.ErrorIs(t, s.Save(ctx, &stale), ErrHelpDocumentConflict)
	d.Slug = "other"
	require.ErrorIs(t, s.Save(ctx, d), ErrHelpDocumentInvalid)
	d.Slug = "clients/claude-code"
	d.SelectorSchema = map[string]any{"api_key": "secret"}
	require.ErrorIs(t, s.Save(ctx, d), ErrHelpDocumentInvalid)
	d.SelectorSchema = map[string]any{"clients": []string{"claude-code"}}
	require.NoError(t, s.Preview(d))
	d.SelectorSchema = map[string]any{"clients": "claude-code"}
	require.ErrorIs(t, s.Preview(d), ErrHelpDocumentInvalid)
}

func TestHelpDesktopClientCategories(t *testing.T) {
	for _, category := range []string{"cherry-studio", "cursor", "cline", "roo-code"} {
		t.Run(category, func(t *testing.T) {
			d := &HelpDocument{Slug: "clients/" + category, Title: category, Category: category, ContentMarkdown: "## Configuration\nUse the default endpoint"}
			require.NoError(t, ValidateHelpDocument(d))
		})
	}
}
