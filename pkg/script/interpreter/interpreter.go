package interpreter

import (
	"chainnet/pkg/crypto/hash"
	"chainnet/pkg/crypto/sign"
	"chainnet/pkg/kernel"
	"chainnet/pkg/script"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"strconv"
	"strings"
)

const (
	opCheckSigVerifyMinStackLength = 2
	opDupMinStackLength            = 1
	opHash160MinStackLength        = 1
	opEqualVerifyMinStackLength    = 2
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
				var sig []byte
				// generate the signature
				if stack.Len() < 1 {
					return "", fmt.Errorf("invalid stack length while generating signature in OP_CHECKSIG")
				}

				// we pop the public key from the stack but we don't need it
				// todo() or maybe we really need it? Verify the output maybe?
				_ = stack.Pop()

				sig, err = rpn.signer.Sign(tx.AssembleForSigning(), privKey)
				if err != nil {
					return "", fmt.Errorf("couldn't sign transaction: %w", err)
				}
				stack.Push(string(sig))
			default:
			}
		}

		if !token.IsOperator() {
			stack.Push(scriptString[index])
		}
	}

	ret := []string{}
	for _ = range stack.Len() {
		ret = append(ret, base58.Encode([]byte(stack.Pop())))
	}

	return strings.Join(ret, " "), nil
}

// VerifyScriptPubKey verifies the scriptPubKey by reconstructing the script and evaluating it
func (rpn *RPNInterpreter) VerifyScriptPubKey(scriptPubKey string, scriptSig string, tx *kernel.Transaction) (bool, error) {
	stack := script.NewStack()

	// converts script pub key into list of tokens and list of strings
	scriptTokens, scriptString, err := script.StringToScript(scriptPubKey)
	if err != nil {
		return false, err
	}

	// iterate over the scriptSig and push values to the stack
	for _, element := range strings.Split(scriptSig, " ") {
		stack.Push(string(base58.Decode(element)))
	}

	// start evaluation of scriptPubKey
	for index, token := range scriptTokens {
		if token.IsUndefined() {
			return false, fmt.Errorf("undefined token %s in position %d", scriptString[index], index)
		}

		if token.IsOperator() {
			// perform operation based on operator with a and b
			switch token { //nolint:exhaustive // only check operators
			case script.OpChecksig:
				if stack.Len() < opCheckSigVerifyMinStackLength {
					return false, fmt.Errorf("invalid stack length for OP_CHECKSIG")
				}

				pubKey := stack.Pop()
				sig := stack.Pop()

				// verify the signature
				ret, err := rpn.signer.Verify([]byte(sig), tx.AssembleForSigning(), []byte(pubKey))
				if err != nil {
					return false, fmt.Errorf("couldn't verify signature: %w", err)
				}
				stack.Push(strconv.FormatBool(ret))
			case script.OpDup:
				if stack.Len() < opDupMinStackLength {
					return false, fmt.Errorf("invalid stack length for OP_DUP")
				}

				val := stack.Pop()
				stack.Push(val)
				stack.Push(val)
			case script.OpHash160:
				if stack.Len() < opHash160MinStackLength {
					return false, fmt.Errorf("invalid stack length for OP_HASH160")
				}

				val := stack.Pop()
				hashedVal, err := hash.NewRipemd160().Hash([]byte(val))
				if err != nil {
					return false, fmt.Errorf("couldn't hash value: %w", err)
				}

				stack.Push(string(hashedVal))
			case script.OpEqualVerify:
				if stack.Len() < opEqualVerifyMinStackLength {
					return false, fmt.Errorf("invalid stack length for OP_EQUALVERIFY")
				}

				val1 := stack.Pop()
				val2 := stack.Pop()

				stack.Push(strconv.FormatBool(val1 == val2))
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
