package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type Creator struct {
	tableName string
	columns   []string
}

func NewCreator(tableName string, columns []string) *Creator {
	return &Creator{
		tableName: tableName,
		columns:   columns,
	}
}

func (c *Creator) Create(ctx context.Context, dbEntity interface{}) error {
	if dbEntity == nil {
		return errors.New("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var values []string
	for _, c := range c.columns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	stmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", c.tableName, strings.Join(c.columns, ", "), strings.Join(values, ", "))

	_, err = persist.NamedExec(stmt, dbEntity)
	if pqerr, ok := err.(*pq.Error); ok {
		if pqerr.Code == persistence.UniqueViolation {
			return &notUniqueError{}
		}
	}

	return errors.Wrapf(err, "while inserting row to '%s' table", c.tableName)
}
