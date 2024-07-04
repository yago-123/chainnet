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
	// represents the script script pub key as a list of tokens in string format
	scriptString := strings.Split(scriptPubKey, scriptSeparator)

	// represents the script pub key as a list of tokens in script format
	scriptTokens := getTokenListFromScript(scriptPubKey)

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
	// represents the script script pub key as a list of tokens in string format
	scriptString := strings.Split(scriptPubKey, scriptSeparator)

	// represents the script pub key as a list of tokens in script format
	scriptTokens := getTokenListFromScript(scriptPubKey)

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
func getTokenListFromScript(scriptPubKey string) []script.ScriptElement {
	scriptTokens := []script.ScriptElement{}
	for _, element := range strings.Split(scriptPubKey, scriptSeparator) {
		var token script.ScriptElement

		// todo() to be developed in case of addition of more script types

		// todo() improve pubKey detection...
		if len(element) >= sign.PubKeyLengthECDSAP256Min {
			token = script.PubKey
			scriptTokens = append(scriptTokens, token)
			continue
		}

		token = script.ConvertToScriptElement(element)
		scriptTokens = append(scriptTokens, token)
	}

	return scriptTokens
}
