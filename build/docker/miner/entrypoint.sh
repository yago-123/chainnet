#!/bin/bash

# Check if the config file exists
if [ -n "$CONFIG_FILE" ]; then
    # Run the miner with the config file
    /app/chainnet-miner --config "$CONFIG_FILE"
else
    # Run the miner without the config file
    /app/chainnet-miner
fi
