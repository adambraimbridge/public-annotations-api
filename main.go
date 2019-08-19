package main

import (
	"net/http"
	"os"

	"fmt"
	"strconv"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	//"github.com/Financial-Times/http-handlers-go/httphandlers"

	//"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/upp-micro-kit"

	"github.com/Financial-Times/public-annotations-api/annotations"
	status "github.com/Financial-Times/service-status-go/httphandlers"

	"github.com/Financial-Times/go-logger"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rcrowley/go-metrics"
)

const (
	appDescription = "A public RESTful API for accessing Annotations in neo4j"
)

func main() {
	app := cli.App("public-annotations-api", appDescription)
	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://localhost:7474/db/data",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL"})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})
	env := app.String(cli.StringOpt{
		Name:  "env",
		Value: "local",
		Desc:  "environment this app is running in",
	})
	cacheDuration := app.String(cli.StringOpt{
		Name:   "cache-duration",
		Value:  "30s",
		Desc:   "Duration Get requests should be cached for. e.g. 2h45m would set the max-age value to '7440' seconds",
		EnvVar: "CACHE_DURATION",
	})
	//logLevel := app.String(cli.StringOpt{
	//	Name:   "log-level",
	//	Value:  "info",
	//	Desc:   "Log level for the service",
	//	EnvVar: "LOG_LEVEL",
	//})

	//lvl, err := log.ParseLevel(*logLevel)
	//if err != nil {
	//	lvl = log.InfoLevel
	//}
	//log.SetLevel(lvl)
	//
	log := logger.NewInfoLogger("public-annotations-api")

	app.Action = func() {
		log.Infof("public-annotations-api will listen on port: %s, connecting to: %s", *port, *neoURL)
		runServer(*neoURL, *port, *cacheDuration, *env, log)
	}

	log.Infof("Application started with args %s", os.Args)
	err := app.Run(os.Args)
	if err != nil {
		log.WithError(err).Error("public-annotations-api could not start!")
		return
	}
}

func runServer(neoURL string, port string, cacheDuration string, env string, log *logger.UPPLogger) {
	if duration, durationErr := time.ParseDuration(cacheDuration); durationErr != nil {
		log.Fatalf("Failed to parse cache duration string, %v", durationErr)
	} else {
		annotations.CacheControlHeader = fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))
	}

	conf := neoutils.ConnectionConfig{
		BatchSize:     1024,
		Transactional: false,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
			},
			Timeout: 1 * time.Minute,
		},
		BackgroundConnect: true,
	}
	db, err := neoutils.Connect(neoURL, &conf)

	if err != nil {
		log.Fatalf("Error connecting to neo4j %s", err)
	}

	annotations.AnnotationsDriver = annotations.NewCypherDriver(db, env)
	routeRequests(port, log)
}

func routeRequests(port string, log *logger.UPPLogger) {

	// Standard endpoints
	healthCheck := fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  "annotationsapi",
			Name:        "public-annotations-api",
			Description: appDescription,
			Checks: []fthealth.Check{
				annotations.HealthCheck(),
			},
		},
		Timeout: 10 * time.Second,
	}
	http.HandleFunc("/__health", fthealth.Handler(healthCheck))
	http.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(annotations.GoodToGo))
	http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	// API specific endpoints
	servicesRouter := mux.NewRouter()

	servicesRouter.HandleFunc("/content/{uuid}/annotations", annotations.GetAnnotations).Methods("GET")
	servicesRouter.HandleFunc("/content/{uuid}/annotations", annotations.MethodNotAllowedHandler)

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = upp_micro_kit.TransactionAwareRequestLoggingHandler(log, monitoringRouter)
	monitoringRouter = upp_micro_kit.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	http.Handle("/", monitoringRouter)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
