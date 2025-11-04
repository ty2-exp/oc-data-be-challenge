package main

import (
	"encoding/json"
	"log/slog"
	"os"

	"dario.cat/mergo"
)

type Config struct {
	// InfluxDBClient holds configuration for InfluxDB client.
	InfluxDBClient InfluxDBClientConfig `json:"influxdb_client,omitempty"`
	// DataServerClient holds configuration for the data server client.
	DataServerClient DataServerClientConfig `json:"data_server_client,omitempty"`
	// HTTPServer holds configuration for the HTTP server.
	HTTPServer HTTPServerConfig `json:"http_server,omitempty"`
	// DataServerCollector holds configuration for the data server collector.
	DataServerCollector DataServerCollectorConfig `json:"data_server_collector,omitempty"`
}

func (o Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("influxdb_client", slog.GroupValue(
			slog.String("host", o.InfluxDBClient.Host),
			slog.String("database", o.InfluxDBClient.Database),
			slog.String("org", o.InfluxDBClient.Org),
		)), // Just show the host
		slog.Any("data_server_client", o.DataServerClient),
		slog.Any("http_server", o.HTTPServer),
		slog.Any("data_server_collector", o.DataServerCollector),
	)
}

// InfluxDBClientConfig holds configuration for InfluxDB's client.
type InfluxDBClientConfig struct {
	// Host is the InfluxDB server host.
	Host string `json:"host,omitempty"`
	// Token is the authentication token for InfluxDB.
	Token string `json:"token,omitempty"`
	// Database is the InfluxDB database name.
	Database string `json:"database,omitempty"`
	// Org is the InfluxDB organization name.
	Org string `json:"org,omitempty"`
}

func DefaultInfluxDBClientConfig() InfluxDBClientConfig {
	return InfluxDBClientConfig{
		Host:     "http://influxdb3-core:8181",
		Database: "dev",
	}
}

// DataServerClientConfig holds configuration for the data server's client.
type DataServerClientConfig struct {
	// Host is the data server host.
	Host string `json:"host,omitempty"`
}

func DefaultDataServerConfig() DataServerClientConfig {
	return DataServerClientConfig{
		Host: "http://localhost:28462",
	}
}

// HTTPServerConfig holds configuration for the HTTP server.
type HTTPServerConfig struct {
	// Port is the port on which the HTTP server listens.
	Port string `json:"port,omitempty"`
}

func DefaultHTTPServerConfig() HTTPServerConfig {
	return HTTPServerConfig{
		Port: ":8080",
	}
}

// DataServerCollectorConfig holds configuration for the data server collector.
type DataServerCollectorConfig struct {
	PollIntervalMs int `json:"poll_interval_ms,omitempty"`
}

func DefaultDataServerCollectorConfig() DataServerCollectorConfig {
	return DataServerCollectorConfig{
		PollIntervalMs: 1000,
	}
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		InfluxDBClient:      DefaultInfluxDBClientConfig(),
		DataServerClient:    DefaultDataServerConfig(),
		HTTPServer:          DefaultHTTPServerConfig(),
		DataServerCollector: DefaultDataServerCollectorConfig(),
	}
}

// LoadConfigFromFile loads configuration from a JSON file and merges it with default values.
func LoadConfigFromFile(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{}
	if err = json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}

	if err = mergo.Merge(&cfg, DefaultConfig()); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
