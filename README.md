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
    port: 9100
  - address: "seed-3.chainnet.yago.ninja"
    peer-id: "peerID-3"
    port: 9100

storage-file: "bin/miner-storage"         # File used for persisting the chain status
miner:
  pub-key-reward:                         # Public wallet key encoded in base58, used for receiving mining rewards
    "aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTJVEyUnnaMDSRgUZKJzwFAdWKhSv8HTtbQbecee5xew2DPfqm467oef3KEW7bT54WdDWbvEqEhFv1YT3aPZZVqgKc"
  mining-interval: "10m"                  # Interval between block creation
  adjustment-interval: 6                  # Number of blocks before adjusting the difficulty

chain:
  max-txs-mempool: 10000                  # Maximum number of transactions allowed in the mempool

p2p:
  enabled: true                           # Enable or disable network communication
  identity-path: "identity.pem"           # ECDSA peer private key path in PEM format (leave empty to generate a random identity)
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

### Step 2: Use the Wallet to Send Transactions
You can use this wallet by running the `chainnet-nespv` wallet to send transactions as follows:
```bash
$ ./bin/chainnet-nespv send --config default-config.yaml --address random --amount 1 --fee 10 --wallet-key-path <wallet.pem>
```

### Step 3: Extract the Public Key in Base58 Format
To receive rewards, you'll need to extract the public key from the wallet in `base58` format. This can be done as follows:
```bash
$ openssl ec -in <wallet.pem> -pubout -outform DER 2>/dev/null | base58
```

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

### Remote nodes with Ansible
Running the `chainnet-node` on a remote node:
```bash
$ ansible-playbook -i ansible/hosts.ini ansible/deploy.yml -e "target=node config=../config/examples/seed-node-config.yaml"
```

Running the `chainnet-miner` on a remote node:
```bash
$ ansible-playbook -i ansible/hosts.ini ansible/deploy.yml -e "target=miner config=../config/examples/seed-node-config.yaml"
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


## Architecture
```ascii
┌──────────────────┐                 ┌──────────────────┐
│                  │                 │                  │
│  ChainNet Node   ├────────────────►│  ChainNet Miner  │
│                  │                 │                  │
└──────────────────┘                 └──────────────────┘
```
