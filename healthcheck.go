package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/service-status-go/gtg"
)

const healthPath = "/__health"

type healthService struct {
	config *healthConfig
	checks []health.Check
}

type healthConfig struct {
	appDesc                    string
	appSystemCode              string
	appName                    string
	port                       string
	writerHealthURI            string
	contentResolverHealthURI   string
	relationsResolverHealthURI string
	producer                   producer.MessageProducer
	client                     *http.Client
}

func newHealthService(config *healthConfig) *healthService {
	service := healthService{config: config}
	service.checks = []health.Check{
		service.writerCheck(),
		service.contentResolverCheck(),
		service.relationsResolverCheck(),
		service.producerCheck(),
	}
	return &service
}

func (service *healthService) buildHealthCheck() health.HC {
	return health.TimedHealthCheck{
		HealthCheck: health.HealthCheck{
			SystemCode:  service.config.appSystemCode,
			Name:        service.config.appName,
			Description: service.config.appDesc,
			Checks:      service.checks,
		},
		Timeout: 10 * time.Second,
	}
}

func (service *healthService) writerCheck() health.Check {
	return health.Check{
		BusinessImpact:   "Content relationships to packages will not be written / updated",
		Name:             "Content collection Neo4j writer health check",
		PanicGuide:       "https://runbooks.in.ft.com/upp-content-collection-rw-neo4j",
		Severity:         2,
		TechnicalSummary: "Checks if the service responsible with writing content collections to Neo4j is healthy",
		Checker:          service.writerChecker,
	}
}

func (service *healthService) contentResolverCheck() health.Check {
	return health.Check{
		BusinessImpact:   "No notifications will be created for the content in unfolded collections",
		Name:             "Document store API health check",
		PanicGuide:       "https://runbooks.in.ft.com/document-store-api",
		Severity:         2,
		TechnicalSummary: "Checks if the service responsible with saving and retrieving content is healthy",
		Checker:          service.contentResolverChecker,
	}
}

func (service *healthService) relationsResolverCheck() health.Check {
	return health.Check{
		BusinessImpact:   "No notifications will be created for the content in unfolded collections",
		Name:             "Relations API health check",
		PanicGuide:       "https://runbooks.in.ft.com/upp-relations-api",
		Severity:         2,
		TechnicalSummary: "Checks if the service responsible with collection relations is healthy",
		Checker:          service.relationsResolverChecker,
	}
}

func (service *healthService) producerCheck() health.Check {
	return health.Check{
		BusinessImpact:   "No notifications will be created for the content in unfolded collections",
		Name:             "Message producer health check",
		PanicGuide:       "https://runbooks.in.ft.com/kafka-proxy",
		Severity:         2,
		TechnicalSummary: "Checks if Kafka can be accessed through http proxy",
		Checker:          service.producerChecker,
	}
}

func (service *healthService) writerChecker() (string, error) {
	return service.httpAvailabilityChecker(service.config.writerHealthURI)
}

func (service *healthService) contentResolverChecker() (string, error) {
	return service.httpAvailabilityChecker(service.config.contentResolverHealthURI)
}

func (service *healthService) relationsResolverChecker() (string, error) {
	return service.httpAvailabilityChecker(service.config.relationsResolverHealthURI)
}

func (service *healthService) producerChecker() (string, error) {
	return service.config.producer.ConnectivityCheck()
}

func (service *healthService) httpAvailabilityChecker(healthURI string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, healthURI, nil)
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

func (service *healthService) GTG() gtg.Status {
	writerCheck := func() gtg.Status {
		return gtgCheck(service.writerChecker)
	}

	contentResolverCheck := func() gtg.Status {
		return gtgCheck(service.contentResolverChecker)
	}

	relationsResolverCheck := func() gtg.Status {
		return gtgCheck(service.relationsResolverChecker)
	}

	producerCheck := func() gtg.Status {
		return gtgCheck(service.producerChecker)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{
		writerCheck,
		contentResolverCheck,
		relationsResolverCheck,
		producerCheck,
	})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}
