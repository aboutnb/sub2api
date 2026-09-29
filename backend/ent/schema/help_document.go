package schema

import (
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// HelpDocument stores versioned, operator-maintained help articles.
type HelpDocument struct{ ent.Schema }

func (HelpDocument) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "help_documents"}}
}

func (HelpDocument) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").MaxLen(160).NotEmpty().Unique(),
		field.String("title").MaxLen(240).NotEmpty(),
		field.String("category").MaxLen(60).NotEmpty(),
		field.String("summary").MaxLen(1000).Optional(),
		field.String("content_markdown").SchemaType(map[string]string{dialect.Postgres: "text"}).NotEmpty(),
		field.JSON("selector_schema", map[string]any{}).Optional(),
		field.JSON("published_snapshot", &domain.HelpDocumentSnapshot{}).Optional(),
		field.JSON("publication_history", []domain.HelpDocumentSnapshot{}).Optional(),
		field.String("status").MaxLen(20).Default("draft"),
		field.Int("version").Default(1),
		field.Time("published_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int64("updated_by").Optional().Nillable(),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}
