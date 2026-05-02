package v1beta

import (
	"encoding/hex"
	"fmt"

	"github.com/yago-123/chainnet-sdk-go/v1beta/generated"
)

func encodeHex(value []byte) string {
	return hex.EncodeToString(value)
}

func uintToInt(field string, value uint) (int, error) {
	maxInt := ^uint(0) >> 1
	if value > maxInt {
		return 0, fmt.Errorf("%s exceeds int max", field)
	}

	return int(value), nil
}

func getHeadersParams(opts *GetHeadersOptions) *generated.GetHeadersParams {
	if opts == nil {
		return nil
	}

	params := &generated.GetHeadersParams{
		Limit: opts.Limit,
	}
	if opts.Order != nil {
		order := generated.GetHeadersParamsOrder(*opts.Order)
		params.Order = &order
	}

	return params
}
