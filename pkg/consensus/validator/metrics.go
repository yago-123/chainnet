package validator

type HValidatorMetrics struct {
	txMetrics     *HValidatorTxMetrics
	headerMetrics *HValidatorHeaderMetrics
	blockMetrics  *HValidatorBlockMetrics
}

type HValidatorTxMetrics struct {
	totalAnalyzed, totalRejected uint64
}

type HValidatorHeaderMetrics struct {
	totalAnalyzed, totalRejected uint64
}

type HValidatorBlockMetrics struct {
	totalAnalyzed, totalRejected uint64
}
