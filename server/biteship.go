package main

import (
	"encoding/json"
	"fmt"
	"github.com/bendt-indonesia/env"
	"github.com/bendt-indonesia/util"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/onee-platform/onee-go/cached"
	con "github.com/onee-platform/onee-go/pkg/db/mysql"
	"github.com/onee-platform/onee-go/pkg/wlog"
	"github.com/onee-platform/onee-go/repo"
	"github.com/onee-platform/onee-public-api/internal/repository"
	"github.com/onee-platform/onee-public-api/internal/services"
	"github.com/onee-platform/onee-public-api/internal/view"
	"github.com/sirupsen/logrus"
	"io"

	"net/http"
	"strings"
	"time"
)

type AreaResponse struct {
	Success bool   `json:"success"`
	Areas   []Area `json:"areas"`
}

type Area struct {
	ID          string `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Country     string `db:"country_name" json:"country_name"`
	CountryCode string `db:"country_code" json:"country_code"`

	Admin1Name string `db:"administrative_division_level_1_name" json:"administrative_division_level_1_name"`
	Admin1Type string `db:"administrative_division_level_1_type" json:"administrative_division_level_1_type"`

	Admin2Name string `db:"administrative_division_level_2_name" json:"administrative_division_level_2_name"`
	Admin2Type string `db:"administrative_division_level_2_type" json:"administrative_division_level_2_type"`

	Admin3Name string `db:"administrative_division_level_3_name" json:"administrative_division_level_3_name"`
	Admin3Type string `db:"administrative_division_level_3_type" json:"administrative_division_level_3_type"`

	Zip string `db:"zip" json:"zip"`
}

func main() {
	//Loading env files
	err := env.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}
	//Establish DB Connection
	wlog.InitLog()
	con.InitSqlX()
	con.InitGoqu()
	cached.InitAll()

	err = services.InitAPICouriers()
	if err != nil {
		logrus.Fatal(err)
	}

	rates, err := services.AvailableRegularRates("eb0ae4d9a87f4fd6b2e17d7cbab71853", "4613", nil, 1000)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(util.ToJSON(rates))

	//for true {
	//	zip, err := repository.GqGetZip(nil, "DISTINCT province_name, city_name, kec_name, zip", []exp.Expression{
	//		goqu.L("biteship_id = ?", ""),
	//		//goqu.L("zip >= ?", prevZip),
	//	}, []exp.OrderedExpression{
	//		goqu.C("zip").Asc(),
	//	}, 0, 100)
	//
	//	for _, z := range zip {
	//		err = fetchAndStore("kec_name", z)
	//		if err != nil {
	//			fmt.Println(err.Error())
	//			break
	//		}
	//		//time.Sleep(time.Millisecond * 300)
	//	}
	//}

}

func fetchAndStore(by string, z *view.Zip) error {
	var input string
	if by == "zip" && z.Zip != nil {
		input = *z.Zip
	} else if by == "kec_name" && z.KecName != nil {
		input = *z.KecName
	}

	if input == "" {
		return fmt.Errorf("No Input")
	}
	resp, err := fetchAreas(input)
	if err != nil {
		fmt.Println(err.Error())
		return fmt.Errorf("ERR-1")
	}

	if resp != nil && resp.Success {
		for _, area := range resp.Areas {
			name := area.Name
			name = util.TrimSuffix(name, ".")
			name = util.TrimSuffix(name, " ")
			area.Zip = util.ExtractSuffix(name, 5)
			err = InsertArea(&area)
			if err != nil {
				fmt.Printf("[INSERT] %s\n", err.Error())
			}
		}

		if by == "zip" {
			if len(resp.Areas) == 1 {
				err = repo.ExecRawQuery("UPDATE `zip` SET biteship_id = ? WHERE zip = ?", resp.Areas[0].ID, input)
				if err != nil {
					fmt.Println(err.Error())
					return fmt.Errorf("ERR-2")
				}
			} else {
				for _, area := range resp.Areas {
					err = repo.ExecRawQuery("UPDATE `zip` SET biteship_id = ? WHERE zip = ? AND kec_name = ?", area.ID, input, area.Admin3Name)
					if err != nil {
						fmt.Println(err.Error())
						fmt.Println("ERR-4")
					}
				}
				fmt.Println(util.ToJSON(*resp))
				fmt.Println("=============================================")
			}
		} else if by == "kec_name" {
			for _, area := range resp.Areas {
				name := area.Name
				name = util.TrimSuffix(name, ".")
				name = util.TrimSuffix(name, " ")
				area.Zip = util.ExtractSuffix(name, 5)
				provinceName := strings.ToLower(area.Admin1Name)
				cityName := strings.ToLower(area.Admin2Name)
				kecName := strings.ToLower(area.Admin3Name)

				loc, _ := repository.GqFindLocation("id, province_name, city_name, kec_name, zip, label, biteship_id as kel_id", []exp.Expression{
					goqu.L("province_name = ?", provinceName),
					goqu.L("city_name = ?", cityName),
					goqu.L("kec_name = ?", kecName),
				})
				if loc != nil {
					fmt.Printf("[FINDR] Province: `%s` City: `%s` Kec: `%s` Zip: `%s`\n", *z.ProvinceName, *z.CityName, *z.KecName, *z.Zip)
					fmt.Printf("[ API ] Province: `%s` City: `%s` Kec: `%s` Zip: `%s` Label: `%s`\n", area.Admin1Name, area.Admin2Name, area.Admin3Name, area.Zip, area.Name)
					fmt.Printf("[%-5s] Province: `%s` City: `%s` Kec: `%s` Zip: `%s` Label: `%s` BiteshipID: `%s`\n", loc.ID, *loc.ProvinceName, *loc.CityName, *loc.KecName, *loc.Zip, *loc.Label, *loc.KelId)
					confirm := util.ConfirmCmd("", fmt.Sprintf("ID: %s", loc.ID))
					if confirm {
						fmt.Println("Updated.")
					} else {
						fmt.Println("Skipped!")
					}
				}
			}
			if len(resp.Areas) == 1 {

				//err = repo.ExecRawQuery("UPDATE `zip` SET biteship_id = ? WHERE zip = ?", resp.Areas[0].ID, input)
				//if err != nil {
				//	fmt.Println(err.Error())
				//	return fmt.Errorf("ERR-2")
				//}
			}
		}
	}
	return nil
}

func fetchAreas(zip string) (*AreaResponse, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiURL := fmt.Sprintf("https://api.biteship.com/v1/maps/areas?countries=ID&input=%s", zip)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println(apiURL)
	// Add Authorization header
	req.Header.Set("Authorization", "biteship_live.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiT25lZSBEZXYiLCJ1c2VySWQiOiI2M2UyMGU5OGQxMDIwNjA0Mjc0NDE2ZjkiLCJpYXQiOjE3Njg4OTMzNDR9.z-k8g1upHjO-iIXdj8YaSSPAotQTbne6ILdnOxhONPY")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}

	var result AreaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func InsertArea(area *Area) error {
	query := `
		INSERT INTO zip_bite (
			id, name, country_name, country_code,
			administrative_division_level_1_name, administrative_division_level_1_type,
			administrative_division_level_2_name, administrative_division_level_2_type,
			administrative_division_level_3_name, administrative_division_level_3_type,
			zip
		) VALUES (
			:id, :name, :country_name, :country_code,
			:administrative_division_level_1_name, :administrative_division_level_1_type,
			:administrative_division_level_2_name, :administrative_division_level_2_type,
			:administrative_division_level_3_name, :administrative_division_level_3_type,
			:zip
		)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			country_name = VALUES(country_name),
			country_code = VALUES(country_code),
			administrative_division_level_1_name = VALUES(administrative_division_level_1_name),
			administrative_division_level_1_type = VALUES(administrative_division_level_1_type),
			administrative_division_level_2_name = VALUES(administrative_division_level_2_name),
			administrative_division_level_2_type = VALUES(administrative_division_level_2_type),
			administrative_division_level_3_name = VALUES(administrative_division_level_3_name),
			administrative_division_level_3_type = VALUES(administrative_division_level_3_type),
			zip = VALUES(zip)
	`

	_, err := con.DBX.NamedExec(query, area)
	return err
}
