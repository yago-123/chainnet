# ChainNet
Running the `chainnet` node: 
```bash
$ ./chainnet
```

Running the cli: 
```bash
$ ./chainnet-cli balance --address <address> 
```

## Architecture
```ascii
┌──────────────────┐                 ┌──────────────────┐
│                  │                 │                  │
│  ChainNet Node   ├────────────────►│  ChainNet Miner  │
│                  │                 │                  │
└──────────────────┘                 └──────────────────┘
```