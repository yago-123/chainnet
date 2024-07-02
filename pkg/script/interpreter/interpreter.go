package interpreter

import (
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"fmt"
	"strconv"
	"strings"
)

const (
	scriptSeparator = " "

	opCheckSigVerifyStackLength = 2
)

type RPNInterpreter struct {
	signer sign.Signature
}

func NewScriptInterpreter(signer sign.Signature) *RPNInterpreter {
	return &RPNInterpreter{
		signer: signer,
	}
}

func (rpn *RPNInterpreter) GenerateScriptSig(scriptPubKey string, privKey []byte, tx kernel.Transaction) (string, error) {
	// represents the script script pub key as a list of tokens in string format
	scriptString := strings.Split(scriptPubKey, scriptSeparator)

	// represents the script pub key as a list of tokens in script format
	scriptTokens := getTokenListFromScript(scriptPubKey)

	// push the signature to the stack
	stack := script.NewStack()

	// start generation of scriptSig unlocker
	for index, token := range scriptTokens {
		if token.IsUndefined() {
			if token.IsUndefined() {
				return "", fmt.Errorf("undefined token %s in position %d", scriptString[index], index)
			}
		}

		if token.IsOperator() {
			switch token {
			case script.OpChecksig:
				// generate the signature
				sig, err := rpn.generateOpCheckSig(stack, &tx, privKey)
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

func (rpn *RPNInterpreter) VerifyScriptPubKey(scriptPubKey string, signature string, tx *kernel.Transaction) (string, error) {
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
			if token.IsUndefined() {
				return "", fmt.Errorf("undefined token %s in position %d", scriptString[index], index)
			}
		}

		if token.IsOperator() {
			// perform operation based on operator with a and b
			switch token {
			case script.OpChecksig:
				ret, err := rpn.verifyOpCheckSig(stack, tx)
				if err != nil {
					return "", err
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
		return "", fmt.Errorf("invalid stack length after script execution")
	}

	return stack.Pop(), nil
}

func (rpn *RPNInterpreter) verifyOpCheckSig(stack *script.Stack, tx *kernel.Transaction) (bool, error) {
	if stack.Len() < opCheckSigVerifyStackLength {
		return false, fmt.Errorf("invalid stack length for OP_CHECKSIG")
	}

	pubKey := stack.Pop()
	signature := stack.Pop()

	// verify the signature
	return rpn.signer.Verify([]byte(signature), tx.AssembleForSigning(), []byte(pubKey))
}

// getTokenListFromScript converts a string script to a list of script elements
func getTokenListFromScript(scriptPubKey string) []script.ScriptElement {
	scriptTokens := []script.ScriptElement{}
	for _, element := range strings.Split(scriptPubKey, scriptSeparator) {
		token := script.ConvertToScriptElement(element)
		scriptTokens = append(scriptTokens, token)
	}

	return scriptTokens
}
