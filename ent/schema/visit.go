package schema

import (
	"github.com/makesalekz/agents/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/shopspring/decimal"
)

type Visit struct {
	ent.Schema
}

func (Visit) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("tenant_id").Immutable(),
		field.Int64("route_point_id"),
		field.Int64("agent_id"),
		field.Float("checkin_lat").
			GoType(decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric"}),
		field.Float("checkin_lon").
			GoType(decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric"}),
		field.Time("checkin_at"),
		field.Float("checkout_lat").
			GoType(decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric"}).
			Optional(),
		field.Float("checkout_lon").
			GoType(decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "numeric"}).
			Optional(),
		field.Time("checkout_at").Optional().Nillable(),
		field.Int64("duration_seconds").Optional().Default(0),
	}
}

func (Visit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("photos", VisitPhoto.Type),
	}
}

func (Visit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_id"),
		index.Fields("route_point_id"),
	}
}

func (Visit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.CreateUpdateMixin{},
	}
}
