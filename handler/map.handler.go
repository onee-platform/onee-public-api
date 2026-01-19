package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gorilla/mux"
	"github.com/onee-platform/onee-public-api/internal/repository"
	"github.com/onee-platform/onee-public-api/internal/view"
	"net/http"
)

type ZipErrorResponse struct {
	Error string `json:"error"`
}
type ZipPayload struct {
	ID         *string `json:"id,omitempty"`
	ProvinceId *string `json:"province_id,omitempty"`
	CityId     *string `json:"city_id,omitempty"`
	KecId      *string `json:"kec_id,omitempty"`
	KelId      *string `json:"kel_id,omitempty"`

	Zip *string `json:"zip,omitempty"`

	Q   string
	key string
}

func decodeQueryToPayload(r *http.Request) ZipPayload {
	q := r.URL.Query()
	var req ZipPayload

	if v := q.Get("province_id"); v != "" {
		req.ProvinceId = &v
	}
	if v := q.Get("city_id"); v != "" {
		req.CityId = &v
	}
	if v := q.Get("kec_id"); v != "" {
		req.KecId = &v
	}
	if v := q.Get("kel_id"); v != "" {
		req.KelId = &v
	}

	if v := q.Get("province"); v != "" {
		req.ProvinceId = &v
	}
	if v := q.Get("city"); v != "" {
		req.CityId = &v
	}
	if v := q.Get("kec"); v != "" {
		req.KecId = &v
	}
	if v := q.Get("kel"); v != "" {
		req.KelId = &v
	}

	if v := q.Get("zip"); v != "" {
		if len(v) == 5 {
			req.Zip = &v
		}
	}

	return req
}

func authenticateRequest(w http.ResponseWriter, r *http.Request) *ZipPayload {
	authToken := r.Header.Get("Authorization-Id")
	if authToken != "" {
		resp := ZipErrorResponse{
			Error: "Unauthorized!",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
		return nil
	}

	req := decodeQueryToPayload(r)
	return &req
}

func getZip(req ZipPayload) ([]*view.Zip, *ZipErrorResponse) {
	var we []exp.Expression
	if req.ID != nil {
		we = append(we, goqu.L("id = ?", *req.ID))
	}
	if req.ProvinceId != nil {
		we = append(we, goqu.L("province_id = ?", *req.ProvinceId))
	}
	if req.CityId != nil {
		we = append(we, goqu.L("city_id = ?", *req.CityId))
	}
	if req.KecId != nil {
		we = append(we, goqu.L("kec_id = ?", *req.KecId))
	}
	if req.KelId != nil {
		we = append(we, goqu.L("kel_id = ?", *req.KelId))
	}
	if req.Zip != nil {
		we = append(we, goqu.L("zip = ?", *req.Zip))
	}

	list, err := repository.GqGetZip(nil, req.Q, we, []exp.OrderedExpression{
		goqu.C(req.key).Asc(),
	}, 0, 100)
	if err != nil && err != sql.ErrNoRows {
		resp := ZipErrorResponse{
			Error: "Server Busy",
		}
		return nil, &resp
	}

	return list, nil
}

func province(w http.ResponseWriter, r *http.Request) ([]*view.Zip, bool) {
	w.Header().Set("Content-Type", "application/json")

	req := authenticateRequest(w, r)
	if req == nil {
		return nil, true
	}
	req.Q = "DISTINCT province_id, province_name"
	req.key = "province_id"

	list, zipErr := getZip(*req)
	if zipErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(*zipErr)
		return nil, true
	}
	return list, false
}

func city(provinceId string, w http.ResponseWriter, r *http.Request) ([]*view.Zip, bool) {
	w.Header().Set("Content-Type", "application/json")

	req := authenticateRequest(w, r)
	if req == nil {
		return nil, true
	}
	req.Q = "DISTINCT province_id, province_name, city_id, city_type, city_name"
	req.key = "city_id"
	if provinceId != "" {
		req.ProvinceId = &provinceId
	}

	list, zipErr := getZip(*req)
	if zipErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(*zipErr)
		return nil, true
	}
	return list, false
}

func kecamatan(cityId string, w http.ResponseWriter, r *http.Request) ([]*view.Zip, bool) {
	w.Header().Set("Content-Type", "application/json")

	req := authenticateRequest(w, r)
	if req == nil {
		return nil, true
	}
	req.Q = "DISTINCT province_id, province_name, city_id, city_type, city_name, kec_id, kec_name"
	req.key = "kec_id"
	if cityId != "" {
		req.CityId = &cityId
	}

	list, zipErr := getZip(*req)
	if zipErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(*zipErr)
		return nil, true
	}
	return list, false
}

func kelurahan(kecId string, w http.ResponseWriter, r *http.Request) ([]*view.Zip, bool) {
	w.Header().Set("Content-Type", "application/json")

	req := authenticateRequest(w, r)
	if req == nil {
		return nil, true
	}
	req.Q = "DISTINCT province_id, province_name, city_id, city_type, city_name, kec_id, kec_name, kel_id, kel_name, zip"
	req.key = "kel_id"
	if kecId != "" {
		req.KecId = &kecId
	}

	list, zipErr := getZip(*req)
	if zipErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(*zipErr)
		return nil, true
	}
	return list, false
}

func ProvinceListHandler(w http.ResponseWriter, r *http.Request) {
	provinces, isError := province(w, r)
	if isError {
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(provinces)
}

func CityListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provinceId := vars["province_id"]
	data, isError := city(provinceId, w, r)
	if isError {
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func KecListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cityId := vars["city_id"]
	data, isError := kecamatan(cityId, w, r)
	if isError {
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func KelListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kecId := vars["kec_id"]
	data, isError := kelurahan(kecId, w, r)
	if isError {
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func LocationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	req := authenticateRequest(w, r)
	if req == nil {
		return
	}

	req.Q = "DISTINCT id, province_id, province_name, city_id, city_type, city_name, kec_id, kec_name, kel_id, kel_name, zip"
	req.key = "id"
	if id != "" {
		req.ID = &id
	}

	list, zipErr := getZip(*req)
	if zipErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(*zipErr)
		return
	}
	if len(list) > 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(list[0])
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	resp := ZipErrorResponse{
		Error: "Location is not found!",
	}
	json.NewEncoder(w).Encode(resp)
	return
}
