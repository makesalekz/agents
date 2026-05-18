package schema

import (
	"github.com/makesalekz/agents/ent/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type VisitPhoto struct {
	ent.Schema
}

func (VisitPhoto) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("visit_id"),
		field.String("media_url").NotEmpty(),
	}
}

func (VisitPhoto) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("visit", Visit.Type).Ref("photos").Unique().Required().Field("visit_id"),
	}
}

func (VisitPhoto) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("visit_id"),
	}
}

func (VisitPhoto) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.CreateUpdateMixin{},
	}
}
