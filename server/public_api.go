package main

import (
	"fmt"
	"github.com/bendt-indonesia/env"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/onee-platform/onee-go/cached"
	con "github.com/onee-platform/onee-go/pkg/db/mysql"
	"github.com/onee-platform/onee-go/pkg/wlog"
	"github.com/onee-platform/onee-public-api/handler"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
)

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

	r := mux.NewRouter()
	r.HandleFunc("/couriers", handler.CourierListHandler).Methods("GET")
	r.HandleFunc("/map/provinces", handler.ProvinceListHandler).Methods("GET")
	r.HandleFunc("/map/city", handler.CityListHandler).Methods("GET")
	r.HandleFunc("/map/city/{province_id:[0-9]*}", handler.CityListHandler).Methods("GET")
	r.HandleFunc("/map/kec", handler.KecListHandler).Methods("GET")
	r.HandleFunc("/map/kec/{city_id:[0-9]*}", handler.KecListHandler).Methods("GET")
	r.HandleFunc("/map/kel", handler.KelListHandler).Methods("GET")
	r.HandleFunc("/map/kel/{kec_id:[0-9]*}", handler.KelListHandler).Methods("GET")
	r.HandleFunc("/map/location", handler.LocationHandler).Methods("GET")
	r.HandleFunc("/map/location/{id:[0-9]*}", handler.LocationHandler).Methods("GET")

	port := os.Getenv("PORT")
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // ganti ke domain spesifik di prod
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.ExposedHeaders([]string{"Content-Length"}),
		handlers.AllowCredentials(),
	)

	log.Printf("[%s] Public API Running on localhost:%s", os.Getenv("APP_ENV"), port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), cors(r)))
}
