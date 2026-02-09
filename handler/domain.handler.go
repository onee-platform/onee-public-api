package handler

import (
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gorilla/mux"
	"github.com/onee-platform/onee-go/repo"
	"github.com/onee-platform/onee-public-api/internal/repository"
	"github.com/onee-platform/onee-public-api/internal/view_pub"
	"github.com/sirupsen/logrus"
	"net/http"
)

type DomainTldsResponse struct {
	Success  bool                  `json:"success"`
	Data     []*view_pub.DomainTld `json:"data,omitempty"`
	Validate map[string]string     `json:"validate,omitempty"`
	Message  string                `json:"message,omitempty"`
}

func DomainTldsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isAuthenticated, shopId := authenticateRequest(w, r)
	if isAuthenticated == false {
		return
	}

	logrus.Tracef("[%s] DomainHandler", shopId)

	vars := mux.Vars(r)
	wh := vars["with_handshake"]

	var resp DomainTldsResponse
	s, err := repo.GqFindOneShopTx(nil, "id", shopId, "is_active, slug", []exp.Expression{})
	if err != nil || s == nil || *s.IsActive == false {
		resp.Message = "Shop is inactive."
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(resp)
		return
	}

	we := []exp.Expression{
		goqu.L("is_active = ?", 1),
	}

	if wh != "1" {
		we = append(we, goqu.L("is_handshake = ?", 0))
	}

	sq := "extension, is_dns_management, is_handshake, is_protected, is_sale, is_hot, register_price, register_promo_price, renewal_price, renewal_promo_price,transfer_price,transfer_promo_price"
	resp.Data, err = repository.GqGetDomainTld(sq, we, []exp.OrderedExpression{
		goqu.C("is_hot").Desc(),
		goqu.C("is_sale").Desc(),
		goqu.C("sort_no").Desc(),
	}, 3000, 0)
	if err != nil {
		resp.Message = "Server is busy."
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
