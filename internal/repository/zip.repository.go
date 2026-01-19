package repository

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/jmoiron/sqlx"
	"github.com/onee-platform/onee-go/model"
	con "github.com/onee-platform/onee-go/pkg/db/mysql"
	"github.com/onee-platform/onee-grab/internal/view"
	"github.com/sirupsen/logrus"
)

func GqGetVariantFull(Tx *sqlx.Tx, selectQuery string, whereExpr []exp.Expression) ([]*view.VariantFull, error) {
	var results []*view.VariantFull

	//Build query
	query, _, _ := con.GQX.
		Select(goqu.L(selectQuery)).
		From(model.VariantTableName).
		LeftJoin(
			goqu.T(model.ProductTableName),
			goqu.On(goqu.Ex{
				model.ProductTableName + ".id": goqu.I(model.VariantTableName + ".product_id"),
			}),
		).
		Where(whereExpr...).
		ToSQL()

	logrus.Trace(query)
	var err error

	if Tx != nil {
		err = Tx.Select(&results, query)
	} else {
		err = con.DBX.Select(&results, query)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  "VariantFullRepository",
			"func":  "GqFindOneCourierOrderFull",
			"error": err,
		}).Trace("No GqFindOneCourierOrderFull was found")
		return nil, err
	}

	return results, nil
}
