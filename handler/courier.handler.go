package handler

import (
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/onee-platform/onee-go/scalar"
	"github.com/onee-platform/onee-public-api/internal/repository"
	"net/http"
	"time"
)

type CourierListResponse struct {
	Data []Courier `json:"data"`
}

type Courier struct {
	ID              string           `json:"id"`
	CourierCode     *string          `json:"courier_code"`
	CourierName     *string          `json:"courier_name"`
	IsActive        scalar.Bool      `json:"is_active"`
	CourierServices []CourierService `json:"courier_services"`
}

type CourierService struct {
	ServiceCode      *string      `json:"service_code,omitempty"`
	ServiceName      *string      `json:"service_name,omitempty"`
	IsActive         *scalar.Bool `json:"is_active,omitempty"`
	IsInstant        *scalar.Bool `json:"is_instant,omitempty"`
	Height           *uint        `json:"height,omitempty"`
	Length           *uint        `json:"length,omitempty"`
	Width            *uint        `json:"width,omitempty"`
	MaxWeight        *uint        `json:"max_weight,omitempty"`
	MinWeight        *uint        `json:"min_weight,omitempty"`
	MaintenanceNotes *string      `json:"maintenance_notes,omitempty"`
	UpdatedAt        *time.Time   `json:"updated_at,omitempty"`
}

func CourierListHandler(w http.ResponseWriter, r *http.Request) {
	couriers, err := repository.GqGetCourier("courier_code, courier_name, is_active, is_ods",
		[]exp.Expression{},
		[]exp.OrderedExpression{
			goqu.C("is_active").Desc(),
		},
		100, 0)
	if err != nil {
		resp := ErrorResponse{
			Error: "Server is busy",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var resp CourierListResponse

	for _, row := range couriers {

		c := Courier{
			ID:          row.ID,
			CourierCode: row.CourierCode,
			CourierName: row.CourierName,
		}
		if row.IsActive != nil && *row.IsActive == 1 {
			c.IsActive = true
		}
		resp.Data = append(resp.Data, c)
	}
	//sq := "service_code,service_name,is_active,is_instant,height,length,width,max_weight,min_weight,maintenance_notes,updated_at"

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
