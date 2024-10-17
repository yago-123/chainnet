package blockchain

import (
	"chainnet/config"
	"chainnet/pkg/chain/explorer"
	"chainnet/pkg/consensus"
	"chainnet/pkg/consensus/util"
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/encoding"
	"chainnet/pkg/kernel"
	"chainnet/pkg/mempool"
	"chainnet/pkg/observer"
	"chainnet/pkg/p2p"
	"chainnet/pkg/storage"
	"chainnet/pkg/util/mutex"
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
)

const (
	BlockchainObserver = "blockchain"
	MaxConcurrentSyncs = 1
)

type Blockchain struct {
	lastBlockHash []byte
	lastHeight    uint
	headers       map[string]kernel.BlockHeader
	// blockTxsBloomFilter map[string]string

	// syncMutex is used to lock the chain while performing syncs with other nodes
	// this avoids collisions when multiple nodes are trying to sync with the local node
	syncMutex *mutex.CtxMutex

	hasher  hash.Hashing
	store   storage.Storage
	mempool *mempool.MemPool

	validator consensus.HeavyValidator

	blockSubject observer.BlockSubject

	p2pActive    bool
	p2pNet       *p2p.NodeP2P
	p2pCtx       context.Context
	p2pCancelCtx context.CancelFunc
	p2pEncoder   encoding.Encoding

	logger *logrus.Logger
	cfg    *config.Config
}

func NewBlockchain(
	cfg *config.Config,
	store storage.Storage,
	mempool *mempool.MemPool,
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
	lastHeader, err := store.GetLastHeader()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			// there is no genesis block yet, start chain from scratch
			cfg.Logger.Debugf("no previous block headers found, starting chain from scratch")
		}

		if !errors.Is(err, storage.ErrNotFound) {
			return nil, fmt.Errorf("error retrieving last header: %w", err)
		}
	}

	if err == nil {
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
		headers, err = reconstructHeaders(lastBlockHash, store)
		if err != nil {
			return nil, fmt.Errorf("error reconstructing headers: %w", err)
		}
	}

	return &Blockchain{
		lastBlockHash: lastBlockHash,
		lastHeight:    lastHeight,
		headers:       headers,
		syncMutex:     mutex.NewCtxMutex(MaxConcurrentSyncs),
		hasher:        hasher,
		mempool:       mempool,
		store:         store,
		validator:     validator,
		blockSubject:  subject,
		p2pActive:     false,
		p2pEncoder:    p2pEncoder,
		logger:        cfg.Logger,
		cfg:           cfg,
	}, nil
}

func (bc *Blockchain) InitNetwork(netSubject observer.NetSubject, blockSubject observer.BlockSubject) error {
	var p2pNet *p2p.NodeP2P

	// check if the network is supposed to be enabled
	if !bc.cfg.P2P.Enabled {
		return fmt.Errorf("p2p network is not supposed to be enabled, check configuration")
	}

	// check if the network has been initialized before
	if bc.p2pActive {
		return fmt.Errorf("p2p network already active")
	}

	// create new P2P node
	bc.p2pCtx, bc.p2pCancelCtx = context.WithCancel(context.Background())
	p2pNet, err := p2p.NewNodeP2P(bc.p2pCtx, bc.cfg, netSubject, bc.p2pEncoder, explorer.NewExplorer(bc.store))
	if err != nil {
		return fmt.Errorf("error creating p2p node discovery: %w", err)
	}

	// register the block subject to the network
	// todo() consider moving this in other part
	blockSubject.Register(p2pNet)

	// start the p2p node
	if err = p2pNet.Start(); err != nil {
		return fmt.Errorf("error starting p2p node: %w", err)
	}

	bc.p2pNet = p2pNet
	bc.p2pActive = true

	// todo() relocate this, in general this InitNetwork method seems OFF
	// connect to node seeds
	if err = p2pNet.ConnectToSeeds(); err != nil {
		return fmt.Errorf("error connecting to seeds: %w", err)
	}

	return nil
}

// AddBlock adds a new block to the blockchain. The block is validated before being added to the chain
func (bc *Blockchain) AddBlock(block *kernel.Block) error {
	if err := bc.validator.ValidateBlock(block); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// persist block header, once the header has been persisted the block has been commited to the chain
	if err := bc.store.PersistHeader(block.Hash, *block.Header); err != nil {
		return fmt.Errorf("block header persistence failed: %w", err)
	}

	// STARTING FROM HERE: the code can fail without becoming an issue, the header has been already commited
	// no need to store the block itself, will be commited to store as part of the observer call
	bc.logger.Debugf("added to the chain block %x with height %d", block.Hash, block.Header.Height)

	// update the last block and save the block header
	bc.lastHeight++
	bc.lastBlockHash = block.Hash
	bc.headers[string(block.Hash)] = *block.Header

	// notify observers of a new block added
	bc.blockSubject.NotifyBlockAdded(block)

	return nil
}

// func (bc *Blockchain) generalSync() {
// 	try to sync regularly with other nodes to see if there is some sort of fork (take into account that +6 blocks of
// 	difference is already considered kind of a fork)
// 		1. Ask for the latest header to all peers that the chain is connected to
// 		2. Once all headers are retrieved check if more or less is in sync with the local node (to be developed what that means)
// 		3. If it's not in sync ask all the peers for the header number N (where N is the lowest header that has been
//    	   retrieved, but should be bigger than current height)
//      		3.1 What happens if after doing that the sync still not works? Maybe we should start erasing local headers and
//        			reduce the local height?
// 		4. Once all headers with height N arrive, compare them and choose the header with the hash most popular (the one
//    	   that most nodes have
// 		5. Once you have the header most popular, ask one of the peers that contained that header for all headers
// 		6. Once you have all headers, try to start pulling blocks and adding to the local chain until the header height
//         that was targeted at the step 3.
// 		7. Repeat the process until is considered that the local node is in sync (TO BE DEVELOPED WHAT IN SYNC MEANS)
// }

// syncWithPeer function is in charge of handling all the logic related to node synchronization. Simple algorithm:
//  1. Ask the remote node for the last header
//  2. If the local height is smaller or equal than the remote HEADER height, try to synchronize via headers
//  3. If the local height is bigger than the remote HEADER height, there is nothing to synchronize, just return
func (bc *Blockchain) syncWithPeer(ctx context.Context, peerID peer.ID) error {
	localCurrentHeight := bc.GetLastHeight()

	// ask new peer for last header
	lastHeaderPeer, err := bc.p2pNet.AskLastHeader(ctx, peerID)
	if err != nil {
		return fmt.Errorf("error asking for last header: %w", err)
	}

	// in case current height is smaller or equal than the remote latest header, try to upgrade local node
	if localCurrentHeight <= lastHeaderPeer.Height {
		if err = bc.syncFromHeaders(ctx, peerID, localCurrentHeight); err != nil {
			return fmt.Errorf("error trying to sync with headers from height %d: %w", localCurrentHeight, err)
		}
	}

	if localCurrentHeight > lastHeaderPeer.Height {
		bc.logger.Debugf("local height bigger or equal than remote height for %s: nothing to sync", peerID.String())
	}

	// in case current height is bigger than latest remote header height, there is nothing to sync, just return
	return nil
}

// syncFromHeaders is in charge of synchronizing the local node with the remote node via headers. First retrieve all the
// remote headers and ignore those that are already in the local chain. Then retrieve the block instance for each header
// and try to add it.
//
// If there is some problem while adding the block, return the error (most likely the validator have not accepted the block)
func (bc *Blockchain) syncFromHeaders(ctx context.Context, peerID peer.ID, localCurrentHeight uint) error {
	var block *kernel.Block
	var remoteBlockHash []byte

	// retrieve all headers from the remote node
	remoteHeaders, err := bc.p2pNet.AskAllHeaders(ctx, peerID)
	if err != nil {
		return fmt.Errorf("error asking for all headers: %w", err)
	}

	// sort headers by height
	sort.Slice(remoteHeaders, func(i, j int) bool {
		return remoteHeaders[i].Height < remoteHeaders[j].Height
	})

	// retrieve the block for each header and try to add it to the chain
	for _, header := range remoteHeaders {
		// skip those headers that are already in the local chain
		if localCurrentHeight > header.Height {
			continue
		}

		remoteBlockHash, err = util.CalculateBlockHash(header, bc.hasher)
		if err != nil {
			return fmt.Errorf("error calculating block hash from header: %w", err)
		}

		// retrieve complete block from peer
		block, err = bc.p2pNet.AskSpecificBlock(ctx, peerID, remoteBlockHash)
		if err != nil {
			return fmt.Errorf("error asking for block %x: %w", remoteBlockHash, err)
		}

		// try to add block to the chain, if it fails, log it and finish the sync (blocks are validated insiside AddBlock)
		if err = bc.AddBlock(block); err != nil {
			// todo(): maybe the node should be blamed and black listed?
			return fmt.Errorf("error adding block %x to the chain: %w", remoteBlockHash, err)
		}
	}

	return nil
}

// RetrieveMempoolTxs return an amount of unconfirmed transactions ready to be added to a block
func (bc *Blockchain) RetrieveMempoolTxs(numTxs uint) ([]*kernel.Transaction, uint) {
	return bc.mempool.RetrieveTransactions(numTxs)
}

// ID returns the observer id
func (bc *Blockchain) ID() string {
	return BlockchainObserver
}

// OnNodeDiscovered is called when a new node is discovered via the observer pattern.
func (bc *Blockchain) OnNodeDiscovered(peerID peer.ID) {
	bc.logger.Infof("discovered new peer %s", peerID)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), bc.cfg.P2P.ConnTimeout)
		defer cancel()

		// todo(): revisit this, not sure if makes sense at all this lock type
		bc.syncMutex.Lock(ctx)
		defer bc.syncMutex.Unlock()
		if err := bc.syncWithPeer(ctx, peerID); err != nil {
			bc.logger.Errorf("error syncing with %s: %s", peerID.String(), err)
		}
	}()
}

func (bc *Blockchain) OnUnconfirmedBlockReceived(block kernel.Block) {
	if err := bc.AddBlock(&block); err != nil {
		bc.logger.Errorf("error adding block %x to the chain: %s", block.Hash, err)
	}
}

func (bc *Blockchain) OnUnconfirmedTxReceived(tx kernel.Transaction) {
	// todo() figure how to retrieve fee
	bc.mempool.AppendTransaction(&tx, 0)
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
func reconstructHeaders(lastBlockHash []byte, store storage.Storage) (map[string]kernel.BlockHeader, error) {
	headers := make(map[string]kernel.BlockHeader)

	// todo(): move to explorer instead?
	for len(lastBlockHash) != 0 {
		blockHeader, err := store.RetrieveHeaderByHash(lastBlockHash)
		if err != nil {
			return nil, fmt.Errorf("error retrieving block header %x: %w", lastBlockHash, err)
		}

		headers[string(lastBlockHash)] = *blockHeader
		lastBlockHash = blockHeader.PrevBlockHash
	}

	return headers, nil
}
