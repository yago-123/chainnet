package script

import (
	"github.com/btcsuite/btcutil/base58"
	"strings"
)

type ScriptType uint //nolint:revive // ScriptType is a type for script types

// Script types
const (
	// P2PK = Pay-to-PubKey
	P2PK ScriptType = iota
	P2PKH
	// P2PKH

	// ...
)

type Script []ScriptElement
type ScriptElement uint //nolint:revive // ScriptElement is a type for script elements

type Literal struct {
	typ  ScriptElement
	data string
}

const (
	// Special elements
	PubKey ScriptElement = iota
	PubKeyHash
	Signature

	// Operators
	OpChecksig

	Undefined
)

var operatorNames = [...]string{ //nolint:gochecknoglobals // must be a global variable
	"PUB_KEY",
	"PUB_KEY_HASH",
	"SIGNATURE",
	"OP_CHECKSIG",

	"UNDEFINED",
}

// ConvertToScriptElement converts a string to a ScriptElement
func ConvertToScriptElement(element string) ScriptElement {
	for i, name := range operatorNames {
		if name == element {
			return ScriptElement(i)
		}
	}

	return Undefined
}

func ConvertRuneLiteralToScriptElement(element rune) ScriptElement {
	if uint(element) >= PubKey.ToUint() && uint(element) <= PubKeyHash.ToUint() {
		return ScriptElement(element)
	}

	return Undefined
}

// OutsideBoundaries checks if the element is outside the boundaries
func (op ScriptElement) OutsideBoundaries() bool {
	return op >= ScriptElement(len(operatorNames)-1)
}

// IsSpecialCase checks if the element is a special case
func (op ScriptElement) IsSpecialCase() bool {
	// todo() extend with more other special cases
	return op >= PubKey && op <= Signature
}

// IsOperator checks if the element is an operator
func (op ScriptElement) IsOperator() bool {
	return op == OpChecksig
}

func (op ScriptElement) IsUndefined() bool {
	return op == Undefined
}

// String returns the string representation of the element
func (op ScriptElement) String() string {
	if op >= ScriptElement(len(operatorNames)) {
		// return Undefined element
		return operatorNames[len(operatorNames)-1]
	}

	return operatorNames[op]
}

func (op ScriptElement) ToUint() uint {
	return uint(op)
}

func NewScript(scriptType ScriptType, pubKey []byte) string {
	script := Script{Undefined}

	// if there is no public key, return undefined directly
	if len(pubKey) == 0 {
		return Undefined.String()
	}

	// generate script based on type
	switch scriptType {
	case P2PK:
		script = Script{PubKey, OpChecksig}
	case P2PKH:
	// todo() implement P2PKH
	default:
	}

	// todo() the render will switch to a hex string eventually
	// render script to string
	return script.String(pubKey)
}

// String returns the string representation of the script
func (s Script) String(pubKey []byte) string {
	rendered := []string{}

	for _, element := range s {
		toRender := ""

		if element.OutsideBoundaries() {
			return Undefined.String()
		}

		// render special cases
		if element.IsSpecialCase() {
			switch element { //nolint:gocritic,exhaustive // number of elements will be increased in the future
			case PubKey:
				toRender = base58.Encode(pubKey)
			}
		}

		// render operators
		if !element.IsSpecialCase() {
			toRender = element.String()
		}

		// if no element has been rendered, return Undefined
		if toRender == "" {
			return Undefined.String()
		}

		rendered = append(rendered, toRender)
	}

	return strings.Join(rendered, " ")
}
