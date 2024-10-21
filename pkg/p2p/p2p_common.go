package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yago-123/chainnet/config"
)

const (
	RouterAddressTxs     = "/address/%s/transactions"
	RouterAddressUTXOs   = "/address/%s/utxos"
	RouterAddressBalance = "/address/%s/balance"
)

func connectToSeeds(cfg *config.Config, host host.Host) error {
	for _, seed := range cfg.SeedNodes {
		addr, err := peer.AddrInfoFromString(
			fmt.Sprintf("/dns4/%s/tcp/%d/p2p/%s", seed.Address, seed.Port, seed.PeerID),
		)
		if err != nil {
			return fmt.Errorf("failed to parse multiaddress: %w", err)
		}

		// todo(): provide this context via argument
		err = host.Connect(context.Background(), *addr)
		if err != nil {
			cfg.Logger.Errorf("failed to connect to seed node %s: %v", addr, err)
			continue
		}

		cfg.Logger.Infof("connected to seed node %s", addr.ID.String())
	}

	return nil
}
