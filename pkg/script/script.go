package script

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
)

type ScriptType uint //nolint:revive // ScriptType is a type for script types

const (
	MinLengthOfLiteral = 2
)

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

const (
	// Special elements
	PubKey ScriptElement = iota
	PubKeyHash
	Signature

	// Operators
	OpChecksig

	Undefined
)

const (
	scriptSeparator = " "
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

// OutsideBoundaries checks if the element is outside the boundaries
func (op ScriptElement) OutsideBoundaries() bool {
	return op >= ScriptElement(len(operatorNames)-1)
}

// IsLiteral checks if the element is of literal type
func (op ScriptElement) IsLiteral() bool {
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
		script = Script{}
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

		// render special cases adding the preffix so we can later know which type of literal was written. This
		// includes pubKey, pubHashKey, signature...
		if element.IsLiteral() {
			toRender = fmt.Sprintf("%c%s", byte(element.ToUint()), base58.Encode(pubKey))
		}

		// render operators
		if !element.IsLiteral() {
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

// StringToScript converts a script string into Script type and array of literals (like pub key, hash pub key, etc)
func StringToScript(script string) (Script, []string, error) {
	scriptTokens := []ScriptElement{}
	scriptString := []string{}

	for _, element := range strings.Split(script, scriptSeparator) {
		var token ScriptElement
		var literal string

		if token, literal = tryExtractTokenLiteral(element); token != Undefined {
			scriptTokens = append(scriptTokens, token)
			scriptString = append(scriptString, string(base58.Decode(literal)))
			continue
		}

		token = ConvertToScriptElement(element)
		scriptTokens = append(scriptTokens, token)
		// in case of simple tokens string does not add any additional unit of information (for now at least)
		scriptString = append(scriptString, "")
	}

	return scriptTokens, scriptString, nil
}

// tryExtractTokenLiteral tries to converts keys, hash keys etc to script.Literal type
func tryExtractTokenLiteral(data string) (ScriptElement, string) {
	// if data have less than 2 elements, it means that there is no possible literal
	// must be at least 1 byte for declaring the type + 1 type for the unit of information
	if len(data) < MinLengthOfLiteral {
		return Undefined, ""
	}

	opcodeByte := data[0]
	token := ScriptElement(uint(opcodeByte))

	if !token.OutsideBoundaries() && token.IsLiteral() {
		return token, data[1:]
	}

	return Undefined, ""
}
