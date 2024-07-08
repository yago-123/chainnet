package interpreter

import (
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"strconv"
	"strings"
)

const (
	scriptSeparator = " "

	opCheckSigVerifyStackLength = 2
)

// RPNInterpreter represents the interpreter for the Reverse Polish Notation (RPN) script
type RPNInterpreter struct {
	signer sign.Signature
}

func NewScriptInterpreter(signer sign.Signature) *RPNInterpreter {
	return &RPNInterpreter{
		signer: signer,
	}
}

// GenerateScriptSig evaluates the scriptPubKey requirement and generates the scriptSig that will unlock the scriptPubKey
func (rpn *RPNInterpreter) GenerateScriptSig(scriptPubKey string, privKey []byte, tx *kernel.Transaction) (string, error) {
	// converts script pub key into list of tokens and list of strings
	scriptTokens, scriptString := getTokenListFromScript(scriptPubKey)

	// push the signature to the stack
	stack := script.NewStack()

	// start generation of scriptSig unlocker
	for index, token := range scriptTokens {
		if token.IsUndefined() {
			return "", fmt.Errorf("undefined token %s in position %d", scriptString[index], index)
		}

		if token.IsOperator() {
			switch token {
			case script.OpChecksig:
				// generate the signature
				sig, err := rpn.generateOpCheckSig(stack, tx, privKey)
				if err != nil {
					return "", err
				}
				stack.Push(sig)
			default:
			}
		}

		if !token.IsOperator() {
			stack.Push(scriptString[index])
		}
	}

	if stack.Len() != 1 {
		return "", fmt.Errorf("invalid stack length after script execution")
	}

	return stack.Pop(), nil
}

// generateOpCheckSig creates the scriptSig that will unlock the scriptPubKey
func (rpn *RPNInterpreter) generateOpCheckSig(stack *script.Stack, tx *kernel.Transaction, privKey []byte) (string, error) {
	if stack.Len() < 1 {
		return "", fmt.Errorf("invalid stack length while generating signature in OP_CHECKSIG")
	}

	// we pop the public key from the stack but we don't need it
	// todo() or maybe we really need it? Verify the output maybe?
	_ = stack.Pop()

	signature, err := rpn.signer.Sign(tx.AssembleForSigning(), privKey)
	return string(signature), err
}

// VerifyScriptPubKey verifies the scriptPubKey by reconstructing the script and evaluating it
func (rpn *RPNInterpreter) VerifyScriptPubKey(scriptPubKey string, signature string, tx *kernel.Transaction) (bool, error) {
	// converts script pub key into list of tokens and list of strings
	scriptTokens, scriptString := getTokenListFromScript(scriptPubKey)

	// push the signature to the stack
	stack := script.NewStack()
	stack.Push(signature)

	// start evaluation of scriptPubKey
	for index, token := range scriptTokens {
		if token.IsUndefined() {
			return false, fmt.Errorf("undefined token %s in position %d", scriptString[index], index)
		}

		if token.IsOperator() {
			// perform operation based on operator with a and b
			switch token {
			case script.OpChecksig:
				ret, err := rpn.verifyOpCheckSig(stack, tx)
				if err != nil {
					return false, err
				}
				stack.Push(strconv.FormatBool(ret))
			default:
			}
		}

		if !token.IsOperator() {
			stack.Push(scriptString[index])
		}
	}

	if stack.Len() != 1 {
		return false, fmt.Errorf("invalid stack length after script execution")
	}

	return stack.Pop() == "true", nil
}

// verifyOpCheckSig verifies the signature of the transaction
func (rpn *RPNInterpreter) verifyOpCheckSig(stack *script.Stack, tx *kernel.Transaction) (bool, error) {
	if stack.Len() < opCheckSigVerifyStackLength {
		return false, fmt.Errorf("invalid stack length for OP_CHECKSIG")
	}

	pubKey := stack.Pop()
	signature := stack.Pop()

	// verify the signature
	return rpn.signer.Verify([]byte(signature), tx.AssembleForSigning(), base58.Decode(pubKey))
}

// getTokenListFromScript converts a string script to a list of script elements
func getTokenListFromScript(scriptPubKey string) ([]script.ScriptElement, []string) {
	scriptTokens := []script.ScriptElement{}
	scriptString := []string{}

	for _, element := range strings.Split(scriptPubKey, scriptSeparator) {
		var token script.ScriptElement

		if token, str, err := tryGetTokenLiteral(element); err != nil {
			scriptTokens = append(scriptTokens, token)
			scriptString = append(scriptString, str)
			continue
		}

		token = script.ConvertToScriptElement(element)
		scriptTokens = append(scriptTokens, token)
		// in case of simple tokens string does not add any additional unit of information (for now at least)
		scriptString = append(scriptString, "")
	}

	return scriptTokens, scriptString
}

// tryGetTokenLiteral tries to converts keys, hash keys etc to script.Literal type
func tryGetTokenLiteral(data string) (script.ScriptElement, string, error) {
	// if data have less than 2 elements, it means that there is no possible literal
	// must be at least 1 byte for declaring the type + 1 type for the unit of information
	if len(data) < 2 {
		return script.Undefined, "", fmt.Errorf("unable to decode token literal %s: not long enough", data)
	}

	runed := []rune(data)
	token := script.ConvertRuneLiteralToScriptElement(runed[0])
	if token == script.Undefined {
		return token, "", fmt.Errorf("unable to decode token literal %s", data)
	}

	return token, string(runed[1:]), nil
}
