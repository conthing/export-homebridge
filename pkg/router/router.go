package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/rest", commandHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/v1/version", versionHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/reboot", rebootHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/homebridge/qrcode", qrcodeHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/ping", pingHandler).Methods(http.MethodGet)
	return r
}
