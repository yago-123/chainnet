# ChainNet
## Setup
```bash
$ 
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