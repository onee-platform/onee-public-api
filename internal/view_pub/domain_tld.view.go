package view_pub

import (
	"github.com/onee-platform/onee-go/scalar"
)

type DomainTld struct {
	Extension string `db:"extension" json:"extension"`

	IsDnsManagement scalar.Bool `db:"is_dns_management" json:"is_dns_management,omitempty"`
	IsEmailForward  scalar.Bool `db:"is_email_forward" json:"is_email_forward,omitempty"`
	IsHandshake     scalar.Bool `db:"is_handshake" json:"is_handshake,omitempty"`
	IsProtected     scalar.Bool `db:"is_protected" json:"is_protected,omitempty"`

	IsHot  scalar.Bool `db:"is_hot" json:"is_hot,omitempty"`
	IsSale scalar.Bool `db:"is_sale" json:"is_sale,omitempty"`

	RegisterPrice      float64 `db:"register_price" json:"register_price"`
	RegisterPromoPrice float64 `db:"register_promo_price" json:"register_promo_price"`
	RenewalPrice       float64 `db:"renewal_price" json:"renewal_price"`
	RenewalPromoPrice  float64 `db:"renewal_promo_price" json:"renewal_promo_price"`
	TransferPrice      float64 `db:"transfer_price" json:"transfer_price"`
	TransferPromoPrice float64 `db:"transfer_promo_price" json:"transfer_promo_price"`
}
