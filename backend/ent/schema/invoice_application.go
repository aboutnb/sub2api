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

// InvoiceApplication stores the local ownership and workflow state for an
// invoice application submitted through the site-wide XZNOAuth client.
type InvoiceApplication struct {
	ent.Schema
}

func (InvoiceApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invoice_applications"},
	}
}

func (InvoiceApplication) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.JSON("order_ids", []int64{}).
			Default(func() []int64 { return []int64{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("order_nos", []string{}).
			Default(func() []string { return []string{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bool("need_pay_tax").Default(false),
		field.JSON("tax_order_nos", []string{}).
			Default(func() []string { return []string{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("validation_snapshot", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("external_id").Optional().Nillable().MaxLen(128),
		field.String("status").MaxLen(32).Default("draft"),
		field.String("title").Optional().Nillable().MaxLen(255),
		field.String("recipient_email").Optional().Nillable().MaxLen(255),
		field.String("total_amount").MaxLen(32).Default(""),
		field.String("currency").MaxLen(8).Default("CNY"),
		field.JSON("external_snapshot", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("request_id").Optional().Nillable().MaxLen(128),
		field.String("error_code").Optional().Nillable().MaxLen(128),
		field.String("error_message").Optional().Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("created_at").Immutable().Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (InvoiceApplication) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("user_id", "status"),
		index.Fields("external_id").Unique().Annotations(entsql.IndexWhere("external_id IS NOT NULL AND external_id <> ''")),
	}
}
