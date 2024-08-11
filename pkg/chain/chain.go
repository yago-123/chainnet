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
	"time"

	"github.com/sirupsen/logrus"
)

const BlockchainObserver = "blockchain"

type Blockchain struct {
	lastBlockHash []byte
	lastHeight    uint
	headers       map[string]kernel.BlockHeader
	// blockTxsBloomFilter map[string]string

	// syncMutex is used to lock the blockchain while performing syncs with other nodes
	// this avoids collisions when multiple nodes are trying to sync with the local node
	// syncMutex sync.Mutex

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

// OnNodeDiscovered is called when a new node is discovered via the observer pattern.
func (bc *Blockchain) OnNodeDiscovered(peerID peer.ID) {
	bc.logger.Infof("discovered new peer %s", peerID)
	// todo() apply the sync mutex here
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second
	// defer cancel()
	go bc.sync(ctx, peerID)
}

// sync function is in charge of handling all the logic related to node synchronization. The algorithm is as follows:
//  1. Ask the new node for the last header
//  2. Check the height of the remote header
//     2.1 if the height of the remote header is 0 (genesis block not yet created), log and return (we can't sync anything)
//     2.2 if there is no local header, start syncing with the remote node directly (we already know that remote is not empty)
//     2.3 if the height of the remote header is equal, check if the hashes are also equal, if they are not, log it
//     and return
//     2.4 if the height of remote header is smaller than local, check if the header hash is contained in the local
//     chain in case it's not contained, log it and return
//     2.5 if the height of the remote header is bigger, retrieve the headers that are in between and retrieve one by
//     one the blocks to try to add them to the chain. Executed by handleNodeSync auxiliar function
func (bc *Blockchain) sync(ctx context.Context, peerID peer.ID) {
	// ask new peer for last header
	lastHeaderPeer, err := bc.p2pNet.AskLastHeader(ctx, peerID)
	if err != nil {
		bc.logger.Errorf("error asking for last header to %s: %s", peerID.String(), err)
		return
	}

	// in case peer has no headers, log and return (we can't sync anything)
	if lastHeaderPeer.Height == 0 {
		bc.logger.Debugf("peer %s has no headers", peerID.String())
		return
	}

	// in case local node have no headers yet, start syncing with the remote node
	if bc.lastHeight == 0 {
		bc.logger.Debugf("local node has no headers, trying to sync with %s", peerID.String())
		bc.syncWithNoLocalHeader(ctx, lastHeaderPeer, peerID)
		return
	}

	// calculate hash of the last header from the peer
	lastHashPeer, err := util.CalculateBlockHash(lastHeaderPeer, bc.hasher)
	if err != nil {
		bc.logger.Errorf("error calculating block hash from peer %s: %s", peerID.String(), err)
		return
	}

	currentLocalHeight := bc.GetLastHeight() - 1
	// if height is equal, check if the hashes are equal or not
	if currentLocalHeight == lastHeaderPeer.Height {
		if bytes.Equal(bc.GetLastBlockHash(), lastHashPeer) {
			bc.logger.Debugf("out of sync: peer %s has the same height and hash as local node", peerID.String())
		}

		bc.logger.Debugf("peer %s has the same height as local node but different hash", peerID.String())
		return
	}

	// if local height is bigger, check if the header hash is contained in the local chain
	if currentLocalHeight > lastHeaderPeer.Height {
		if _, ok := bc.headers[string(lastHashPeer)]; !ok {
			bc.logger.Debugf("out of sync: peer %s has smaller height and last header is not contained in the local chain", peerID.String())
			return
		}

		bc.logger.Debugf("peer %s has smaller height and last header is contained in the local chain", peerID.String())
		return
	}

	// if local height is smaller than remote, try to add the remaining blocks to the chain
	lastHeaderLocal := bc.headers[string(bc.GetLastBlockHash())]
	bc.syncWithHeaders(ctx, &lastHeaderLocal, lastHeaderPeer, peerID)
}

// syncWithNoLocalHeader called when the local node has no headers (not even genesis block) and needs to be synchronized
// with a remote node that have at least the genesis block
func (bc *Blockchain) syncWithNoLocalHeader(ctx context.Context, remoteHeader *kernel.BlockHeader, peerID peer.ID) {
	// in case local header is empty
}

// syncWithHeaders called when a node with a bigger height than the local is found and the local node needs to sync with it
func (bc *Blockchain) syncWithHeaders(ctx context.Context, localHeader, remoteHeader *kernel.BlockHeader, peerID peer.ID) {
	// in case remote header is bigger than local header
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
