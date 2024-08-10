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
