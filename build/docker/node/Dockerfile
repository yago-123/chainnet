# Use an official base image
FROM ubuntu:24.04

# Set the working directory
WORKDIR /app

# Install necessary packages and dependencies
RUN apt-get update && apt-get install -y \
    curl

# Create a directory for the persistence of data
RUN mkdir -p /data

# Set the database directory as a volume
VOLUME /data

# Expose the port the node service runs on
EXPOSE 8080

# Copy the node binary into the container
COPY bin/chainnet-node /app/chainnet

# Run the miner application
# todo: configure the node to use the /data path
CMD ["/app/chainnet-node"]
