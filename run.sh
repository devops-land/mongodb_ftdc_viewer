#!/bin/bash


# default configuration
export PARALLEL=10
export BATCH_SIZE=50
export INFLUX_DB_DATA_DIRECTORY=/tmp/influxdb_data

# --- Parse CLI args ---
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --input-dir) export INPUT_DIR="$2"; shift 2 ;;
    --influx-url) export INFLUX_URL="$2"; shift 2 ;;
    --influx-token) export INFLUX_TOKEN="$2"; shift 2 ;;
    --influx-org) export INFLUX_ORG="$2"; shift 2 ;;
    --influx-bucket) export INFLUX_BUCKET="$2"; shift 2 ;;
    --parallel) export PARALLEL="$2"; shift 2 ;;
    --batch-size) export BATCH_SIZE="$2"; shift 2 ;;
    --influx-data-directory) export INFLUX_DB_DATA_DIRECTORY="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: $0 --input-dir /path/to/data --influx-url http://influx:8086 \\"
      echo "          --influx-token mytoken --influx-org my-org --influx-bucket mybucket [--parallel 5 --batch-size 100]"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done



if [ -z "$INPUT_DIR" ]; then
  echo "Error: --input-dir not specified."
  exit 1
fi


if [ -z "$BATCH_SIZE" ]; then
  echo "Error: No batch size specified."
  echo "Example: --batch-size 100"
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
    $ENVS docker-compose down
    echo "Docker containers stopped."
    exit 0
}

# Trap Ctrl-C (SIGINT) to run the cleanup function
trap cleanup SIGINT


echo "Checking for image changes and building if needed..."
docker-compose build


# Start Docker containers in detached mode
docker-compose up -d

# Tail the logs of the metrics-processor service
docker attach mongodb_ftdc_viewer-ftdc_exporter-1
