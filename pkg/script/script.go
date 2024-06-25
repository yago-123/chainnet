package script

import "strings"

type ScriptType uint

// Script types
const (
	// P2PK = Pay-to-PubKey
	P2PK ScriptType = iota
	// P2PKH = Pay-to-PubKey-Hash
	// P2PKH

	// ...
)

type Script []ScriptElement
type ScriptElement uint

const (
	// Special elements
	PUB_KEY ScriptElement = iota

	// Operators
	OP_CHECKSIG

	UNDEFINED
)

var operatorNames = [...]string{
	"PUB_KEY",
	"OP_CHECKSIG",

	"UNDEFINED",
}

func (op ScriptElement) OutsideBoundaries() bool {
	return op >= ScriptElement(len(operatorNames)-1)
}

func (op ScriptElement) IsSpecialCase() bool {
	// todo() extend with more other special cases
	return op >= PUB_KEY && op <= PUB_KEY
}

func (op ScriptElement) String() string {
	if op >= ScriptElement(len(operatorNames)) {
		// return UNDEFINED element
		return operatorNames[len(operatorNames)-1]
	}

	return operatorNames[op]
}

func NewScript(scriptType ScriptType, pubKey []byte) string {
	script := Script{UNDEFINED}

	// if there is no public key, return undefined directly
	if len(pubKey) == 0 {
		return UNDEFINED.String()
	}

	// generate script based on type
	switch scriptType {
	case P2PK:
		script = Script{PUB_KEY, OP_CHECKSIG}
	}

	// todo() the render will switch to a hex string eventually
	// render script to string
	return script.String(pubKey)
}

func (s Script) String(pubKey []byte) string {
	rendered := []string{}

	for _, element := range s {
		toRender := ""

		if element.OutsideBoundaries() {
			return UNDEFINED.String()
		}

		// render special cases
		if element.IsSpecialCase() {
			switch element {
			case PUB_KEY:
				toRender = string(pubKey)
			}
		}

		// render operators
		if !element.IsSpecialCase() {
			toRender = element.String()
		}

		// if no element has been rendered, return UNDEFINED
		if toRender == "" {
			return UNDEFINED.String()
		}

		rendered = append(rendered, toRender)
	}

	return strings.Join(rendered, " ")
}
