package storage

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/monitor"
)

type metrics struct {
	persistedBlocks        uint64
	persistedHeaders       uint64
	retrievedLastBlock     uint64
	retrievedLastHeader    uint64
	retrievedLastBlockHash uint64
	retrievedGenesisBlock  uint64
	retrievedGenesisHeader uint64
	retrievedBlockByHash   uint64
	retrievedHeaderByHash  uint64
	onBlockAddition        uint64

	persistedBlocksTime        int64
	persistedHeadersTime       int64
	retrievedLastBlockTime     int64
	retrievedLastHeaderTime    int64
	retrievedLastBlockHashTime int64
	retrievedGenesisBlockTime  int64
	retrievedGenesisHeaderTime int64
	retrievedBlockByHashTime   int64
	retrievedHeaderByHashTime  int64
	onBlockAdditionTime        int64
}

type MeteredStorage struct {
	inner Storage
	*metrics
}

func NewMeteredStorage(inner Storage) *MeteredStorage {
	return &MeteredStorage{
		inner:   inner,
		metrics: &metrics{},
	}
}

func (ms *MeteredStorage) PersistBlock(block kernel.Block) error {
	startTime := time.Now()
	defer recordTimeAsync(&ms.persistedBlocks, &ms.persistedBlocksTime, startTime)

	return ms.inner.PersistBlock(block)
}

func (ms *MeteredStorage) PersistHeader(blockHash []byte, blockHeader kernel.BlockHeader) error {
	startTime := time.Now()
	defer recordTimeAsync(&ms.persistedHeaders, &ms.persistedHeadersTime, startTime)

	return ms.inner.PersistHeader(blockHash, blockHeader)
}

func (ms *MeteredStorage) GetLastBlock() (*kernel.Block, error) {
	startTime := time.Now()
	defer recordTimeAsync(&ms.retrievedLastBlock, &ms.retrievedLastBlockTime, startTime)

	return ms.inner.GetLastBlock()
}

func (ms *MeteredStorage) GetLastHeader() (*kernel.BlockHeader, error) {
	startTime := time.Now()
	defer recordTimeAsync(&ms.retrievedLastHeader, &ms.retrievedLastHeaderTime, startTime)

	return ms.inner.GetLastHeader()
}

func (ms *MeteredStorage) GetLastBlockHash() ([]byte, error) {
	startTime := time.Now()
	defer recordTimeAsync(&ms.retrievedLastBlockHash, &ms.retrievedLastBlockHashTime, startTime)

	return ms.inner.GetLastBlockHash()
}

func (ms *MeteredStorage) GetGenesisBlock() (*kernel.Block, error) {
	startTime := time.Now()
	defer recordTimeAsync(&ms.retrievedGenesisBlock, &ms.retrievedGenesisBlockTime, startTime)

	return ms.inner.GetGenesisBlock()
}

func (ms *MeteredStorage) GetGenesisHeader() (*kernel.BlockHeader, error) {
	startTime := time.Now()
	defer recordTimeAsync(&ms.retrievedGenesisHeader, &ms.retrievedGenesisHeaderTime, startTime)

	return ms.inner.GetGenesisHeader()
}

func (ms *MeteredStorage) RetrieveBlockByHash(hash []byte) (*kernel.Block, error) {
	startTime := time.Now()
	defer recordTimeAsync(&ms.retrievedBlockByHash, &ms.retrievedBlockByHashTime, startTime)

	return ms.inner.RetrieveBlockByHash(hash)
}

func (ms *MeteredStorage) RetrieveHeaderByHash(hash []byte) (*kernel.BlockHeader, error) {
	startTime := time.Now()
	defer recordTimeAsync(&ms.retrievedHeaderByHash, &ms.retrievedHeaderByHashTime, startTime)

	return ms.inner.RetrieveHeaderByHash(hash)
}

func (ms *MeteredStorage) Typ() string {
	return ms.inner.Typ()
}

func (ms *MeteredStorage) ID() string {
	return ms.inner.ID()
}

func (ms *MeteredStorage) OnBlockAddition(block *kernel.Block) {
	startTime := time.Now()
	defer recordTimeAsync(&ms.onBlockAddition, &ms.onBlockAdditionTime, startTime)

	ms.inner.OnBlockAddition(block)
}

func (ms *MeteredStorage) OnTxAddition(tx *kernel.Transaction) {
	ms.inner.OnTxAddition(tx)
}

func (ms *MeteredStorage) Close() error {
	return ms.inner.Close()
}

func (ms *MeteredStorage) RegisterMetrics(register *prometheus.Registry) {
	// todo() the type of storage should be added to the labels, but for that we must use the prometheus.NewCounterVec
	monitor.NewMetric(register, monitor.Counter, "storage_num_persisted_blocks", "Number of persisted blocks", func() float64 {
		return float64(atomic.LoadUint64(&ms.persistedBlocks))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_persisted_headers", "Number of persisted headers", func() float64 {
		return float64(atomic.LoadUint64(&ms.persistedHeaders))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_retrieved_last_block", "Number of retrieved last block", func() float64 {
		return float64(atomic.LoadUint64(&ms.retrievedLastBlock))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_retrieved_last_header", "Number of retrieved last header", func() float64 {
		return float64(atomic.LoadUint64(&ms.retrievedLastHeader))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_retrieved_last_block_hash", "Number of retrieved last block hash", func() float64 {
		return float64(atomic.LoadUint64(&ms.retrievedLastBlockHash))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_retrieved_genesis_block", "Number of retrieved genesis block", func() float64 {
		return float64(atomic.LoadUint64(&ms.retrievedGenesisBlock))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_retrieved_genesis_header", "Number of retrieved genesis header", func() float64 {
		return float64(atomic.LoadUint64(&ms.retrievedGenesisHeader))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_retrieved_block_by_hash", "Number of retrieved block by hash", func() float64 {
		return float64(atomic.LoadUint64(&ms.retrievedBlockByHash))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_retrieved_header_by_hash", "Number of retrieved header by hash", func() float64 {
		return float64(atomic.LoadUint64(&ms.retrievedHeaderByHash))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_num_on_block_addition", "Number of on block addition", func() float64 {
		return float64(atomic.LoadUint64(&ms.onBlockAddition))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_persisted_blocks_time", "Nanoseconds taken to persist blocks", func() float64 {
		return float64(atomic.LoadInt64(&ms.persistedBlocksTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_persisted_headers_time", "Nanoseconds taken to persist headers", func() float64 {
		return float64(atomic.LoadInt64(&ms.persistedHeadersTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_retrieved_last_block_time", "Nanoseconds taken to retrieve last block", func() float64 {
		return float64(atomic.LoadInt64(&ms.retrievedLastBlockTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_retrieved_last_header_time", "Nanoseconds taken to retrieve last header", func() float64 {
		return float64(atomic.LoadInt64(&ms.retrievedLastHeaderTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_retrieved_last_block_hash_time", "Nanoseconds taken to retrieve last block hash", func() float64 {
		return float64(atomic.LoadInt64(&ms.retrievedLastBlockHashTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_retrieved_genesis_block_time", "Nanoseconds taken to retrieve genesis block", func() float64 {
		return float64(atomic.LoadInt64(&ms.retrievedGenesisBlockTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_retrieved_genesis_header_time", "Nanoseconds taken to retrieve genesis header", func() float64 {
		return float64(atomic.LoadInt64(&ms.retrievedGenesisHeaderTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_retrieved_block_by_hash_time", "Nanoseconds taken to retrieve block by hash", func() float64 {
		return float64(atomic.LoadInt64(&ms.retrievedBlockByHashTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_retrieved_header_by_hash_time", "Nanoseconds taken to retrieve header by hash", func() float64 {
		return float64(atomic.LoadInt64(&ms.retrievedHeaderByHashTime))
	})

	monitor.NewMetric(register, monitor.Counter, "storage_on_block_addition_time", "Nanoseconds taken to on block addition", func() float64 {
		return float64(atomic.LoadInt64(&ms.onBlockAdditionTime))
	})
}

func recordTimeAsync(counter *uint64, timeCounter *int64, startTime time.Time) {
	atomic.AddUint64(counter, 1)
	atomic.AddInt64(timeCounter, time.Since(startTime).Nanoseconds())
}
