package main

import (
	health "github.com/Financial-Times/go-fthealth/v1_1"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
)

type routing struct {
	router        *mux.Router
	unfolder      *unfolder
	healthService *healthService
}

func newRouting(unfolder *unfolder, health *healthService) *routing {
	r := routing{
		router:        mux.NewRouter(),
		unfolder:      unfolder,
		healthService: health,
	}

	r.routAdminEndpoints()
	r.routProdEndpoints()

	return &r
}

func (r routing) routAdminEndpoints() {
	r.router.HandleFunc(healthPath, health.Handler(r.healthService.buildHealthCheck())).Methods(http.MethodGet)
	r.router.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(r.healthService.gtgCheck)).Methods(http.MethodGet)
	r.router.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler).Methods(http.MethodGet)
	r.router.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler).Methods(http.MethodGet)
	r.router.HandleFunc(status.PingPath, status.PingHandler).Methods(http.MethodGet)
}

func (r routing) routProdEndpoints() {
	r.router.HandleFunc(unfolderPath, r.unfolder.handle).Methods(http.MethodPut)
}

func (r routing) listenAndServe(port string) {
	err := http.ListenAndServe(":"+port, r.router)
	if err != nil {
		log.Fatalf("Error during ListenAndServe: %v\n", err)
	}
}
