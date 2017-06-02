package main

import (
	"errors"
	"fmt"
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
	"net/http"
)

const healthPath = "/__health"

type healthService struct {
	config *healthConfig
	checks []health.Check
}

type healthConfig struct {
	client          *http.Client
	appSystemCode   string
	appName         string
	appDesc         string
	port            string
	writerHealthUri string
}

func newHealthService(config *healthConfig) *healthService {
	service := &healthService{config: config}
	service.checks = []health.Check{
		service.writerCheck(),
	}
	return service
}

func (service *healthService) buildHealthCheck() health.HealthCheck {
	return health.HealthCheck{
		SystemCode:  service.config.appSystemCode,
		Name:        service.config.appName,
		Description: service.config.appDesc,
		Checks:      service.checks,
	}
}

func (service *healthService) writerCheck() health.Check {
	return health.Check{
		BusinessImpact:   "Article relationships to packages will not be written / updated",
		Name:             "Content collection Neo4j writer health check",
		PanicGuide:       "https://dewey.ft.com/upp-content-collection-rw-neo4j.html",
		Severity:         1,
		TechnicalSummary: "Checks if the service responsible with writing content collections to Neo4j is healthy",
		Checker: func() (string, error) {
			return service.httpAvailabilityChecker(service.config.writerHealthUri)
		},
	}
}

func (service *healthService) httpAvailabilityChecker(healthUri string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, healthUri, nil)
	if err != nil {
		msg := fmt.Sprintf("Error while creating http health check request: %v", err)
		return msg, errors.New(msg)
	}

	resp, err := service.config.client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("Error contacting the service: %v", err)
		return msg, errors.New(msg)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Service did not responde with OK. Status was %v", resp.Status)
		return msg, errors.New(msg)
	}

	return "OK", nil
}

func (service *healthService) gtgCheck() gtg.Status {
	for _, check := range service.checks {
		if _, err := check.Checker(); err != nil {
			return gtg.Status{GoodToGo: false, Message: err.Error()}
		}
	}
	return gtg.Status{GoodToGo: true}
}
