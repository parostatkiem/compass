package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

type inMemoryRepository struct {
	store map[string]*model.Application
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.Application)}
}

func (r *inMemoryRepository) GetByID(ctx context.Context, tenant, id string) (*model.Application, error) {
	application := r.store[id]

	if application == nil || application.Tenant != tenant {
		return nil, errors.New("application not found")
	}

	return application, nil
}

//TODO: remove this function after migrating to Database
func (r *inMemoryRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	application := r.store[id]

	if application == nil || application.Tenant != tenant {
		return false, nil
	}

	return true, nil
}

// TODO: Make filtering and paging
func (r *inMemoryRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	var items []*model.Application
	for _, item := range r.store {
		if item.Tenant == tenant {
			items = append(items, item)
		}
	}

	return &model.ApplicationPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

// TODO: add pagination when PR-181 is merged
func (r *inMemoryRepository) ListByScenarios(ctx context.Context, tenantUUID uuid.UUID, scenarios []string, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	var scenariosFilers []*labelfilter.LabelFilter

	for _, scenarioValue := range scenarios {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, scenarioValue)
		scenariosFilers = append(scenariosFilers, &labelfilter.LabelFilter{Key: model.ScenariosKey, Query: &query})
	}

	stmt, err := label.FilterQuery(model.ApplicationLabelableObject, label.UnionSet, tenantUUID, scenariosFilers)
	if err != nil {
		return nil, errors.Wrap(err, "while creating filter query")
	}

	var apps []string

	err = persist.Select(&apps, stmt)

	if err != nil {
		return &model.ApplicationPage{
			Data:       []*model.Application{},
			TotalCount: 0,
			PageInfo: &pagination.Page{
				StartCursor: "",
				EndCursor:   "",
				HasNextPage: false,
			}}, nil
	}

	var items []*model.Application

	for _, id := range apps {
		app, found := r.store[id]
		if found {
			items = append(items, app)
		}
	}

	return &model.ApplicationPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) Create(ctx context.Context, item *model.Application) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	found := r.findApplicationNameWithinTenant(item.Tenant, item.Name)
	if found {
		return errors.New("Application name is not unique within tenant")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Update(ctx context.Context, item *model.Application) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	oldApplication := r.store[item.ID]
	if oldApplication == nil {
		return errors.New("application not found")
	}

	if oldApplication.Name != item.Name {
		found := r.findApplicationNameWithinTenant(item.Tenant, item.Name)
		if found {
			return errors.New("Application name is not unique within tenant")
		}
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Delete(ctx context.Context, item *model.Application) error {
	if item == nil {
		return nil
	}

	delete(r.store, item.ID)

	return nil
}

func (r *inMemoryRepository) findApplicationNameWithinTenant(tenant, name string) bool {
	for _, app := range r.store {
		if app.Name == name && app.Tenant == tenant {
			return true
		}
	}
	return false
}
