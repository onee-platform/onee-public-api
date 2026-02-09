package repository

import (
	"github.com/bendt-indonesia/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/onee-platform/onee-go/model"
	con "github.com/onee-platform/onee-go/pkg/db/mysql"
	"github.com/onee-platform/onee-public-api/internal/view_pub"
	"github.com/sirupsen/logrus"
)

func GqGetDomainTld(selectQuery string, whereExpr []exp.Expression, orderByExpr []exp.OrderedExpression, limit, offset uint64) ([]*view_pub.DomainTld, error) {
	var results []*view_pub.DomainTld

	//Where expressions always appended on the behind
	whereExpr = util.GoqMergeWhereExpression(whereExpr)

	//Default OrderExpressions may have priority order before the arguments
	var orderByExpressions []exp.OrderedExpression
	orderByExpressions = append(orderByExpressions, orderByExpr...)

	//Build query
	query, _, _ := con.GQX.
		Select(goqu.L(selectQuery)).
		From(model.DomainTldTableName).
		Where(whereExpr...).
		Order(orderByExpressions...).
		Limit(uint(limit)).
		Offset(uint(offset)).
		ToSQL()

	logrus.Trace(query)
	err := con.DBX.Select(&results, query)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  "DomainTldRepository",
			"func":  "GqGetDomainTld",
			"error": err,
		}).Warn("Unable to [GET] DomainTld")
		return nil, err
	}

	return results, err
}
