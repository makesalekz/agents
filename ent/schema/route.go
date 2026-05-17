package schema

import (
	"gitlab.calendaria.team/services/agents/ent/enum"
	"gitlab.calendaria.team/services/agents/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Route struct {
	ent.Schema
}

func (Route) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("tenant_id").Immutable(),
		field.Int64("agent_id"),
		field.Time("date"),
		field.Enum("status").GoType(enum.RouteStatus("")).Default(enum.Planned.Value()),
	}
}

func (Route) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("points", RoutePoint.Type),
	}
}

func (Route) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_id", "date"),
	}
}

func (Route) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.CreateUpdateMixin{},
	}
}
