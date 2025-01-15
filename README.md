# ChainNet
## Setup
Install dependencies: 
```bash
$ sudo apt install protobuf-compiler base58
```

Install go packages: 
```bash
$ go install google.golang.org/protobuf/cmd/protoc-gen-go
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

Increase UDP size to [optimize](https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes) P2P communication:  
```bash
$ sudo sysctl -w net.core.rmem_max=7500000
$ sudo sysctl -w net.core.wmem_max=7500000
```
## Configuration
Default configuration:
```yaml
seed-nodes:                               # List of seed nodes
  - address: "seed-1.chainnet.yago.ninja"
    peer-id: "QmNXM4W7om3FdYDutuPjYCgTsazRNWhNn6fNyimf7SUHhR"
    port: 9100
  - address: "seed-2.chainnet.yago.ninja"
    peer-id: "peerID-2"
    port: 8081
  - address: "seed-3.chainnet.yago.ninja"
    peer-id: "peerID-3"
    port: 8082

storage-file: "bin/miner-storage"         # File used for persisting the chain status
miner:
  pub-key-reward:                         # Public wallet key encoded in base58, used for receiving mining rewards
    "aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTJVEyUnnaMDSRgUZKJzwFAdWKhSv8HTtbQbecee5xew2DPfqm467oef3KEW7bT54WdDWbvEqEhFv1YT3aPZZVqgKc"
  mining-interval: "10m"                  # Interval between block creation
  adjustment-interval: 6                  # Number of blocks before adjusting the difficulty

chain:
  max-txs-mempool: 10000                  # Maximum number of transactions allowed in the mempool

prometheus:
  enabled: true                           # Enable or disable prometheus metrics
  port: 9091                              # Port exposed for prometheus metrics
  path: "/metrics"                        # Path for prometheus metrics endpoint

p2p:
  enabled: true                           # Enable or disable network communication
  #identity-path: "identity.pem"          # ECDSA peer private key path in PEM format (leave empty to generate a random identity)
  peer-port: 9100                         # Port used for network communication with other peers
  http-api-port: 8080                     # Port exposed for the router API (required for nespv wallets)
  min-conn: 5                             # Minimum number of connections
  max-conn: 100                           # Maximum number of connections
  conn-timeout: "60s"                     # Maximum duration of a connection
  write-timeout: "20s"                    # Maximum duration of a write stream
  read-timeout: "20s"                     # Maximum duration of a read stream
  buffer-size: 4096                       # Read buffer size over the network

wallet:
  wallet-key-path: priv-key.pem           # ECDSA wallet private key path in PEM format
  server-address: "seed-1.chainnet.yago.ninja"
  server-port: 8080
```
## Build
Building the `chainnet-nespv` wallet:
```bash
$ make nespv 
```

Building the `chainnet-node`: 
```bash
$ make node
```

Building the `chainnet-miner` node:
```bash
$ make miner
```

Building a `chainnet-nespv` wallet:
```bash
$ make nespv 
````

Here's the corrected documentation:

## Creating and Running Wallets

### Step 1: Generate a Private Key
First, create a wallet by generating a private key with `OpenSSL`:
```bash
$ openssl ecparam -name prime256v1 -genkey -noout -out <wallet.pem>
```
This `wallet.pem` file will contain both the private and public keys.

**IMPORTANT:** this code is only compatible with `prime256v1` elliptic curves (so far).

### Step 2: Use the Wallet to Send Transactions
You can use this wallet by running the `chainnet-nespv` wallet to send transactions as follows:
```bash
$ ./bin/chainnet-nespv send --config default-config.yaml --address random --amount 1 --fee 10 --wallet-key-path <wallet.pem>
```

`todo()`: add example with P2PKH payment too  

### Step 3: Extract the Public Key in Base58 Format
To receive rewards, you'll need to extract the public key from the wallet in `base58` format. This can be done as follows:
```bash
$ openssl ec -in <wallet.pem> -pubout -outform DER 2>/dev/null | base58
```

**Note:** You can copy and paste the key obtained for using the wallet directly into the configuration file. The chain uses 
the encoded DER format for keys, as it remains unclear which signing algorithm will be used in the future.

### Step 4: Configure the Miner for Rewards
Once you have the public key, paste it into the `config.yaml` file of the miner to receive mining rewards:

```yaml
miner:
  pub-key-reward:                         # Public wallet key encoded in base58, used for receiving mining rewards
    "aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTJVEyUnnaMDSRgUZKJzwFAdWKhSv8HTtbQbecee5xew2DPfqm467oef3KEW7bT54WdDWbvEqEhFv1YT3aPZZVqgKc"
  mining-interval: "10m"                  # Interval between block creation
  adjustment-interval: 6                  # Number of blocks before adjusting the difficulty
```

This ensures your mining rewards will be sent to the public key generated from your wallet.

## Creating and running nodes and miners 
### Bare metal
Running the `chainnet-node`: 
```bash
$ ./bin/chainnet-node --config default-config.yaml 
```

Running the `chainnet-miner`: 
```bash
$ ./bin/chainnet-miner --config default-config.yaml 
```
### Docker
Running the `chainnet-node`: 
```bash 
$ mkdir /path/to/data
$ cp config/examples/docker-config.yaml /path/to/data/config.yaml
$ docker run -v ./path/to/data:/data -e CONFIG_FILE=/data/config.yaml -p 8080:8080 yagoninja/chainnet-node:latest
```
Running the `chainnet-miner`: 
```bash 
$ mkdir /path/to/data
$ cp config/examples/docker-config.yaml /path/to/data/config.yaml
$ docker run -v ./path/to/data:/data -e CONFIG_FILE=/data/config.yaml -p 8080:8080 yagoninja/chainnet-miner:latest
```

### Remote Nodes with Ansible
To run the `chainnet-node` on a remote node:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini -e @ansible/config/node-seed.yml ansible/playbooks/blockchain.yml
```

To run the `chainnet-miner` on a remote node:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini -e @ansible/config/miner-seed.yml ansible/playbooks/blockchain.yml
```

After the initial chain has been set up, you can also install logging and monitoring with default dashboards. To do this, 
you must first install Grafana:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini ansible/playbooks/grafana.yml
```

Once Grafana is installed, you can configure your domain or access the Grafana instance via `http://localhost:3000` and 
enter the new password (default credentials: `admin`/`admin`). If you need to install HTTPS certificates for the domain, 
you can run `Certbot` using the following playbook and then rerun the Grafana playbook to ensure the reverse proxy 
updates the HTTPS endpoint:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini ansible/playbooks/install-SSL.yml
$ ansible-playbook -i ansible/inventories/seed/hosts.ini ansible/playbooks/grafana.yml
```

Once the chain is running and Grafana is up and accessible, you can install monitoring and/or logging via the following
playbooks:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini ansible/playbooks/monitoring.yml
$ ansible-playbook -i ansible/inventories/seed/hosts.ini ansible/playbooks/logging.yml
```

There is a set of default dashboards available to monitor the chain; however, it may take a few minutes for them to start 
loading real data.

```
If wanna install locally miner: 
$ sudo ansible-playbook -c local -i 127.0.0.1, -e @ansible/config/miner-seed.yml -e "identity_path=/home/yago/.ssh/seed-node-1.pem" ansible/playbooks/blockchain.yml
Install grafana: 
$ ansible-playbook -c local -i 127.0.0.1, ansible/playbooks/grafana.yml
Install monitoring and logging: 
$ ansible-playbook -c local -i 127.0.0.1, ansible/playbooks/monitoring.yml
$ ansible-playbook -c local -i 127.0.0.1, ansible/playbooks/logging.yml
```
Retrieve addresses from `nespv` wallet: 
``` 
./bin/chainnet-nespv addresses --wallet-key-path wallet.pem
```

### Run in Kubernetes 
Deploy the helm chart:
```bash
$ helm install chainnet-release ./helm --set-file configFile=config/examples/kubernetes-config.yaml
```

Uninstall the helm chart:
```bash
$ helm uninstall chainnet
```

## Generating node identities
Generate a ECDSA `secp256r1` private key in PEM format: 
```bash
$ openssl ecparam -name prime256v1 -genkey -noout -out ecdsa-priv-key.pem
```

## Setting up monitoring with Prometheus and Grafana
In order to provide monitoring Grafana, Prometheus and Nginx must be installed. You can do so by running the following 
Ansible playbook: 
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini ansible/playbooks/monitoring.yml
```

In order to install logging: 
```bash 
$ ansible-playbook -i ansible/inventories/seed/hosts.ini ansible/playbooks/logging.yml
```

Once the monitoring stack is installed and you have configured the domain requested to the correct IP, you can access 
the Grafana dashboard at `URL` and admin credentials. 

If you need to enable HTTPS, you can use `Certbot` to generate the keys and certificates for the domain via the following 
playbook: 
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini -l seed-1.chainnet.yago.ninja ansible/playbooks/monitoringTLS.yml -e "domain=dashboard.chainnet.yago.ninja certificate_email=me@yago.ninja"
```

## Architecture
```ascii
┌──────────────────┐                 ┌──────────────────┐
│                  │                 │                  │
│  ChainNet Node   ├────────────────►│  ChainNet Miner  │
│                  │                 │                  │
└──────────────────┘                 └──────────────────┘
```
