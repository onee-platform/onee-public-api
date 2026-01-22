package services

import (
	"context"
	"fmt"
	"github.com/bendt-indonesia/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	biteship "github.com/onee-platform/onee-biteship"
	"github.com/onee-platform/onee-courier"
	"github.com/onee-platform/onee-courier/courier_repo"
	"github.com/onee-platform/onee-go/cached"
	"github.com/onee-platform/onee-go/enum"
	"github.com/onee-platform/onee-go/location"
	"github.com/onee-platform/onee-go/model"
	"github.com/onee-platform/onee-go/view"
	gos "github.com/onee-platform/onee-gosend"
	grb "github.com/onee-platform/onee-grab"
	jne "github.com/onee-platform/onee-jne"
	lala "github.com/onee-platform/onee-lalamove"
	paxel "github.com/onee-platform/onee-paxel"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"strings"
)

func getOriginBranch(shopId string, originBranchId *string) (*cached.Location, *model.AddressInput, error) {
	var branch *view.Branch
	addressQuery := "location_id, address1, address2, phone1, phone2, province_name, city_type, city_name, kec_name, kel_name, zip, lat, lng"
	if originBranchId == nil {
		branch = FindDefaultBranch(shopId, addressQuery)
	} else {
		branch = FindDefaultBranch(shopId, addressQuery)
	}

	if branch == nil {
		return nil, nil, fmt.Errorf("Cabang tidak ditemukan")
	}
	if branch.LocationId == nil {
		return nil, nil, fmt.Errorf("Lokasi Cabang Toko belum di atur di dashboard.")
	}

	originInput, err := location.AddressInputFromBranch(branch)
	if err != nil {
		return nil, nil, err
	}

	return cached.FindLocationById(*branch.LocationId), originInput, nil
}

func InitAPICouriers() error {
	jne.InitJne()

	err := gos.InitGoSend()
	if err != nil {
		return fmt.Errorf("Error on InitGoSend")
	}

	err = grb.InitGrab()
	if err != nil {
		return fmt.Errorf("Error on InitGrab")
	}

	err = lala.InitLalamove()
	if err != nil {
		return fmt.Errorf("Error on InitLalamove")
	}

	err = paxel.InitPaxel()
	if err != nil {
		return fmt.Errorf("Error on InitPaxel")
	}

	err = biteship.InitBiteShip()
	if err != nil {
		return fmt.Errorf("Error on InitBiteShip")
	}
	return nil
}

func AvailableRegularRates(shopId, destLocationId string, originBranchId *string, totalWeightInGram uint) ([]*model.ShippingRate, error) {
	var allRates []*model.ShippingRate
	origin, originAddressInput, err := getOriginBranch(shopId, originBranchId)
	if err != nil {
		return nil, err
	}

	if totalWeightInGram <= 100 {
		return nil, fmt.Errorf("Weight minimum 100gr")
	}
	if origin == nil {
		return nil, fmt.Errorf("Alamat cabang tidak ditemukan, mohon untuk melakukan refresh halaman.")
	}

	destination := cached.FindLocationById(destLocationId)
	if destination == nil {
		return nil, fmt.Errorf("Cache Lokasi Tujuan tidak ditemukan.")
	}
	regularPivots := courier.MappedActiveCouriersByPivots(shopId, "0")

	courierGateway := make(map[enum.Gateway][]*model.PivotsWithCourier)
	courierGateway[enum.GatewayBiteShip] = []*model.PivotsWithCourier{}

	//Perlu ada filter, karena bbrp api akan mereturn semua service yang tersedia untuk suatu kurir/expedisi
	filteredServiceCodes := make(map[enum.CourierCode][]string)

	for cc, pvs := range regularPivots {
		if _, ex := filteredServiceCodes[cc]; !ex {
			filteredServiceCodes[cc] = []string{}
		}
		for _, pv := range pvs {
			if pv.ServiceCode != nil {
				filteredServiceCodes[cc] = append(filteredServiceCodes[cc], *pv.ServiceCode)
			}
		}
	}

	for cc, pvs := range regularPivots {
		var appendRates []*model.ShippingRate
		if cc.HasAPI() {
			switch cc {
			case enum.JNE:
				appendRates, err = jne.GetJneShippingRate(shopId, origin.JNEOrigin, destination.JNEDestination, int(totalWeightInGram))
				if err != nil {
					logrus.Error(err)
					continue
				}
				break
			case enum.PAXEL:
				var paxelServices []paxel.ServiceType
				for _, r := range pvs {
					if r.ServiceCode != nil {
						s := paxel.ServiceType(*r.ServiceCode)
						if s.IsValid() && s.IsInstant() == false {
							paxelServices = append(paxelServices, s)
						}
					}
				}
				appendRates, err = paxel.PaxelShippingRate(shopId, originAddressInput, &model.AddressInput{
					LocationId: &destLocationId,
					Address:    util.NilString("Onee Platform", true),
				}, float64(totalWeightInGram), paxelServices)
				if err != nil {
					logrus.Error(err)
					continue
				}
				break
			}
		} else {
			courierGateway[enum.GatewayBiteShip] = append(courierGateway[enum.GatewayBiteShip], pvs...)
		}

		if len(appendRates) > 0 {
			for _, rate := range appendRates {
				for _, filteredSc := range filteredServiceCodes[cc] {
					if strings.Contains(rate.ServiceCode, filteredSc) {
						allRates = append(allRates, rate)
					}
				}
			}
		}
	}

	for gt, pvs := range courierGateway {
		var appendRates []*model.ShippingRate
		if gt == enum.GatewayBiteShip {
			appendRates, err = biteShipShippingRates(shopId, origin, destination, totalWeightInGram, pvs)
			if err != nil {
				logrus.Error(err)
				continue
			}
		}
		allRates = append(allRates, appendRates...)
	}

	return allRates, nil
}

func biteShipShippingRates(shopId string, origin, destination *cached.Location, totalWeightInGram uint, pvs []*model.PivotsWithCourier) ([]*model.ShippingRate, error) {
	allRates := []*model.ShippingRate{}

	//Biteship only require courierCodes
	var biteShipCouriers []string

	//activeServiceCodes
	//map[bite_courier_code][]{bite_service_codes}
	filteredServiceCodes := make(map[string][]string)

	//map[bite_courier_code|bite_service_code]
	oneeServiceCodes := make(map[string]string)

	for _, pv := range pvs {
		if pv.CourierCode != nil && pv.ServiceCode != nil {
			cc := *pv.CourierCode
			if bcc, ex := biteship.BiteShipCourierCode[cc]; ex {
				biteShipCouriers = append(biteShipCouriers, bcc)
				if _, ex2 := filteredServiceCodes[bcc]; !ex2 {
					filteredServiceCodes[bcc] = []string{}
				}
				key := fmt.Sprintf("%s|%s", pv.CourierCode.String(), *pv.ServiceCode)
				biteShipServiceCode := biteship.BiteShipServiceCode[key]
				filteredServiceCodes[bcc] = append(filteredServiceCodes[bcc], string(biteShipServiceCode))

				key2 := fmt.Sprintf("%s|%s", bcc, biteShipServiceCode)
				oneeServiceCodes[key2] = *pv.ServiceCode
			}

		}
	}
	biteShipCouriers = funk.UniqString(biteShipCouriers)

	if len(biteShipCouriers) > 0 {
		oz := util.StringToInt64(origin.Zip)
		dz := util.StringToInt64(destination.Zip)
		srs, err := biteship.C.GetRatesByZip(context.Background(), biteship.CourierRatesByZipInput{
			OriginPostalCode:      int(oz),
			DestinationPostalCode: int(dz),
			Couriers:              strings.Join(biteShipCouriers, ","),
			Items: []biteship.Item{
				{
					Name:         "X-Onee-ID: " + shopId,
					SubTotal:     100000,
					Quantity:     1,
					WeightInGram: int(totalWeightInGram),
				},
			},
		})
		if err != nil {
			return nil, err
		}

		for _, r := range srs.Pricing {
			bcc := r.CourierCode
			bsc := r.CourierServiceCode
			if funk.ContainsString(filteredServiceCodes[bcc], r.CourierServiceCode) {
				cc := biteship.CourierCodeFromBiteShipCourierCode[bcc]

				key2 := fmt.Sprintf("%s|%s", bcc, bsc)
				oneeSc := oneeServiceCodes[key2]

				rate := &model.ShippingRate{
					CourierCode: cc,
					ServiceCode: oneeSc,
					ServiceName: r.CourierServiceName,
					Cost:        float64(r.Price),
					Etd:         r.Duration,
					Note:        r.Description,
				}
				allRates = append(allRates, rate)
			}
		}
	}

	return allRates, nil
}

func GetCourierList(shopId string) ([]*model.PivotsWithCourier, error) {
	dwe := []exp.Expression{
		goqu.L("hex(shop_id) = ?", shopId),
		goqu.L("courier.is_active = 1"),
	}
	cs, err := courier_repo.GqGetShopCourierPivotWithCourierParent(
		"courier_service.id as courier_service_id, courier.id as courier_id, "+
			"courier_service.is_active as courier_service_is_active, "+
			"service_name, service_code, courier_code, "+
			"height, length, width, max_weight, add_ons, is_instant, courier_service.maintenance_notes as maintenance_notes, is_intercity",
		dwe,
		[]exp.OrderedExpression{
			goqu.I("courier.sort_no").Asc(),
			goqu.I("courier_service.max_weight").Asc(),
		},
	)
	if err != nil {
		logrus.Error(err)
		return nil, fmt.Errorf("Mohon untuk melakukan pengaturan kurir pengiriman.")
	}

	return cs, nil
}
