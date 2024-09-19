# ChainNet
## Setup
Install dependencies: 
```bash
$ sudo apt install protobuf-compiler 
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
pub-key:                                  # Public wallet key encoded in base58, used for receiving mining rewards
  "aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTG7ZBzTqdDQvpbDVh5j5yCpKYU6MVZ35PW9KegkuX1JZDLHdkaTAbKXwfx4Pjy2At82Dda9ujs8d5ReXF22QHk2JA"
mining-interval: "10m"                    # Interval between block creation

p2p:
  enabled: true                           # Enable or disable network communication
  identity:
    priv-key-path: "ecdsa-priv-key.pem"   # ECDSA peer private key path in PEM format (leave empty to generate a random identity)

  peer-port: 9100                         # Port used for network communication with other peers
  min-conn: 5                             # Minimum number of connections
  max-conn: 100                           # Maximum number of connections
  conn-timeout: "60s"                     # Maximum duration of a connection
  write-timeout: "20s"                    # Maximum duration of a write stream
  read-timeout: "20s"                     # Maximum duration of a read stream
  buffer-size: 4096                       # Read buffer size over the network
```
## Build
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

## Running
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

Running the `chainnet-nespv` wallet: 
```bash

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
