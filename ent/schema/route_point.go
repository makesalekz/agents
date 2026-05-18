package schema

import (
	"github.com/makesalekz/agents/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type RoutePoint struct {
	ent.Schema
}

func (RoutePoint) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("route_id"),
		field.Int64("store_tenant_id"),
		field.Int32("order_num"),
		field.Int64("visit_id").Optional().Default(0),
	}
}

func (RoutePoint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("route", Route.Type).Ref("points").Unique().Required().Field("route_id"),
	}
}

func (RoutePoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("route_id", "order_num"),
	}
}

func (RoutePoint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.CreateUpdateMixin{},
	}
}
