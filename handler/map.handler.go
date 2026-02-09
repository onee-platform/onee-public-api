package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gorilla/mux"
	"github.com/onee-platform/onee-public-api/internal/repository"
	"github.com/onee-platform/onee-public-api/internal/view_pub"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}
type ZipPayload struct {
	ID         *string `json:"id,omitempty"`
	ProvinceId *string `json:"province_id,omitempty"`
	CityId     *string `json:"city_id,omitempty"`
	KecId      *string `json:"kec_id,omitempty"`
	KelId      *string `json:"kel_id,omitempty"`

	Zip *string `json:"zip,omitempty"`

	Q      string
	key    string
	Search string
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
	if v := q.Get("zip"); v != "" {
		if len(v) == 5 {
			req.Zip = &v
		}
	}

	return req
}

// return IsAuthenticated
func authenticateRequest(w http.ResponseWriter, r *http.Request) (bool, string) {
	authToken := r.Header.Get("Authorization")
	if authToken != "" {
		resp := ErrorResponse{
			Error: "Unauthorized!",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
		return false, ""
	}

	shopId := r.Header.Get("X-Onee-Id")
	if shopId == "" {
		resp := ErrorResponse{
			Error: "Unauthorized!",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
		return false, ""
	}

	return true, shopId
}

func getZip(req ZipPayload, limit uint) ([]*view_pub.Zip, *ErrorResponse) {
	var we []exp.Expression
	if req.ID != nil {
		we = append(we, goqu.L("id = ?", *req.ID))
	}
	if req.Search != "" {
		we = append(we, goqu.L("label LIKE ?", req.Search))
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
	}, 0, limit)
	if err != nil && err != sql.ErrNoRows {
		resp := ErrorResponse{
			Error: "Server Busy",
		}
		return nil, &resp
	}

	return list, nil
}

func province(w http.ResponseWriter, r *http.Request) ([]*view_pub.Zip, bool) {
	w.Header().Set("Content-Type", "application/json")

	isAuthenticated, _ := authenticateRequest(w, r)
	if isAuthenticated == false {
		return nil, true
	}

	req := decodeQueryToPayload(r)
	req.Q = "DISTINCT province_id, province_name"
	req.key = "province_id"

	list, zipErr := getZip(req, 100)
	if zipErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(*zipErr)
		return nil, true
	}
	return list, false
}

func city(provinceId string, w http.ResponseWriter, r *http.Request) ([]*view_pub.Zip, bool) {
	w.Header().Set("Content-Type", "application/json")

	isAuthenticated, _ := authenticateRequest(w, r)
	if isAuthenticated == false {
		return nil, true
	}

	req := decodeQueryToPayload(r)
	req.Q = "DISTINCT province_id, province_name, city_id, city_type, city_name"
	req.key = "city_id"
	if provinceId != "" {
		req.ProvinceId = &provinceId
	}

	list, zipErr := getZip(req, 1000)
	if zipErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(*zipErr)
		return nil, true
	}
	return list, false
}

func kecamatan(cityId string, w http.ResponseWriter, r *http.Request) ([]*view_pub.Zip, bool) {
	w.Header().Set("Content-Type", "application/json")

	isAuthenticated, _ := authenticateRequest(w, r)
	if isAuthenticated == false {
		return nil, true
	}

	req := decodeQueryToPayload(r)
	req.Q = "DISTINCT province_id, province_name, city_id, city_type, city_name, kec_id, kec_name"
	req.key = "kec_id"
	if cityId != "" {
		req.CityId = &cityId
	}

	list, zipErr := getZip(req, 100)
	if zipErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(*zipErr)
		return nil, true
	}
	return list, false
}

func kelurahan(kecId string, w http.ResponseWriter, r *http.Request) ([]*view_pub.Zip, bool) {
	w.Header().Set("Content-Type", "application/json")

	isAuthenticated, _ := authenticateRequest(w, r)
	if isAuthenticated == false {
		return nil, true
	}

	req := decodeQueryToPayload(r)
	req.Q = "DISTINCT province_id, province_name, city_id, city_type, city_name, kec_id, kec_name, kel_id, kel_name, zip"
	req.key = "kel_id"
	if kecId != "" {
		req.KecId = &kecId
	}

	list, zipErr := getZip(req, 100)
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
	search := vars["search"]

	w.Header().Set("Content-Type", "application/json")
	isAuthenticated, _ := authenticateRequest(w, r)
	if isAuthenticated == false {
		return
	}

	req := decodeQueryToPayload(r)
	req.Q = "DISTINCT id, province_id, province_name, city_id, city_type, city_name, kec_id, kec_name, kel_id, kel_name, zip"
	req.key = "id"
	if id != "" {
		req.ID = &id
	}
	req.Search = search

	list, zipErr := getZip(req, 100)
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
	resp := ErrorResponse{
		Error: "Location is not found!",
	}
	json.NewEncoder(w).Encode(resp)
	return
}
