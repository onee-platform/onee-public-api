package handler

import (
	"encoding/json"
	"fmt"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/onee-platform/onee-go/repo"
	"github.com/onee-platform/onee-order/quick_checkout"
	"github.com/onee-platform/onee-public-api/internal/services"
	"github.com/onee-platform/onee-public-api/pkg/validate"
	"net/http"
)

type QuickCheckoutInput struct {
	Name              string                            `validate:"required" json:"name"`
	Items             []*quick_checkout.QuickDetailItem `validate:"required,min=1" json:"items"`
	TotalWeightInGram int                               `validate:"required,min=100" json:"total_weight_gr"`
	MaxCheckout       *int                              `json:"max_checkout,omitempty"`
	Email             *string                           `json:"email,omitempty"`
	Phone             *string                           `json:"phone,omitempty"`
	SendNotification  bool                              `json:"send_notification"`
}

type QuickCheckoutResponse struct {
	Success  bool               `json:"success"`
	Data     *QuickCheckoutData `json:"data,omitempty"`
	Validate map[string]string  `json:"validate,omitempty"`
	Message  string             `json:"message,omitempty"`
}

type QuickCheckoutData struct {
	CheckoutURL       string  `json:"checkout_url"`
	Email             *string `json:"email"`
	Name              *string `json:"name"`
	Phone             *string `json:"phone"`
	TotalWeightInGram int     `json:"total_weight_gr"`
	TotalItems        int     `json:"total_items"`
}

func QuickCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	w.Header().Set("Content-Type", "application/json")

	isAuthenticated, shopId := authenticateRequest(w, r)
	if !isAuthenticated {
		return
	}

	var resp QuickCheckoutResponse
	var input QuickCheckoutInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		resp.Message = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Validate
	validateErrs := validate.ValidateStruct(&input)
	if validateErrs != nil {
		resp.Message = "Validation error, please check validate"
		resp.Validate = validateErrs
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	quick, err := services.QuickCheckout(shopId, input.Name, input.Phone, input.Email, input.SendNotification, input.MaxCheckout, float64(input.TotalWeightInGram), input.Items)
	if err != nil {
		resp.Message = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	s, _ := repo.GqFindOneShopTx(nil, "id", shopId, "is_active, slug", []exp.Expression{})
	if s == nil || *s.IsActive == false {
		resp.Message = "Shop is inactive."
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(resp)
		return
	}

	qc := &QuickCheckoutData{
		CheckoutURL:       fmt.Sprintf("https://%s.onee.id/checkout/%s", *s.Slug, quick.ID),
		Name:              quick.Name,
		Phone:             quick.Phone,
		Email:             quick.Email,
		TotalWeightInGram: input.TotalWeightInGram,
		TotalItems:        len(input.Items),
	}

	resp.Success = true
	resp.Message = "OK!"
	resp.Data = qc
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
