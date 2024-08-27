#!/bin/bash

# Define the configuration file path using an environment variable or default to /data/config
CONFIG_FILE="${CONFIG_FILE:-/data/config.yaml}"

# Check if the config file exists
if [ -f "$CONFIG_FILE" ]; then
    # Run the miner with the config file
    /app/chainnet-node --config "$CONFIG_FILE"
else
    # Run the miner without the config file
    /app/chainnet-node
fi
