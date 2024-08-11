package blockchain

import (
	"bytes"
	"chainnet/config"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/chain/observer"
	"chainnet/pkg/chain/p2p"
	"chainnet/pkg/consensus"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/storage"
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sirupsen/logrus"
)

const BlockchainObserver = "blockchain"

type Blockchain struct {
	lastBlockHash []byte
	lastHeight    uint
	headers       map[string]kernel.BlockHeader
	// blockTxsBloomFilter map[string]string

	hasher    hash.Hashing
	storage   storage.Storage
	validator consensus.HeavyValidator

	blockSubject observer.BlockSubject

	p2pActive    bool
	p2pNet       *p2p.NodeP2P
	p2pCtx       context.Context
	p2pCancelCtx context.CancelFunc

	p2pEncoder encoding.Encoding

	logger *logrus.Logger
	cfg    *config.Config
}

func NewBlockchain(
	cfg *config.Config,
	storage storage.Storage,
	hasher hash.Hashing,
	validator consensus.HeavyValidator,
	subject observer.BlockSubject,
	p2pEncoder encoding.Encoding,
) (*Blockchain, error) {
	var err error
	var lastHeight uint
	var lastBlockHash []byte

	headers := make(map[string]kernel.BlockHeader)

	// retrieve the last header stored
	lastHeader, err := storage.GetLastHeader()
	if err != nil {
		return nil, fmt.Errorf("error retrieving last header: %w", err)
	}

	if lastHeader.IsEmpty() {
		// there is no genesis block yet, start chain from scratch
		cfg.Logger.Debugf("no previous block headers found, starting chain from scratch")
	}

	if !lastHeader.IsEmpty() {
		// if exists a last header, sync the actual status of the chain
		// specify the current height
		lastHeight = lastHeader.Height + 1

		// get the last block hash by hashing the latest block header
		lastBlockHash, err = util.CalculateBlockHash(lastHeader, hasher)
		if err != nil {
			return nil, fmt.Errorf("error retrieving last block hash: %w", err)
		}

		cfg.Logger.Debugf("recovering chain with last hash: %x", lastBlockHash)

		// reload the headers into memory
		headers, err = reconstructHeaders(lastBlockHash, storage)
		if err != nil {
			return nil, fmt.Errorf("error reconstructing headers: %w", err)
		}
	}

	return &Blockchain{
		lastBlockHash: lastBlockHash,
		lastHeight:    lastHeight,
		headers:       headers,
		hasher:        hasher,
		storage:       storage,
		validator:     validator,
		blockSubject:  subject,
		p2pActive:     false,
		p2pEncoder:    p2pEncoder,
		logger:        cfg.Logger,
		cfg:           cfg,
	}, nil
}

func (bc *Blockchain) InitNetwork() error {
	var p2pNet *p2p.NodeP2P

	// check if the network is supposed to be enabled
	if !bc.cfg.P2PEnabled {
		return fmt.Errorf("p2p network is not supposed to be enabled, check configuration")
	}

	// check if the network has been initialized before
	if bc.p2pActive {
		return fmt.Errorf("p2p network already active")
	}

	// create a new blockchain observer that will react to network events
	netSubject := observer.NewNetSubject()
	netSubject.Register(bc)

	// create new P2P node
	bc.p2pCtx, bc.p2pCancelCtx = context.WithCancel(context.Background())
	p2pNet, err := p2p.NewP2PNode(bc.p2pCtx, bc.cfg, netSubject, bc.p2pEncoder, explorer.NewExplorer(bc.storage))
	if err != nil {
		return fmt.Errorf("error creating p2p node discovery: %w", err)
	}

	// initialize network handlers
	p2pNet.InitHandlers()

	// start the p2p node
	if err = p2pNet.Start(); err != nil {
		return fmt.Errorf("error starting p2p node: %w", err)
	}

	bc.p2pNet = p2pNet
	bc.p2pActive = true

	return nil
}

func (bc *Blockchain) AddBlock(block *kernel.Block) error {
	if err := bc.validator.ValidateBlock(block); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// persist block header, once the header has been persisted the block has been commited to the chain
	if err := bc.storage.PersistHeader(block.Hash, *block.Header); err != nil {
		return fmt.Errorf("block header persistence failed: %w", err)
	}

	// STARTING FROM HERE: the code can fail without becoming an issue, the header has been already commited
	// no need to store the block itself, will be commited to storage as part of the observer call

	// update the last block and save the block header
	bc.lastHeight++
	bc.lastBlockHash = block.Hash
	bc.headers[string(block.Hash)] = *block.Header

	// notify observers of a new block added
	bc.blockSubject.NotifyBlockAdded(block)

	return nil
}

func (bc *Blockchain) Sync() {
	// if there is latest block (node has been started before)
	// - retrieve latest block hash
	// - compare height and hash
	// - decide with the other peers next steps (download next headers or execute some conflict resolution)

	// if there is not latest block (node is new)
	// - ask for headers
	// - download & verify headers
	// - start IBD (Initial Block Download): download block from each header
	// 		- validate each block
}

// ID returns the observer id
func (bc *Blockchain) ID() string {
	return BlockchainObserver
}

// OnNodeDiscovered is called when a new node is discovered via the observer pattern
func (bc *Blockchain) OnNodeDiscovered(peerID peer.ID) {
	var block *kernel.Block
	bc.logger.Infof("discovered new peer %s", peerID)

	// ask new peer for last header
	lastHeaderPeer, err := bc.p2pNet.AskLastHeader(peerID)
	if err != nil {
		bc.logger.Errorf("error asking for last header to %s: %s", peerID.String(), err)
		return
	}

	// retrieve local last header (todo(): consider caching this last header?)
	lastHeaderLocal, ok := bc.headers[string(bc.lastBlockHash)]
	if !ok {
		if len(bc.lastBlockHash) == 0 {
			bc.logger.Infof("last block hash not found, chain may not have genesis block yet")
			return
		}

		bc.logger.Errorf("last block hash %s not found in headers", bc.lastBlockHash)
	}

	// calculate hash of the peer header
	headerHashPeer, err := util.CalculateBlockHash(lastHeaderPeer, bc.hasher)
	if err != nil {
		bc.logger.Errorf("error calculating header hash for peer %s: %s", peerID.String(), err)
		return
	}

	// same height, may be in sync
	if lastHeaderLocal.Height == lastHeaderPeer.Height {
		// in case last hash and height are the same, the peer is in sync
		if bytes.Equal(bc.lastBlockHash, headerHashPeer) {
			bc.logger.Debugf("peer %s is in sync with local chain", peerID)
			return
		}

		// todo() cover case when the peer have same height but different hash

		return
	}

	// local chain have less blocks, ask for headers to try to "escalate"
	if lastHeaderLocal.Height < lastHeaderPeer.Height {
		heightDiff := lastHeaderPeer.Height - lastHeaderLocal.Height
		// in case the peer have only one more block than local chain and previous block hash matches
		// ask for the block and try to add it to the chain
		if heightDiff == 1 && bytes.Equal(lastHeaderPeer.PrevBlockHash, bc.lastBlockHash) {
			// todo() introduce this into an auxiliar function

			// ask for the whole block to the peer
			block, err = bc.p2pNet.AskSpecificBlock(peerID, headerHashPeer)
			if err != nil {
				bc.logger.Errorf("error asking for block %x to %s: %s", headerHashPeer, peerID.String(), err)
				return
			}

			// try to add the block to the chain
			if err = bc.AddBlock(block); err != nil {
				// todo() maybe should be added some trust score to peers, if the peer returns a wrong block
				// todo() consider some sort of blacklist or reporting mechanism (via gossip maybe?)
				bc.logger.Errorf("error adding block %x from peer %s, verification failed: %s", headerHashPeer, peerID.String(), err)
				return
			}
		}

		// todo() take into account if height is bigger than 6 blocks
		// todo() ask other peers if they have this addition too
		return
	}

	// in case local chain have more blocks, do nothing the other node will try to sync with local chain
}

// GetLastBlockHash returns the latest block hash
func (bc *Blockchain) GetLastBlockHash() []byte {
	return bc.lastBlockHash
}

// GetLastHeight returns the latest block height
func (bc *Blockchain) GetLastHeight() uint {
	return bc.lastHeight
}

// reconstructHeaders
func reconstructHeaders(lastBlockHash []byte, storage storage.Storage) (map[string]kernel.BlockHeader, error) {
	headers := make(map[string]kernel.BlockHeader)

	// todo(): move to explorer instead?
	for len(lastBlockHash) != 0 {
		blockHeader, err := storage.RetrieveHeaderByHash(lastBlockHash)
		if err != nil {
			return nil, fmt.Errorf("error retrieving block header %x: %w", lastBlockHash, err)
		}

		headers[string(lastBlockHash)] = *blockHeader
		lastBlockHash = blockHeader.PrevBlockHash
	}

	return headers, nil
}
