package services

import (
	"github.com/bendt-indonesia/util"
	"github.com/onee-platform/onee-go/enum"
	"github.com/onee-platform/onee-go/repo"
	"github.com/onee-platform/onee-go/view"
	"github.com/onee-platform/onee-order/quick_checkout"
)

func QuickCheckout(shopId, name string, phone, email *string, sendWA bool, totalWeightInGram float64, items []*quick_checkout.QuickDetailItem) (*view.Quick, error) {
	Tx := repo.BeginTransaction()
	qc := quick_checkout.QuickCheckout{
		Tx:                Tx,
		CreatedById:       util.NilString("1", false),
		Source:            enum.NotificationSourceSystem,
		ShopId:            shopId,
		QuickDetailItems:  items,
		Phone:             phone,
		Email:             email,
		Name:              &name,
		TotalWeightInGram: &totalWeightInGram,
	}
	err := qc.Execute()
	if err != nil {
		_ = repo.RollbackTransaction(Tx)
		return nil, err
	}
	_ = repo.CommitTransaction(Tx)

	if sendWA {
		if qc.NewQuick != nil && phone != nil {
			waPhone := util.PhonePrefix(*phone)
			go WhatsappNewQuickCheckout(qc.NewQuick.ID, waPhone)
		}
	}

	return qc.NewQuick, nil
}
