package blockchain

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/yago-123/chainnet/pkg/common"

	cerror "github.com/yago-123/chainnet/pkg/errs"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yago-123/chainnet/pkg/monitor"
	"github.com/yago-123/chainnet/pkg/utxoset"

	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/pkg/chain/explorer"
	"github.com/yago-123/chainnet/pkg/consensus"
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"github.com/yago-123/chainnet/pkg/encoding"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/mempool"
	"github.com/yago-123/chainnet/pkg/network"
	"github.com/yago-123/chainnet/pkg/observer"
	"github.com/yago-123/chainnet/pkg/storage"
	"github.com/yago-123/chainnet/pkg/util"
	"github.com/yago-123/chainnet/pkg/util/mutex"

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

	// todo() may be smarter to have the target as a field of the blockchain (saving the previous header interval
	// todo() too), but generalSync must be implemented before that to ensure consistency

	// syncMutex is used to lock the chain while performing syncs with other nodes
	// this avoids collisions when multiple nodes are trying to sync with the local node
	syncMutex *mutex.CtxMutex

	hasher  hash.Hashing
	store   storage.Storage
	mempool *mempool.MemPool
	utxoSet *utxoset.UTXOSet

	validator consensus.HeavyValidator

	blockSubject observer.ChainSubject

	p2pActive    bool
	p2pNet       *network.NodeP2P
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
	utxoSet *utxoset.UTXOSet,
	hasher hash.Hashing,
	validator consensus.HeavyValidator,
	subject observer.ChainSubject,
	p2pEncoder encoding.Encoding,
) (*Blockchain, error) {
	var err error
	var lastHeight uint
	var lastBlockHash []byte

	headers := make(map[string]kernel.BlockHeader)

	// retrieve the last header stored
	lastHeader, err := store.GetLastHeader()
	if err != nil {
		if errors.Is(err, cerror.ErrStorageElementNotFound) {
			// there is no genesis block yet, start chain from scratch
			cfg.Logger.Debugf("no previous block headers found, starting chain from scratch")
		}

		if !errors.Is(err, cerror.ErrStorageElementNotFound) {
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
		if err = reconstructState(store, utxoSet, headers, lastBlockHash); err != nil {
			return nil, fmt.Errorf("error reconstructing chain state: %w", err)
		}
	}

	return &Blockchain{
		lastBlockHash: lastBlockHash,
		lastHeight:    lastHeight,
		headers:       headers,
		syncMutex:     mutex.NewCtxMutex(MaxConcurrentSyncs),
		hasher:        hasher,
		mempool:       mempool,
		utxoSet:       utxoSet,
		store:         store,
		validator:     validator,
		blockSubject:  subject,
		p2pActive:     false,
		p2pEncoder:    p2pEncoder,
		logger:        cfg.Logger,
		cfg:           cfg,
	}, nil
}

func (bc *Blockchain) InitNetwork(netSubject observer.NetSubject) (*network.NodeP2P, error) {
	var p2pNet *network.NodeP2P

	// check if the network is supposed to be enabled
	if !bc.cfg.P2P.Enabled {
		return nil, fmt.Errorf("p2p network is not supposed to be enabled, check configuration")
	}

	// check if the network has been initialized before
	if bc.p2pActive {
		return nil, fmt.Errorf("p2p network already active")
	}

	// create new P2P node
	chainExplorer := explorer.NewChainExplorer(bc.store, bc.hasher)
	mempoolExplorer := mempool.NewMemPoolExplorer(bc.mempool)
	bc.p2pCtx, bc.p2pCancelCtx = context.WithCancel(context.Background())

	p2pNet, err := network.NewNodeP2P(bc.p2pCtx, bc.cfg, netSubject, bc.p2pEncoder, chainExplorer, mempoolExplorer)
	if err != nil {
		return nil, fmt.Errorf("error creating p2p node discovery: %w", err)
	}

	// start the p2p node
	if err = p2pNet.Start(); err != nil {
		return nil, fmt.Errorf("error starting p2p node: %w", err)
	}

	bc.p2pNet = p2pNet
	bc.p2pActive = true

	// todo() relocate this, in general this InitNetwork method seems OFF
	// connect to node seeds
	if err = p2pNet.ConnectToSeeds(); err != nil {
		return nil, fmt.Errorf("error connecting to seeds: %w", err)
	}

	return p2pNet, nil
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

// AddTransaction adds a new transaction to the mempool. The transaction is validated before being added to the mempool
func (bc *Blockchain) AddTransaction(tx *kernel.Transaction) error {
	// make sure that the tx uses proper UTXOs and contains valid signatures
	if err := bc.validator.ValidateTx(tx); err != nil {
		return fmt.Errorf("error validating transaction %x: %w", tx.ID, err)
	}

	// calculate the transaction fee
	fee, err := bc.calculateTxFee(tx)
	if err != nil {
		return fmt.Errorf("error calculating transaction fee for %x: %w", tx.ID, err)
	}

	// append the transaction to the mempool
	if errMempool := bc.mempool.AppendTransaction(tx, fee); errMempool != nil {
		return fmt.Errorf("error appending transaction %x to mempool: %w", tx.ID, errMempool)
	}

	bc.logger.Debugf("transaction %x added to mempool", tx.ID)

	// notify chain observers of a new transaction added into the mempool. This is required because
	// although mempool is a separate module, it represents an important part of the chain. Important
	// modules like the storage and the network (propagates via pubsub to other nodes) need to be
	// aware of this event
	bc.blockSubject.NotifyTxAdded(tx)

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

		// validate the header compatibility with the local chain (prevent downloading block if is incompatible)
		if err = bc.validator.ValidateHeader(header); err != nil {
			return fmt.Errorf("error validating header %s: %w", header.String(), err)
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

// OnUnconfirmedHeaderReceived is called when a remote node publishes that has added a new
// block to the (remote node) chain. Once the header is received, this function is executed
// and (the local chain) tries to add the block
func (bc *Blockchain) OnUnconfirmedHeaderReceived(peer peer.ID, header kernel.BlockHeader) {
	// make sure that the header is compatible with the local chain
	if err := bc.validator.ValidateHeader(&header); err != nil {
		bc.logger.Tracef("error validating header %s sent by %s: %s", header.String(), peer.String(), err)
	}

	// calculate the block hash
	hash, err := util.CalculateBlockHash(&header, bc.hasher)
	if err != nil {
		bc.logger.Errorf("error calculating block hash: %s", err)
	}

	// ask the peer for the block
	ctx, cancel := context.WithTimeout(context.Background(), bc.cfg.P2P.ConnTimeout)
	defer cancel()

	block, err := bc.p2pNet.AskSpecificBlock(ctx, peer, hash)
	if err != nil {
		bc.logger.Errorf("error asking for block %x to %s: %s", hash, peer.String(), err)
		return
	}

	// try to add the block to the chain
	if err = bc.AddBlock(block); err != nil {
		bc.logger.Tracef("error adding block %x to the chain: %s", block.Hash, err)
	}
}

// OnUnconfirmedTxReceived is called when a new transaction is received from the network
func (bc *Blockchain) OnUnconfirmedTxReceived(tx kernel.Transaction) error {
	if err := bc.AddTransaction(&tx); err != nil {
		return fmt.Errorf("error adding transaction %x to the chain: %w", tx.ID, err)
	}

	return nil
}

// OnUnconfirmedTxIDReceived is called when a new transaction ID is received from the network
func (bc *Blockchain) OnUnconfirmedTxIDReceived(peer peer.ID, txID []byte) {
	containsTx := bc.mempool.ContainsTx(string(txID))
	// if the transaction is already in the mempool, skip execution
	if containsTx {
		return
	}

	// ask the peer for the transaction
	ctx, cancel := context.WithTimeout(context.Background(), bc.cfg.P2P.ConnTimeout)
	defer cancel()

	// ask the peer for the whole transaction
	tx, err := bc.p2pNet.AskSpecificTx(ctx, peer, txID)
	if err != nil {
		bc.logger.Errorf("error asking for transaction %x to %s: %s", txID, peer.String(), err)
		return
	}

	if err = bc.AddTransaction(tx); err != nil {
		bc.logger.Errorf("error adding transaction %x to the chain: %s", tx.ID, err)
	}
}

func (bc *Blockchain) RegisterMetrics(registry *prometheus.Registry) {
	monitor.NewMetric(registry, monitor.Gauge, "chain_height", "Number of blocks in the chain", func() float64 {
		return float64(bc.GetLastHeight())
	})

	monitor.NewMetric(registry, monitor.Gauge, "chain_circulating_supply", "Circulating supply of the chain", func() float64 {
		totalSupply := 0
		remainingHeight := int(bc.lastHeight)
		reward := common.InitialCoinbaseReward

		for remainingHeight > 0 {
			// determine the number of blocks in the current halving period
			blocksInPeriod := common.HalvingInterval
			if remainingHeight < common.HalvingInterval {
				blocksInPeriod = remainingHeight
			}

			// calculate this halving period and add to total
			totalSupply += blocksInPeriod * reward

			// move to the next halving period
			remainingHeight -= blocksInPeriod
			reward /= 2
		}

		return kernel.ConvertFromChannoshisToCoins(uint(totalSupply))
	})

	monitor.NewMetric(registry, monitor.Gauge, "chain_last_block_size", "Size of latest block added to chain", func() float64 {
		block, err := bc.store.GetLastBlock()
		if err != nil {
			return 0.0
		}

		return float64(block.Size())
	})

	monitor.NewMetric(registry, monitor.Gauge, "chain_last_block_inputs_size", "Size of inputs in latest block added to chain", func() float64 {
		block, err := bc.store.GetLastBlock()
		if err != nil {
			return 0.0
		}

		inputSize := uint(0)
		for _, tx := range block.Transactions {
			for _, in := range tx.Vin {
				inputSize += in.Size()
			}
		}

		return float64(inputSize)
	})

	monitor.NewMetric(registry, monitor.Gauge, "chain_last_block_outputs_size", "Size of outputs in latest block added to chain", func() float64 {
		block, err := bc.store.GetLastBlock()
		if err != nil {
			return 0.0
		}

		outputSize := uint(0)
		for _, tx := range block.Transactions {
			for _, out := range tx.Vout {
				outputSize += out.Size()
			}
		}

		return float64(outputSize)
	})
	monitor.NewMetric(registry, monitor.Gauge, "chain_last_block_txs", "Number of transactions in latest block added to chain", func() float64 {
		block, err := bc.store.GetLastBlock()
		if err != nil {
			return 0.0
		}

		return float64(len(block.Transactions))
	})
}

func (bc *Blockchain) calculateTxFee(tx *kernel.Transaction) (uint, error) {
	// calculate the funds provided by the inputs
	inputBalance, err := bc.utxoSet.RetrieveInputsBalance(tx.Vin)
	if err != nil {
		return 0, fmt.Errorf("error retrieving inputs balance: %w", err)
	}

	// calculate the funds spent by the outputs
	outputBalance := tx.OutputAmount()

	// make sure that the output balance is not greater than the input balance
	if outputBalance > inputBalance {
		return 0, fmt.Errorf("output balance %d is greater than input balance %d", outputBalance, inputBalance)
	}

	// calculate the fee and append the transaction to the mempool
	return inputBalance - outputBalance, nil
}

// GetLastBlockHash returns the latest block hash
func (bc *Blockchain) GetLastBlockHash() []byte {
	return bc.lastBlockHash
}

// GetLastHeight returns the latest block height
func (bc *Blockchain) GetLastHeight() uint {
	return bc.lastHeight
}

// reconstructState retrieves all headers from the last block to the genesis block and reconstructs the UTXO set
func reconstructState(store storage.Storage, utxoSet *utxoset.UTXOSet, headers map[string]kernel.BlockHeader, lastBlockHash []byte) error {
	if len(lastBlockHash) == 0 {
		return fmt.Errorf("last block hash is empty")
	}

	listHashes := make([][]byte, 0)
	// retrieve all headers from the last block to the genesis block
	for len(lastBlockHash) != 0 {
		blockHeader, err := store.RetrieveHeaderByHash(lastBlockHash)
		if err != nil {
			return fmt.Errorf("error retrieving block header %x: %w", lastBlockHash, err)
		}

		// add the header to the map
		headers[string(lastBlockHash)] = *blockHeader
		// keep the hash in the list to reconstruct the UTXO set
		listHashes = append(listHashes, lastBlockHash)
		// move to the next block
		lastBlockHash = blockHeader.PrevBlockHash
	}

	// iterate the list of hashes in reverse order to reconstruct the UTXO set
	for i := range listHashes {
		blockHash := listHashes[len(listHashes)-1-i]
		block, err := store.RetrieveBlockByHash(blockHash)
		if err != nil {
			return fmt.Errorf("error retrieving block %x: %w", blockHash, err)
		}

		// add the block to the UTXO set
		err = utxoSet.AddBlock(block)
		if err != nil {
			return fmt.Errorf("error adding block %x to UTXO set: %w", blockHash, err)
		}
	}

	return nil
}
