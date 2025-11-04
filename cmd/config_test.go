package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfigFromFile_ValidJSON tests loading configuration from a valid JSON file.
func TestLoadConfigFromFile_ValidJSON(t *testing.T) {
	// Create a temporary file with test configuration
	tmpfile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write test configuration to file
	testConfig := Config{
		InfluxDBClient: InfluxDBClientConfig{
			Host:     "http://custom-influxdb:9999",
			Token:    "test-token",
			Database: "production",
			Org:      "myorg",
		},
		DataServerClient: DataServerClientConfig{
			Host: "http://custom-server:9090",
		},
		HTTPServer: HTTPServerConfig{
			Port: ":9000",
		},
	}

	data, err := json.Marshal(testConfig)
	require.NoError(t, err)

	_, err = tmpfile.Write(data)
	require.NoError(t, err)
	tmpfile.Close()

	// Load configuration from file
	cfg, err := LoadConfigFromFile(tmpfile.Name())
	require.NoError(t, err)

	// Verify InfluxDB config
	assert.Equal(t, "http://custom-influxdb:9999", cfg.InfluxDBClient.Host)
	assert.Equal(t, "test-token", cfg.InfluxDBClient.Token)
	assert.Equal(t, "production", cfg.InfluxDBClient.Database)
	assert.Equal(t, "myorg", cfg.InfluxDBClient.Org)

	// Verify Data Server config
	assert.Equal(t, "http://custom-server:9090", cfg.DataServerClient.Host)

	// Verify HTTP Server config
	assert.Equal(t, ":9000", cfg.HTTPServer.Port)
}

// TestLoadConfigFromFile_PartialJSON tests loading configuration with partial JSON (merging with defaults).
func TestLoadConfigFromFile_PartialJSON(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write partial configuration to file (only custom InfluxDB host)
	partialConfig := `{
		"influxdb_client": {
			"host": "http://custom-influxdb:9999"
		}
	}`

	_, err = tmpfile.WriteString(partialConfig)
	require.NoError(t, err)
	tmpfile.Close()

	// Load configuration from file
	cfg, err := LoadConfigFromFile(tmpfile.Name())
	require.NoError(t, err)

	// Verify custom value
	assert.Equal(t, "http://custom-influxdb:9999", cfg.InfluxDBClient.Host)

	// Verify default values were merged
	assert.Equal(t, "dev", cfg.InfluxDBClient.Database)
	assert.Equal(t, "http://localhost:8080", cfg.DataServerClient.Host)
	assert.Equal(t, ":8080", cfg.HTTPServer.Port)
}

// TestLoadConfigFromFile_InvalidFile tests loading configuration from a non-existent file.
func TestLoadConfigFromFile_InvalidFile(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "non-existent-config.json")

	_, err := LoadConfigFromFile(nonExistentPath)
	assert.Error(t, err)
}

// TestLoadConfigFromFile_InvalidJSON tests loading configuration from a file with invalid JSON.
func TestLoadConfigFromFile_InvalidJSON(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write invalid JSON
	_, err = tmpfile.WriteString(`{ invalid json }`)
	require.NoError(t, err)
	tmpfile.Close()

	_, err = LoadConfigFromFile(tmpfile.Name())
	assert.Error(t, err)
}

// TestLoadConfigFromFile_EmptyJSON tests loading configuration from a file with empty JSON object.
func TestLoadConfigFromFile_EmptyJSON(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write empty JSON object
	_, err = tmpfile.WriteString(`{}`)
	require.NoError(t, err)
	tmpfile.Close()

	// Load configuration from file
	cfg, err := LoadConfigFromFile(tmpfile.Name())
	require.NoError(t, err)

	// Verify all defaults are applied
	assert.Equal(t, "http://influxdb3-core:8181", cfg.InfluxDBClient.Host)
	assert.Equal(t, "dev", cfg.InfluxDBClient.Database)
	assert.Equal(t, "http://localhost:8080", cfg.DataServerClient.Host)
	assert.Equal(t, ":8080", cfg.HTTPServer.Port)
}
