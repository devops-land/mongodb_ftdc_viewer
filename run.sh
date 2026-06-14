#!/bin/bash


# default configuration
export PARALLEL=10
export BATCH_SIZE=200
export INFLUX_DB_DATA_DIRECTORY=/tmp/influxdb_data
export INFLUX_ADMIN_PASSWORD="$(uuidgen)"
export INFLUX_API_TOKEN="$(uuidgen)"
export GRAFANA_ADMIN_PASSWORD="$(uuidgen)"
export INFLUX_ORG="org"
export INFLUX_BUCKET="bucket"
export METRICS_INCLUDE_FILE="$(dirname "$0")/metrics_to_get.txt"

# --- Parse CLI args ---
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --input-dir) export INPUT_DIR="$2"; shift 2 ;;
    --influx-url) export INFLUX_URL="$2"; shift 2 ;;
    --influx-org) export INFLUX_ORG="$2"; shift 2 ;;
    --influx-bucket) export INFLUX_BUCKET="$2"; shift 2 ;;
    --parallel) export PARALLEL="$2"; shift 2 ;;
    --batch-size) export BATCH_SIZE="$2"; shift 2 ;;
    --influx-data-directory) export INFLUX_DB_DATA_DIRECTORY="$2"; shift 2 ;;
    --metrics-include-file) export METRICS_INCLUDE_FILE="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: $0 --input-dir /path/to/data [options]"
      echo ""
      echo "Required:"
      echo "  --input-dir DIR              Path to the directory containing FTDC files"
      echo ""
      echo "Optional:"
      echo "  --metrics-include-file FILE  Path to metrics filter list (default: ./metrics_to_get.txt)"
      echo "  --influx-org ORG             InfluxDB organization (default: org)"
      echo "  --influx-bucket BUCKET       InfluxDB bucket (default: bucket)"
      echo "  --influx-data-directory DIR  InfluxDB data directory (default: /tmp/influxdb_data)"
      echo "  --parallel N                 Number of files to process in parallel (default: 10)"
      echo "  --batch-size N               Number of FTDC metrics per batch (default: 200)"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [ -z "$INPUT_DIR" ]; then
  echo "Error: --input-dir not specified. Run '$0 --help' for usage."
  exit 1
fi

if [ ! -d "$INPUT_DIR" ]; then
  echo "Error: --input-dir '$INPUT_DIR' does not exist or is not a directory."
  exit 1
fi

if [ ! -f "$METRICS_INCLUDE_FILE" ]; then
  echo "Error: metrics include file '$METRICS_INCLUDE_FILE' does not exist."
  echo "Use --metrics-include-file to specify a different path."
  exit 1
fi


# List contents of the directory
echo "Contents of '$INPUT_DIR':"
ls -la "$INPUT_DIR"

# Ask for user confirmation
read -p "Do you want to proceed with this directory? (y/n): " answer

case "$answer" in
  y|Y )
    echo "Proceeding..."
    ;;
  * )
    echo "Aborting."
    exit 1
    ;;
esac



# Function to handle Ctrl-C (SIGINT)
cleanup() {
    echo "Stopping Docker containers..."
    $ENVS docker-compose down -v
    echo "Docker containers stopped."
    exit 0
}

# Trap Ctrl-C (SIGINT) to run the cleanup function
trap cleanup SIGINT


echo "Checking for image changes and building if needed..."
docker-compose build

# remove any running container
docker-compose down -v

# Start Docker containers in detached mode
docker-compose up -d

echo "🔐 Credentials:"
echo "--------------------------"
echo "Influx URL               = http://localhost:8086/"
echo "Influx UI User           = admin"
echo "Influx UI Password       = $INFLUX_ADMIN_PASSWORD"
echo "Influx API Token         = $INFLUX_API_TOKEN"
echo "Grafana Dashboard URL    = http://localhost:3001/d/ddnw277huiv40ae/ftdc-dashboard"
echo "Grafana user             = admin"
echo "Grafana Password         = $GRAFANA_ADMIN_PASSWORD"


# Stream logs from the ftdc_exporter service until it exits
docker-compose logs -f ftdc_exporter
