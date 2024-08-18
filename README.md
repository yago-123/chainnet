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
$ sysctl -w net.core.rmem_max=7500000
$ sysctl -w net.core.wmem_max=7500000
```
## Build
Running the `chainnet-miner` node:
```bash
$ make
```

## Configuration
Default configuration: 
```yaml
node-seeds: [                       # List of seed nodes
  "seed-1.chainnet.yago.ninja",
  "seed-2.chainnet.yago.ninja",
  "seed-3.chainnet.yago.ninja",
]
storage-file: "bin/miner-storage"   # File used for persisting the chain status
pub-key:                            # Public key in hex format, used for receiving mining rewards
  "aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTG7ZBzTqdDQ...eXF22QHk2JA"
mining-interval: "30s"              # Interval between block creation
p2p-enabled: true                   # Enable or disable network communication
p2p-min-conn: 5                     # Minimum number of connections
p2p-max-conn: 100                   # Maximum number of connections
p2p-conn-timeout: "60s"             # Maximum duration of a connection
p2p-write-timeout: "20s"            # Maximum duration of a write stream
p2p-read-timeout: "20s"             # Maximum duration of a read stream
p2p-buffer-size: 4096               # Read buffer over the network

```

## Deploy seed node with Ansible 
Deploy seed node: 
```bash
$ ansible-playbook -i ansible/hosts.ini ansible/deploy.yml -e "target=node config=../default-config.yaml"
```

Deploy seed node as miner: 
```bash
$ ansible-playbook -i ansible/hosts.ini ansible/deploy.yml -e "target=miner config=../default-config.yaml"
```

## Run
Running the miner:
```bash
$ ./bin/chainnet-miner 
```

Running the node:
```bash

```

Running `nespv` wallet:
```bash

```

## Run in Kubernetes 
Deploy the helm chart: 
```bash
$ helm install chainnet ./helm
```

Uninstall the helm chart: 
```bash
$ helm uninstall chainnet
```

## Architecture
```ascii
┌──────────────────┐                 ┌──────────────────┐
│                  │                 │                  │
│  ChainNet Node   ├────────────────►│  ChainNet Miner  │
│                  │                 │                  │
└──────────────────┘                 └──────────────────┘
```
