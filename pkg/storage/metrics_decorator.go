package storage

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yago-123/chainnet/pkg/kernel"
	"github.com/yago-123/chainnet/pkg/monitor"
	"sync/atomic"
	"time"
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
	atomic.AddUint64(&ms.persistedBlocks, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.persistedBlocksTime, time.Since(startTime).Nanoseconds())

	return ms.inner.PersistBlock(block)
}

func (ms *MeteredStorage) PersistHeader(blockHash []byte, blockHeader kernel.BlockHeader) error {
	atomic.AddUint64(&ms.persistedHeaders, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.persistedHeadersTime, time.Since(startTime).Nanoseconds())

	return ms.inner.PersistHeader(blockHash, blockHeader)
}

func (ms *MeteredStorage) GetLastBlock() (*kernel.Block, error) {
	atomic.AddUint64(&ms.retrievedLastBlock, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.retrievedLastBlockTime, time.Since(startTime).Nanoseconds())

	return ms.inner.GetLastBlock()
}

func (ms *MeteredStorage) GetLastHeader() (*kernel.BlockHeader, error) {
	atomic.AddUint64(&ms.retrievedLastHeader, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.retrievedLastHeaderTime, time.Since(startTime).Nanoseconds())

	return ms.inner.GetLastHeader()
}

func (ms *MeteredStorage) GetLastBlockHash() ([]byte, error) {
	atomic.AddUint64(&ms.retrievedLastBlockHash, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.retrievedLastBlockHashTime, time.Since(startTime).Nanoseconds())

	return ms.inner.GetLastBlockHash()
}

func (ms *MeteredStorage) GetGenesisBlock() (*kernel.Block, error) {
	atomic.AddUint64(&ms.retrievedGenesisBlock, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.retrievedGenesisBlockTime, time.Since(startTime).Nanoseconds())

	return ms.inner.GetGenesisBlock()
}

func (ms *MeteredStorage) GetGenesisHeader() (*kernel.BlockHeader, error) {
	atomic.AddUint64(&ms.retrievedGenesisHeader, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.retrievedGenesisHeaderTime, time.Since(startTime).Nanoseconds())

	return ms.inner.GetGenesisHeader()
}

func (ms *MeteredStorage) RetrieveBlockByHash(hash []byte) (*kernel.Block, error) {
	atomic.AddUint64(&ms.retrievedBlockByHash, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.retrievedBlockByHashTime, time.Since(startTime).Nanoseconds())

	return ms.inner.RetrieveBlockByHash(hash)
}

func (ms *MeteredStorage) RetrieveHeaderByHash(hash []byte) (*kernel.BlockHeader, error) {
	atomic.AddUint64(&ms.retrievedHeaderByHash, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.retrievedHeaderByHashTime, time.Since(startTime).Nanoseconds())

	return ms.inner.RetrieveHeaderByHash(hash)
}

func (ms *MeteredStorage) Typ() string {
	return ms.inner.Typ()
}

func (ms *MeteredStorage) ID() string {
	return ms.inner.ID()
}

func (ms *MeteredStorage) OnBlockAddition(block *kernel.Block) {
	atomic.AddUint64(&ms.onBlockAddition, 1)
	startTime := time.Now()
	defer atomic.AddInt64(&ms.onBlockAdditionTime, time.Since(startTime).Nanoseconds())

	ms.inner.OnBlockAddition(block)
}

func (ms *MeteredStorage) OnTxAddition(tx *kernel.Transaction) {
	ms.inner.OnTxAddition(tx)
}

func (ms *MeteredStorage) Close() error {
	return ms.inner.Close()
}

func (ms *MeteredStorage) RegisterMetrics(register *prometheus.Registry) {
	monitor.NewMetricWithLabelsSync(register, monitor.Gauge, "storage_num_persisted_blocks", "Number of persisted blocks",
		[]string{monitor.StorageLabel},
		func(gaugeVec *prometheus.GaugeVec) {
			gaugeVec.WithLabelValues(ms.inner.Typ()).Set(float64(atomic.LoadUint64(&ms.persistedBlocks)))
		},
	)

}
