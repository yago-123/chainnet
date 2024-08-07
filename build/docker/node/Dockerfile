# Use an official base image
FROM ubuntu:24.04

# Set the working directory
WORKDIR /app

# Copy the miner binary into the container
COPY bin/chainnet-node /app/chainnet

# Install necessary packages and dependencies
RUN apt-get update && apt-get install -y \
    curl

# Create a directory for the database
RUN mkdir -p /data

# Set the database directory as a volume
VOLUME /data

# Expose the port the miner service runs on
EXPOSE 8080

# Run the miner application
# todo: configure the miner to use the /data path
CMD ["/app/chainnet-node"]
