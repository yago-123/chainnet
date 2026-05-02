package network

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yago-123/chainnet/config"
)

const (
	RouterV1BetaRetrieveAddressTxs   = "/api/v1beta/addresses/%s/transactions"
	RouterV1BetaRetrieveAddressUTXOs = "/api/v1beta/addresses/%s/utxos"
	RouterV1BetaAddressIsActive      = "/api/v1beta/addresses/%s/activity"
	RouterV1BetaSendTx               = "/api/v1beta/transactions"
	RouterV1BetaTransactionByID      = "/api/v1beta/transactions/%s"
	RouterV1BetaLatestBlock          = "/api/v1beta/blocks/latest"
	RouterV1BetaBlockByHash          = "/api/v1beta/blocks/%s"
	RouterV1BetaLatestHeader         = "/api/v1beta/headers/latest"
	RouterV1BetaHeaderByHeight       = "/api/v1beta/headers/%s"
	RouterV1BetaHeaders              = "/api/v1beta/headers"

	ContentTypeHeader = "Content-Type"
)

func extractAddrInfo(addr string, port uint, id string) (*peer.AddrInfo, error) {
	addrInfo, err := peer.AddrInfoFromString(
		fmt.Sprintf("/dns4/%s/tcp/%d/p2p/%s", addr, port, id),
	)
	if err != nil {
		return &peer.AddrInfo{}, fmt.Errorf("failed to parse multiaddress: %w", err)
	}

	return addrInfo, nil
}

func connectToSeeds(cfg *config.Config, host host.Host) error {
	for _, seed := range cfg.SeedNodes {
		addr, err := extractAddrInfo(seed.Address, uint(seed.Port), seed.PeerID) //nolint:gosec // this int to uint is safe
		if err != nil {
			cfg.Logger.Errorf("failed to extract address info from seed node %s: %v", seed.Address, err)
			continue
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
