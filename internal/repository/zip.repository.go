package repository

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/jmoiron/sqlx"
	"github.com/onee-platform/onee-go/model"
	con "github.com/onee-platform/onee-go/pkg/db/mysql"
	"github.com/onee-platform/onee-public-api/internal/view"
	"github.com/sirupsen/logrus"
)

func GqFindLocation(selectQuery string, whereExpr []exp.Expression) (*model.Location, error) {
	m := model.Location{}

	if selectQuery == "" {
		selectQuery = "id, province_id, province_name, city_type, city_id, city_name, kec_id, kec_name, kel_id, kel_name, zip, jne_origin, jne_destination"
	}

	//Build query
	query, _, _ := con.GQX.
		Select(goqu.L(selectQuery)).
		From(model.ZipTableName).
		Where(whereExpr...).
		Limit(1).
		ToSQL()

	logrus.Trace(query)
	err := con.DBX.Get(&m, query)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  "ZipRepository",
			"func":  "GqFindLocation",
			"error": err,
		}).Warn("Unable to [GET] Courier Ids")
		return nil, err
	}

	return &m, nil
}

func GqGetZip(Tx *sqlx.Tx, selectQuery string, whereExpr []exp.Expression, orderedExp []exp.OrderedExpression, offset, limit uint) ([]*view.Zip, error) {
	var results []*view.Zip

	//Build query
	query, _, _ := con.GQX.
		Select(goqu.L(selectQuery)).
		From(model.ZipTableName).
		Where(whereExpr...).
		Order(orderedExp...).
		Offset(offset).
		Limit(limit).
		ToSQL()

	logrus.Debug(query)
	var err error

	if Tx != nil {
		err = Tx.Select(&results, query)
	} else {
		err = con.DBX.Select(&results, query)
	}

	if err != nil {
		return nil, err
	}

	return results, nil
}
