package handler

import (
	"encoding/json"
	"github.com/onee-platform/onee-go/view"
	"github.com/onee-platform/onee-order/quick_checkout"
	"github.com/onee-platform/onee-public-api/internal/services"
	"github.com/onee-platform/onee-public-api/pkg/validate"
	"net/http"
)

type QuickCheckoutInput struct {
	Name              string                            `validate:"required" json:"name"`
	Items             []*quick_checkout.QuickDetailItem `validate:"required,min=1" json:"items"`
	TotalWeightInGram int64                             `validate:"required,min=100" json:"total_weight_gr"`
	Phone             *string                           `json:"phone"`
	Email             *string                           `json:"email"`
	SendNotification  bool                              `json:"send_notification"`
}

type QuickCheckoutResponse struct {
	Success  bool              `json:"success"`
	Data     *view.Quick       `json:"data,omitempty"`
	Validate map[string]string `json:"validate,omitempty"`
	Message  string            `json:"message,omitempty"`
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
		resp.Message = "Invalid Payload"
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

	var err error
	resp.Data, err = services.QuickCheckout(shopId, input.Name, input.Phone, input.Email, input.SendNotification, float64(input.TotalWeightInGram), input.Items)
	if err != nil {
		resp.Message = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Success = true
	resp.Message = "OK!"
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
