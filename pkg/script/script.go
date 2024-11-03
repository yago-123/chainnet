package script

import (
	"fmt"
	"strings"

	"github.com/yago-123/chainnet/pkg/crypto/hash"

	"github.com/btcsuite/btcutil/base58"
)

type ScriptType uint //nolint:revive // ScriptType is a type for script types

const (
	MinLengthOfLiteral = 2
)

type Script []ScriptElement
type ScriptElement uint //nolint:revive // ScriptElement is a type for script elements

// Script types
const (
	P2PK ScriptType = iota
	P2PKH

	UndefinedScriptType
	// ...
)

var scriptStructure = map[ScriptType]Script{ //nolint:gochecknoglobals // must be a global variable
	P2PK:                {PubKey, OpChecksig},
	P2PKH:               {OpDup, OpHash160, PubKeyHash, OpEqualVerify, OpChecksig},
	UndefinedScriptType: {Undefined},
}

// scripTypeStrings is a map that contains the string representation of the script types
var scripTypeStrings = map[string]ScriptType{
	"P2PK":  P2PK,
	"P2PKH": P2PKH,
}

const (
	// Special elements
	PubKey ScriptElement = iota
	PubKeyHash
	Signature

	// Operators
	OpChecksig
	OpDup
	OpHash160
	OpEqualVerify

	Undefined
)

const (
	scriptSeparator = " "
)

var operatorNames = [...]string{ //nolint:gochecknoglobals // must be a global variable
	// literals
	"PUB_KEY",
	"PUB_KEY_HASH",
	"SIGNATURE",

	// operations
	"OP_CHECKSIG",
	"OP_DUP",
	"OP_HASH160",
	"OP_EQUALVERIFY",

	// undefined
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
	return op >= OpChecksig && op <= OpEqualVerify
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

// NewScript generates a new script based on the type and public key
func NewScript(scriptType ScriptType, pubKey []byte) string {
	// if there is no public key, return undefined directly
	if len(pubKey) == 0 {
		return scriptStructure[UndefinedScriptType].String(pubKey)
	}

	// generate script based on type
	script := scriptStructure[scriptType]

	// todo() the render will switch to a hex string eventually
	// render script to string
	return script.String(pubKey)
}

// String returns the string representation of the script
func (s Script) String(pubKey []byte) string {
	var err error
	var rendered []string

	for _, element := range s {
		toRender := ""

		if element.OutsideBoundaries() {
			return Undefined.String()
		}

		// render special cases adding the preffix so we can later know which type of literal was written. This
		// includes pubKey, pubHashKey, signature...
		if element.IsLiteral() {
			literalRendered := []byte{}

			if element == PubKey {
				literalRendered = pubKey
			}

			if element == PubKeyHash {
				ripemd160 := hash.NewRipemd160()
				literalRendered, err = ripemd160.Hash(pubKey)
				if err != nil {
					// highly unlikely that hash initialization will fail, but if it does, abort the operation by
					// returning undefined, no point in making the code more unintelligible by returning an error
					return Undefined.String()
				}
			}

			toRender = fmt.Sprintf("%c%s", byte(element.ToUint()), base58.Encode(literalRendered))
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

// StringToScript converts a script pub key string into Script type and array of literals (like pub key, hash pub key, etc)
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

// DetermineScriptType tries to derive the script type based on a set of elements that form a script
func DetermineScriptType(script Script) ScriptType {
	for k, v := range scriptStructure {
		if scriptsMatch(v, script) {
			return k
		}
	}

	return UndefinedScriptType
}

// DetermineScriptTYpeFromStringType returns the script type based on a string representation. For example:
// - "P2PK" -> P2PK
// - "P2PKH" -> P2PKH
func DetermineScriptTypeFromStringType(script string) ScriptType {
	typ, ok := scripTypeStrings[script]
	if !ok {
		return UndefinedScriptType
	}

	return typ
}

// scriptsMatch checks if two scripts contain the same script elements in the same order
func scriptsMatch(script1, script2 Script) bool {
	if len(script1) != len(script2) {
		return false
	}

	for i, element := range script1 {
		if element != script2[i] {
			return false
		}
	}

	return true
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
