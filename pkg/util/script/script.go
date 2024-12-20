package script

import (
	"strings"

	"github.com/btcsuite/btcutil/base58"
)

const (
	// ScriptSigSeparator represents the value used to join script sig arguments into a single string
	ScriptSigSeparator = " "
)

func EncodeScriptSig(scriptSig [][]byte) string {
	ret := []string{}
	for _, val := range scriptSig {
		ret = append(ret, base58.Encode(val))
	}

	return strings.Join(ret, ScriptSigSeparator)
}

func DecodeScriptSig(scriptSig string) [][]byte {
	ret := [][]byte{}
	for _, val := range strings.Split(scriptSig, ScriptSigSeparator) {
		ret = append(ret, base58.Decode(val))
	}

	return ret
}
