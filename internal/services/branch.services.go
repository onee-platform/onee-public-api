package services

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/onee-platform/onee-go/repo"
	"github.com/onee-platform/onee-go/view"
	"github.com/sirupsen/logrus"
)

func FindDefaultBranch(shopId, addQuery string) *view.Branch {
	if len(addQuery) > 2 {
		addQuery = ", " + addQuery
	}
	b, err := repo.GqFindOneBranch(nil, "shop_id", shopId, "id, code, name "+addQuery, []exp.Expression{
		goqu.L("hex(shop_id) = ?", shopId),
		goqu.L("is_default = ?", "1"),
		goqu.L("is_active = ?", "1"),
	})
	if err != nil {
		logrus.Error(err.Error())
		return nil
	}

	return b
}
