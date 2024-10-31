package net

import (
	"context"
	"fmt"

	"github.com/yago-123/chainnet/pkg/encoding"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/yago-123/chainnet/config"
)

const (
	RouterRetrieveAddressTxs     = "/api/v1/address/%s/txs"
	RouterRetrieveAddressUTXOs   = "/api/v1/address/%s/utxos"
	RouterRetrieveAddressBalance = "/api/v1/address/%s/balance"
	RouterSendTx                 = "/api/v1/sendTx"

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
		addr, err := extractAddrInfo(seed.Address, uint(seed.Port), seed.PeerID)
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

func getContentTypeFrom(encoder encoding.Encoding) string {
	switch encoder.Type() {
	case encoding.GobEncodingType:
		return "application/gob"
	case encoding.ProtoEncodingType:
		return "application/x-protobuf"
	default:
		return "application/octet-stream"
	}
}
