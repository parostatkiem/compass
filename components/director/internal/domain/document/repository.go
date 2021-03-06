package document

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const documentTable = "public.documents"

var documentColumns = []string{"id", "tenant_id", "app_id", "title", "display_name", "description", "format", "kind", "data"}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in model.Document) (Entity, error)
	FromEntity(in Entity) (model.Document, error)
}

type repository struct {
	*repo.ExistQuerier
	*repo.SingleGetter
	*repo.Deleter
	*repo.PageableQuerier
	*repo.Creator

	conv Converter
}

func NewRepository(conv Converter) *repository {
	return &repository{
		ExistQuerier:    repo.NewExistQuerier(documentTable, "tenant_id"),
		SingleGetter:    repo.NewSingleGetter(documentTable, "tenant_id", documentColumns),
		Deleter:         repo.NewDeleter(documentTable, "tenant_id"),
		PageableQuerier: repo.NewPageableQuerier(documentTable, "tenant_id", documentColumns),
		Creator:         repo.NewCreator(documentTable, documentColumns),

		conv: conv,
	}
}

func (r *repository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.ExistQuerier.Exists(ctx, tenant, repo.Conditions{{Field: "id", Val: id}})
}

func (r *repository) GetByID(ctx context.Context, tenant, id string) (*model.Document, error) {
	var entity Entity
	if err := r.SingleGetter.Get(ctx, tenant, repo.Conditions{{Field: "id", Val: id}}, &entity); err != nil {
		return nil, err
	}

	docModel, err := r.conv.FromEntity(entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Document entity to model")
	}

	return &docModel, nil
}

func (r *repository) Create(ctx context.Context, item *model.Document) error {
	if item == nil {
		return errors.New("Document cannot be empty")
	}

	entity, err := r.conv.ToEntity(*item)
	if err != nil {
		return errors.Wrap(err, "while creating Document entity from model")
	}

	return r.Creator.Create(ctx, entity)
}

func (r *repository) CreateMany(ctx context.Context, items []*model.Document) error {
	for _, item := range items {
		if item == nil {
			return errors.New("Document cannot be empty")
		}
		err := r.Create(ctx, item)
		if err != nil {
			return errors.Wrapf(err, "while creating Document with ID %s", item.ID)
		}
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, tenant, id string) error {
	return r.Deleter.DeleteOne(ctx, tenant, repo.Conditions{{Field: "id", Val: id}})
}

func (r *repository) DeleteAllByApplicationID(ctx context.Context, tenant string, applicationID string) error {
	return r.Deleter.DeleteMany(ctx, tenant, repo.Conditions{{Field: "app_id", Val: applicationID}})
}

func (r *repository) ListByApplicationID(ctx context.Context, tenant string, applicationID string, pageSize int, cursor string) (*model.DocumentPage, error) {
	appCondition := fmt.Sprintf("%s = %s", "app_id", pq.QuoteLiteral(applicationID))

	var entityCollection Collection
	page, totalCount, err := r.PageableQuerier.List(ctx, tenant, pageSize, cursor, "id", &entityCollection, appCondition)
	if err != nil {
		return nil, err
	}

	var items []*model.Document

	for _, entity := range entityCollection {
		docModel, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Document entity to model")
		}

		items = append(items, &docModel)
	}
	return &model.DocumentPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}
