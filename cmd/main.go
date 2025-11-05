package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"oc-data-be-challenge/internal/client"
	"oc-data-be-challenge/internal/collector"
	"oc-data-be-challenge/internal/data/repository"
	httptransport "oc-data-be-challenge/internal/transport/http"
	"oc-data-be-challenge/internal/usecase"
	"oc-data-be-challenge/internal/utils/version"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/go-chi/httplog/v3"
)

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "config file", "./config.json", "Path to configuration file")
}

func main() {
	// Setup Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: httplog.SchemaECS.ReplaceAttr,
	}))
	slog.SetDefault(logger)
	logger.Info("application started", "build_info", version.BuildInfo{}.Info())

	// Parse flags
	flag.Parse()

	// Load Config
	cfg, err := LoadConfigFromFile(cfgPath)
	if err != nil {
		panic(err)
	}
	logger.Info("application config", "config", cfg)

	// Setup InfluxDB Client
	influxdb3Client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:         cfg.InfluxDBClient.Host,
		Token:        cfg.InfluxDBClient.Token,
		Database:     cfg.InfluxDBClient.Database,
		Organization: cfg.InfluxDBClient.Org,
		WriteOptions: &influxdb3.WriteOptions{
			NoSync: true,
		},
	})
	if err != nil {
		panic(err)
	}

	// Setup Data Server Client
	dataServerClient := client.NewDataServerClient(cfg.DataServerClient.Host, nil)

	// Setup Repository
	repo := repository.NewDataPoint(influxdb3Client)

	// Setup UseCase
	uc := usecase.NewDataPointUseCase(repo, dataServerClient)

	// Setup and Start Data Collector
	dataCollector := collector.NewDataServerCollector(uc, time.Millisecond*time.Duration(cfg.DataServerCollector.PollIntervalMs))
	dataCollectorWg := sync.WaitGroup{}
	go func() {
		dataCollectorWg.Add(1)
		defer dataCollectorWg.Done()
		dataCollector.Start()
	}()

	// Setup and Start HTTP server
	handler := httptransport.HandlerWithOptions(httptransport.NewChiServer(uc), httptransport.ChiServerOptions{
		Middlewares: []httptransport.MiddlewareFunc{
			httplog.RequestLogger(logger.With("component", "HTTPServer"), &httplog.Options{
				Level:         slog.LevelInfo,
				Schema:        httplog.SchemaECS,
				RecoverPanics: true,
			}),
		},
	})

	server := &http.Server{
		Addr:    cfg.HTTPServer.Port,
		Handler: handler,
	}

	// Start HTTP server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", "port", cfg.HTTPServer.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		logger.Error("HTTP server error", "error", err)
		panic(err)
	case sig := <-shutdown:
		logger.Info("Shutdown signal received", "signal", sig)

		// Stop the data collector
		logger.Info("Stopping data collector")
		dataCollector.Stop()
		dataCollectorWg.Wait()

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown HTTP server gracefully
		logger.Info("Shutting down HTTP server")
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("HTTP server shutdown error", "error", err)
			_ = server.Close()
		}

		// Close InfluxDB client
		logger.Info("Closing InfluxDB client")
		if err := influxdb3Client.Close(); err != nil {
			logger.Error("InfluxDB client close error", "error", err)
		}

		logger.Info("Application shutdown complete")
	}
}
