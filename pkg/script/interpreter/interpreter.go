package interpreter

import (
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"fmt"
	"strconv"
)

const (
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
	scriptTokens, scriptString, err := script.StringToScript(scriptPubKey)
	if err != nil {
		return "", err
	}

	// push the signature to the stack
	stack := script.NewStack()

	// start generation of scriptSig unlocker
	for index, token := range scriptTokens {
		if token.IsUndefined() {
			return "", fmt.Errorf("undefined token %s in position %d", scriptString[index], index)
		}

		if token.IsOperator() {
			switch token { //nolint:exhaustive // only check operators
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
	scriptTokens, scriptString, err := script.StringToScript(scriptPubKey)
	if err != nil {
		return false, err
	}

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
			switch token { //nolint:exhaustive // only check operators
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
	return rpn.signer.Verify([]byte(signature), tx.AssembleForSigning(), []byte(pubKey))
}
