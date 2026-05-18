package data

import (
	"context"

	"github.com/makesalekz/agents/ent"
	"github.com/makesalekz/agents/ent/visitphoto"
)

type VisitPhotosRepo interface {
	Create(ctx context.Context, dto VisitPhotoDto) (*ent.VisitPhoto, error)
	ListByVisit(ctx context.Context, visitID int64) ([]*ent.VisitPhoto, error)
}

type visitPhotosRepo struct {
	db *ent.Client
}

func NewVisitPhotosRepo(d *Data) VisitPhotosRepo {
	return &visitPhotosRepo{db: d.db}
}

func (r *visitPhotosRepo) Create(ctx context.Context, dto VisitPhotoDto) (*ent.VisitPhoto, error) {
	return r.db.VisitPhoto.Create().
		SetVisitID(dto.VisitID).
		SetMediaURL(dto.MediaURL).
		Save(ctx)
}

func (r *visitPhotosRepo) ListByVisit(ctx context.Context, visitID int64) ([]*ent.VisitPhoto, error) {
	return r.db.VisitPhoto.Query().
		Where(visitphoto.VisitID(visitID)).
		Order(ent.Asc(visitphoto.FieldID)).
		All(ctx)
}
