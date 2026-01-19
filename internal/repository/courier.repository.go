package repository

import (
	"github.com/bendt-indonesia/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/onee-platform/onee-go/model"
	con "github.com/onee-platform/onee-go/pkg/db/mysql"
	"github.com/onee-platform/onee-go/view"
	"github.com/sirupsen/logrus"
)

func GqGetCourier(selectQuery string, whereExpr []exp.Expression, orderByExpr []exp.OrderedExpression, limit uint64, offset uint64) ([]*view.Courier, error) {
	var results []*view.Courier

	//Where expressions always appended on the behind
	whereExpr = util.GoqMergeWhereExpression(whereExpr)

	//Default OrderExpressions may have priority order before the arguments
	var orderByExpressions []exp.OrderedExpression
	orderByExpressions = append(orderByExpressions, orderByExpr...)

	//Build query
	query, _, _ := con.GQX.
		Select(goqu.L(selectQuery)).
		From(model.CourierTableName).
		Where(whereExpr...).
		Order(orderByExpressions...).
		Limit(uint(limit)).
		Offset(uint(offset)).
		ToSQL()

	logrus.Trace(query)
	err := con.DBX.Select(&results, query)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  "CourierRepository",
			"func":  "GqGetCourier",
			"error": err,
		}).Warn("Unable to [GET] CourierService")
		return nil, err
	}

	return results, err
}

func GqGetCourierService(selectQuery string, whereExpr []exp.Expression, orderByExpr []exp.OrderedExpression, limit uint64, offset uint64) ([]*view.CourierService, error) {
	var results []*view.CourierService

	//Where expressions always appended on the behind
	whereExpr = util.GoqMergeWhereExpression(whereExpr)

	//Default OrderExpressions may have priority order before the arguments
	var orderByExpressions []exp.OrderedExpression
	orderByExpressions = append(orderByExpressions, orderByExpr...)

	//Build query
	query, _, _ := con.GQX.
		Select(goqu.L(selectQuery)).
		From(model.CourierServiceTableName).
		Where(whereExpr...).
		Order(orderByExpressions...).
		Limit(uint(limit)).
		Offset(uint(offset)).
		ToSQL()

	logrus.Trace(query)
	err := con.DBX.Select(&results, query)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  "CourierRepository",
			"func":  "GqGetCourierService",
			"error": err,
		}).Warn("Unable to [GET] CourierService")
		return nil, err
	}

	return results, err
}
