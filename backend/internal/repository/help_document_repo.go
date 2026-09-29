package repository

import (
	"context"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/helpdocument"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type helpDocumentRepository struct{ client *dbent.Client }

func NewHelpDocumentRepository(client *dbent.Client) service.HelpDocumentRepository {
	return &helpDocumentRepository{client: client}
}
func helpDocumentError(err error) error {
	if dbent.IsNotFound(err) {
		return service.ErrHelpDocumentNotFound
	}
	if dbent.IsConstraintError(err) {
		return service.ErrHelpDocumentConflict
	}
	return err
}
func toHelpDocument(e *dbent.HelpDocument) *service.HelpDocument {
	return &service.HelpDocument{ID: e.ID, Slug: e.Slug, Title: e.Title, Category: e.Category, Summary: e.Summary, ContentMarkdown: e.ContentMarkdown, SelectorSchema: e.SelectorSchema, Status: e.Status, Version: e.Version, PublishedAt: e.PublishedAt, UpdatedBy: e.UpdatedBy, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt, PublishedSnapshot: e.PublishedSnapshot, PublicationHistory: e.PublicationHistory, CanRollback: len(e.PublicationHistory) > 0}
}
func (r *helpDocumentRepository) Create(ctx context.Context, d *service.HelpDocument) error {
	e, err := r.client.HelpDocument.Create().SetSlug(d.Slug).SetTitle(d.Title).SetCategory(d.Category).SetSummary(d.Summary).SetContentMarkdown(d.ContentMarkdown).SetSelectorSchema(d.SelectorSchema).SetVersion(1).SetStatus(domain.HelpDocumentStatusDraft).SetNillableUpdatedBy(d.UpdatedBy).Save(ctx)
	if err != nil {
		return helpDocumentError(err)
	}
	*d = *toHelpDocument(e)
	return nil
}
func (r *helpDocumentRepository) GetByID(ctx context.Context, id int64) (*service.HelpDocument, error) {
	e, err := r.client.HelpDocument.Get(ctx, id)
	if err != nil {
		return nil, helpDocumentError(err)
	}
	return toHelpDocument(e), nil
}
func (r *helpDocumentRepository) GetBySlug(ctx context.Context, slug string, published bool) (*service.HelpDocument, error) {
	q := r.client.HelpDocument.Query().Where(helpdocument.SlugEQ(slug))
	if published {
		q = q.Where(helpdocument.StatusEQ(domain.HelpDocumentStatusPublished), helpdocument.PublishedSnapshotNotNil())
	}
	e, err := q.Only(ctx)
	if err != nil {
		return nil, helpDocumentError(err)
	}
	return toHelpDocument(e), nil
}
func (r *helpDocumentRepository) List(ctx context.Context, p pagination.PaginationParams, category, status string) ([]service.HelpDocument, *pagination.PaginationResult, error) {
	q := r.client.HelpDocument.Query()
	if category != "" {
		q = q.Where(helpdocument.CategoryEQ(category))
	}
	if status != "" {
		q = q.Where(helpdocument.StatusEQ(status))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	es, err := q.Order(dbent.Asc(helpdocument.FieldID)).Offset(p.Offset()).Limit(p.Limit()).All(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]service.HelpDocument, 0, len(es))
	for _, e := range es {
		out = append(out, *toHelpDocument(e))
	}
	return out, &pagination.PaginationResult{Total: int64(total), Page: p.Page, PageSize: p.Limit(), Pages: (total + p.Limit() - 1) / p.Limit()}, nil
}
func (r *helpDocumentRepository) Mutate(ctx context.Context, id int64, change func(*service.HelpDocument) error) (*service.HelpDocument, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	e, err := tx.HelpDocument.Query().Where(helpdocument.IDEQ(id)).ForUpdate().Only(ctx)
	if err != nil {
		return nil, helpDocumentError(err)
	}
	d := toHelpDocument(e)
	if err = change(d); err != nil {
		return nil, err
	}
	b := tx.HelpDocument.UpdateOneID(id).SetTitle(d.Title).SetCategory(d.Category).SetSummary(d.Summary).SetContentMarkdown(d.ContentMarkdown).SetSelectorSchema(d.SelectorSchema).SetStatus(d.Status).SetVersion(d.Version).SetNillableUpdatedBy(d.UpdatedBy)
	if d.PublishedSnapshot != nil {
		b.SetPublishedSnapshot(d.PublishedSnapshot)
	}
	if d.PublicationHistory != nil {
		b.SetPublicationHistory(d.PublicationHistory)
	}
	if d.PublishedAt != nil {
		b.SetPublishedAt(*d.PublishedAt)
	} else {
		b.ClearPublishedAt()
	}
	e, err = b.Save(ctx)
	if err != nil {
		return nil, helpDocumentError(err)
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return toHelpDocument(e), nil
}
func (r *helpDocumentRepository) Delete(ctx context.Context, id int64) error {
	return helpDocumentError(r.client.HelpDocument.DeleteOneID(id).Exec(ctx))
}
