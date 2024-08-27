# Use an official base image
FROM ubuntu:24.04

# Set the working directory
WORKDIR /app

# Install necessary packages and dependencies
RUN apt-get update && apt-get install -y \
    curl

# Create a directory for the database and configuration
RUN mkdir -p /data

# Set the /data directory as a volume
VOLUME /data

# Copy the miner binary into the container
COPY bin/chainnet-miner /app/chainnet-miner

# Copy the entrypoint script into the container
COPY build/docker/miner/entrypoint.sh /app/entrypoint.sh

# Ensure the entrypoint script is executable
RUN chmod +x /app/entrypoint.sh

# Expose the port the miner service runs on
EXPOSE 8080

# Set a default value for the CONFIG_FILE environment variable
ENV CONFIG_FILE=/data/config

# Set the entrypoint to the script
ENTRYPOINT ["/app/entrypoint.sh"]

