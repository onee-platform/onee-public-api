package services

import (
	"fmt"
	"github.com/bendt-indonesia/util"
	"github.com/onee-platform/onee-go/enum"
	"github.com/onee-platform/onee-go/pkg/kafka"
	whatsapp "github.com/onee-platform/onee-whatsapp"
	"os"
)

func WhatsappNewQuickCheckout(quickId, senderWid string) error {
	n := whatsapp.QuickCheckoutNewNotification{
		QuickId:   quickId,
		SenderWID: senderWid,
		Source:    enum.NotificationSourceAdm.String(),
	}

	payload := util.ToJSON(n)
	groupKey := os.Getenv("APP_ENV") + "_" + whatsapp.WHATSAPP_CONSUMER_GROUP
	topicKey := os.Getenv("APP_ENV") + "_" + whatsapp.QuickCheckoutNewTopic.String()
	_, _, err := kafka.SendGroupMessageV2(groupKey, topicKey, payload)
	if err != nil {
		return fmt.Errorf("Tidak dapat terhubung dengan layanan Onee WhatsApp (OWA)")
	}

	return nil
}
