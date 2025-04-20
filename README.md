![Alt text](https://github.com/user-attachments/assets/3f5f540a-4f28-402f-818b-c44e7bb8ed3e)

# Chainnet: distributed chain
**Access live Chainnet metrics from the main seed node at  [dashboard.chainnet.yago.ninja/list](https://dashboard.chainnet.yago.ninja/list).**

## Features
- [x] Decentralized peer-to-peer connectivity and synchronization
- [x] Node discovery via seed nodes and Kademlia DHT
- [x] Stack based RPN interpreter for payments 
  - [x] P2PK (Pay to Public Key)
  - [x] P2PKH (Pay to Public Key Hash)
  - [ ] P2SH (Pay to Script Hash)
- [x] Block rewards for mining
- [x] Transaction fees 
- [ ] Wallets
  - [x] ECDSA key generation
  - [x] ECDSA signing
  - [x] Hierarchical Deterministic (HD) Wallet
  - [ ] Mnemonic generation
- [x] Block and transaction validation
- [x] Block and transaction propagation
- [x] Mempool holding validated, unconfirmed transactions
- [x] UTXO set for tracking all unspent outputs and balances
- [ ] Block conflict resolution during synchronization
- [ ] Bloom filter for efficient lightweight client support

## Configuration
Default configuration:
```yaml
seed-nodes:                               # List of seed nodes
  - address: "seed-1.chainnet.yago.ninja"
    peer-id: "QmVQ8bj9KPfTiN23vX7sbqn4oXjTSycfULL4oApAZccWL5"
    port: 9100
#  ... more seed nodes ...

storage-file: "bin/miner-storage"         # File used for persisting the chain status
miner:
  pub-key-reward:                         # Public wallet key encoded in base58, used for receiving mining rewards
    "aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTK2r1ViPYeJCMAcSHrt4AEkBouG5vmbAjKMGnZ1RyjP3bPTUhJrRXfEnD3CEhB7Rumao463ayeiU2jbRhjsygwqFp"
  mining-interval: "10m"                  # Interval between block creation
  adjustment-interval: 6                  # Number of blocks before adjusting the difficulty

chain:
  max-txs-mempool: 10000                  # Maximum number of transactions allowed in the mempool

prometheus:
  enabled: true                           # Enable or disable prometheus metrics
  port: 9091                              # Port exposed for prometheus metrics
  libp2p-port: 9099                       # Port exposed for prometheus core libp2p metrics
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

## Building from source
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
$ ./bin/chainnet-nespv send            \
          --config default-config.yaml \
          --address random             \
          --amount 23.5 --fee 0.001    \
          --wallet-key-path <wallet.pem>
```

By default transactions use `P2PK` payments, if you want to use `P2PKH` payments you can use the `--pay-type` flag:
```bash
$ ./bin/chainnet-nespv send            \
          --config default-config.yaml \
          --address random             \
          --amount 23.5 --fee 0.001    \
          --pay-type P2PKH             \ 
          --wallet-key-path <wallet.pem>
```

You can use the `addresses` subcommand to list the addresses attached to this wallet:
```bash
$ ./bin/chainnet-nespv addresses \
           --wallet-key-path <wallet.pem>
```

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

### Remote Nodes with Ansible
To run the `chainnet-node` on a remote node:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini \
                   -e @ansible/config/node-seed.yml      \
                   ansible/playbooks/blockchain.yml
```

To run the `chainnet-miner` on a remote node:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini \
                   -e @ansible/config/miner-seed.yml     \
                   ansible/playbooks/blockchain.yml
```

After the initial chain has been set up, you can also install logging and monitoring with default dashboards. To do this,
you must first install Grafana:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini \
                   ansible/playbooks/visualization.yml
```

Once Grafana is installed, you can configure your domain or access the Grafana instance via `http://localhost:3000` and
enter the new password (default credentials: `admin`/`admin`). If you need to install HTTPS certificates for the domain,
you can run `Certbot` using the following playbook and then rerun the Grafana playbook to ensure the reverse proxy
updates the HTTPS endpoint:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini \
                   ansible/playbooks/install-SSL.yml
$ ansible-playbook -i ansible/inventories/seed/hosts.ini \
                   ansible/playbooks/visualization.yml
```

Once the chain is running and Grafana is up and accessible, you can install monitoring and/or logging via the following
playbooks:
```bash
$ ansible-playbook -i ansible/inventories/seed/hosts.ini \
                   ansible/playbooks/monitoring.yml
$ ansible-playbook -i ansible/inventories/seed/hosts.ini \
                   ansible/playbooks/logging.yml
```

There is a set of default dashboards available to monitor the chain; however, it may take a few minutes for them to start
loading real data.

### Run in Docker (currently not mantained)
Running the `chainnet-node`:
```bash 
$ mkdir /path/to/data
$ cp config/examples/docker-config.yaml /path/to/data/config.yaml
$ docker run -v ./path/to/data:/data           \
             -e CONFIG_FILE=/data/config.yaml  \
             -p 8080:8080                      \
             yagoninja/chainnet-node:latest
```
Running the `chainnet-miner`:
```bash 
$ mkdir /path/to/data
$ cp config/examples/docker-config.yaml /path/to/data/config.yaml
$ docker run -v ./path/to/data:/data            \
             -e CONFIG_FILE=/data/config.yaml   \
             -p 8080:8080                       \
             yagoninja/chainnet-miner:latest
```

### Run in Kubernetes (currently not mantained)
Deploy the helm chart:
```bash
$ helm install chainnet-release ./helm \
  --set-file configFile=config/examples/kubernetes-config.yaml
```

Uninstall the helm chart:
```bash
$ helm uninstall chainnet
```

## Generating Node Identities
To authenticate nodes in P2P connections, you can generate a node identity. Start by generating an ECDSA `secp256r1` private key in PEM format:
```bash  
$ openssl ecparam -name prime256v1 -genkey -noout -out ecdsa-priv-key.pem  
```  

Next, reference the identity path in the configuration file:
```yaml  
p2p:
  enabled: true                           # Enable or disable network communication  
  identity-path: "ecdsa-priv-key.pem"     # Path to the ECDSA peer private key in PEM format (leave empty to generate a random identity)  
```  

Note that this identity can also be used to authenticate the seed nodes via the `peer-id` field:
```yaml  
seed-nodes:                               # List of seed nodes  
  - address: "seed-1.chainnet.yago.ninja"
    peer-id: "QmNXM4W7om3FdYDutuPjYCgTsazRNWhNn6fNyimf7SUHhR"
    port: 9100  
```  