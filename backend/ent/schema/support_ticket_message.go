package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SupportTicketMessage is the persistence description for one ticket message.
// Runtime access uses SQL so this schema intentionally has no ID override and no edges.
type SupportTicketMessage struct {
	ent.Schema
}

func (SupportTicketMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "support_ticket_messages"},
	}
}

func (SupportTicketMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("ticket_id").Comment("工单 ID"),
		field.Int64("author_id").Comment("作者用户 ID"),
		field.String("author_role").MaxLen(16).NotEmpty().Comment("作者角色: user, admin"),
		field.String("body").
			NotEmpty().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("消息正文"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicketMessage) Edges() []ent.Edge {
	return nil
}

func (SupportTicketMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ticket_id"),
	}
}
