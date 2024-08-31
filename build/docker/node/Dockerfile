# Use an official base image
FROM ubuntu:24.04

# Set the working directory
WORKDIR /app

# Install necessary packages and dependencies
RUN apt-get update && apt-get install -y \
    curl

# Copy the miner binary into the container
COPY bin/chainnet-node /app/chainnet-node

# Copy the entrypoint script into the container
COPY build/docker/node/entrypoint.sh /app/entrypoint.sh

# Ensure the entrypoint script is executable
RUN chmod +x /app/entrypoint.sh

# Expose the port the miner service runs on
EXPOSE 8080

# Set the entrypoint to the script
ENTRYPOINT ["/app/entrypoint.sh"]

