package schema

import (
	"gitlab.calendaria.team/services/agents/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Onboarding struct {
	ent.Schema
}

func (Onboarding) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("tenant_id").Immutable(),
		field.Int64("agent_id"),
		field.Int64("new_tenant_id"),
		field.String("store_name"),
		field.String("owner_phone"),
	}
}

func (Onboarding) Edges() []ent.Edge {
	return nil
}

func (Onboarding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "agent_id"),
	}
}

func (Onboarding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.CreateUpdateMixin{},
	}
}
