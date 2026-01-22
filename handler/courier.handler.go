package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/onee-platform/onee-public-api/internal/services"
	"net/http"
)

type FindRateInput struct {
	CourierCode        string `json:"courier_code"`
	CourierServiceCode string `json:"courier_service_code"`
	DestinationId      string `json:"destination_id"`
	WeightInGram       uint   `json:"weight_in_gram"`
	Height             *uint  `json:"height,omitempty"`
	Length             *uint  `json:"length,omitempty"`
	Width              *uint  `json:"width,omitempty"`
}

func CourierListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isAuthenticated, shopId := authenticateRequest(w, r)
	if isAuthenticated == false {
		return
	}

	activeCouriers, err := services.GetCourierList(shopId)
	if err != nil && err != sql.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Server is busy",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(activeCouriers)
}

func RateListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isAuthenticated, shopId := authenticateRequest(w, r)
	if !isAuthenticated {
		return
	}

	destinationId := "4417"
	var weightInGram uint = 1000

	services.AvailableRegularRates(shopId, destinationId, nil, weightInGram)
	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(ei)
}

func EstimateListHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	isAuthenticated, shopId := authenticateRequest(w, r)
	if !isAuthenticated {
		return
	}

	destinationId := "4417"
	var weightInGram uint = 1000

	services.AvailableRegularRates(shopId, destinationId, nil, weightInGram)
	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(ei)
}
