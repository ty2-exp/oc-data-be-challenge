# OC Backend Challenge Project

## Prerequisites

Before starting, ensure you have the following installed:

- **Docker**: version 28.5.1 or higher
- **Taskfile**: version 3.45.4 or higher
- **Go**: version 1.25 or higher

## Quick Start

Follow these steps to set up and run the project:

### 1. Initialize InfluxDB3 Directories

Set up the required directories with proper permissions:

```bash
task dev:docker-influxdb3-setup-dirs
```

This creates the necessary data, plugins, and UI directories for InfluxDB3 with the correct ownership.

### 2. Start InfluxDB3 Services

Launch the Docker Compose services with the initialization profile:

```bash
docker compose --profile influxdb3-init up
```

This will start InfluxDB3 Core and initialize the admin user.

### 3. Configure the Application

#### Get the Admin Token

Retrieve the admin token from the Docker Compose logs:

```bash
task dev:docker-influxdb3-admin-token
```

Copy the token value from the output.

#### Create Configuration File

Copy the example configuration file:

```bash
cp config-example.json config.json
```

Edit `config.json` and set the token value you obtained from the previous step:

```json
{
  "influxdb_client": {
    "host": "http://127.0.0.1:8181",
    "token": "YOUR_TOKEN_HERE",
    "database": "dev"
  },
  "data_server_collector": {
    "poll_interval_ms": 1000
  }
}
```

### 4. Start the Application

Run the application in development mode with auto-reload:

```bash
task app:dev
```

The application will start on `http://127.0.0.1:8080` (default port).

## Configuration

The application is configured via a `config.json` file. If not provided, default values are used.

### Configuration Options

#### InfluxDB Client (`influxdb_client`)

- **`host`** (string, default: `"http://influxdb3-core:8181"`): InfluxDB server host URL
- **`token`** (string, required): Authentication token for InfluxDB
- **`database`** (string, default: `"dev"`): InfluxDB database name
- **`org`** (string, optional): InfluxDB organization name

#### Data Server Client (`data_server_client`)

- **`host`** (string, default: `"http://localhost:28462"`): Data server host URL from which to collect data points

#### HTTP Server (`http_server`)

- **`port`** (string, default: `":8080"`): Port on which the HTTP server listens

#### Data Server Collector (`data_server_collector`)

- **`poll_interval_ms`** (integer, default: `1000`): Interval in milliseconds at which to poll the data server for new data points

### Example Configuration

```json
{
  "influxdb_client": {
    "host": "http://127.0.0.1:8181",
    "token": "your-admin-token-here",
    "database": "dev",
    "org": "myorg"
  },
  "data_server_client": {
    "host": "http://localhost:28462"
  },
  "http_server": {
    "port": ":8080"
  },
  "data_server_collector": {
    "poll_interval_ms": 1000
  }
}
```

## API Documentation

The HTTP API follows the OpenAPI 3.0 specification. The complete API documentation can be found at:

```
api-spec/tsp-output/schema/openapi.yaml
```

### Available Endpoints

#### Query Data Points

```
GET /data-point
```

Query stored data points with optional time range filters.

**Query Parameters:**
- `start` (optional, duration): Start time for the query range
- `until` (optional, duration): End time for the query range

**Response (200 OK):**
```json
[
  {
    "time": "2023-01-01T00:00:00Z",
    "value": 123.45
  }
]
```

**Error Response (500):**
```json
{
  "message": "Error description"
}
```

## Development

### Available Tasks

View all available tasks:

```bash
task
```

### Common Commands

#### Generate API Server Interface

Regenerate the API server interface from the OpenAPI specification:

```bash
task app:server-swagger-codegen
```

#### Build Application Binary

Build the application binary for local development:

```bash
task app:build-app-local
```

The binary will be created at `dist/bin/out` by default.

Run the binary:

```bash
./dist/bin/out
```

#### Run in Development Mode

Run the application with auto-reload on file changes:

```bash
task app:dev
```

Any changes to `.go` files will trigger an automatic restart.
