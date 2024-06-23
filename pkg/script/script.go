package script

import "fmt"

// Script types
const (
	// P2PK = Pay-to-PubKey
	P2PK = iota
	// P2PKH = Pay-to-PubKey-Hash
	// P2PKH

	// ...
)

type ScriptOperator uint

// Operators
const (
	OP_CHECKSIG ScriptOperator = iota
)

var operatorNames = [...]string{
	"OP_CHECKSIG",
}

func (op ScriptOperator) String() string {
	if op >= 0 && op < ScriptOperator(len(operatorNames)) {
		return operatorNames[op]
	}

	return "OP_UNKNOWN"
}

func NewScript(scriptType uint, pubKey []byte) string {
	switch scriptType {
	case P2PK:
		return fmt.Sprintf("%s %s", pubKey, OP_CHECKSIG.String())
	default:
		return "UNKNOWN"
	}

}
