package script //nolint:testpackage // don't create separate package for tests

import (
	"github.com/yago-123/chainnet/pkg/crypto/hash"
	"fmt"
	"reflect"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/stretchr/testify/require"
)

func TestNewScript_P2PK(t *testing.T) {
	type args struct {
		scriptType ScriptType
		pubKey     []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"regular script generation for P2PK", args{scriptType: P2PK, pubKey: []byte("public-key")}, fmt.Sprintf("%c%s OP_CHECKSIG", PubKey, base58.Encode([]byte("public-key")))},
		{"generation of P2PK with empty public key", args{scriptType: P2PK, pubKey: []byte{}}, Undefined.String()},
		{"P2PK with pubkey equal to PubKey token identifier", args{scriptType: P2PK, pubKey: []byte(fmt.Sprintf("%d", PubKey))}, fmt.Sprintf("%cq OP_CHECKSIG", PubKey)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScript(tt.args.scriptType, tt.args.pubKey); got != tt.want {
				t.Errorf("NewScript() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewScript_P2PKH(t *testing.T) {
	ripemd160 := hash.NewRipemd160()
	pubKeyHash, err := ripemd160.Hash([]byte("public-key"))
	require.NoError(t, err)

	tokenIdentifierHashed, err := ripemd160.Hash([]byte(fmt.Sprintf("%d", PubKeyHash)))
	require.NoError(t, err)

	type args struct {
		scriptType ScriptType
		pubKey     []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"regular script generation for P2PKH", args{scriptType: P2PKH, pubKey: []byte("public-key")}, fmt.Sprintf("OP_DUP OP_HASH160 %c%s OP_EQUALVERIFY OP_CHECKSIG", PubKeyHash, base58.Encode(pubKeyHash))},
		{"generation of P2PKH with empty public key", args{scriptType: P2PKH, pubKey: []byte{}}, Undefined.String()},
		{"P2PKH with pubkey equal to PubKeyHash token identifier", args{scriptType: P2PKH, pubKey: []byte(fmt.Sprintf("%d", PubKeyHash))}, fmt.Sprintf("OP_DUP OP_HASH160 %c%s OP_EQUALVERIFY OP_CHECKSIG", PubKeyHash, base58.Encode(tokenIdentifierHashed))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScript(tt.args.scriptType, tt.args.pubKey); got != tt.want {
				t.Errorf("NewScript() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToScript(t *testing.T) {
	type args struct {
		script string
	}
	tests := []struct {
		name    string
		args    args
		want    Script
		want1   []string
		wantErr bool
	}{
		{"", args{NewScript(P2PK, []byte("pubkey1"))}, Script([]ScriptElement{PubKey, OpChecksig}), []string{"pubkey1", ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := StringToScript(tt.args.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToScript() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("StringToScript() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_tryExtractTokenLiteral(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name  string
		args  args
		want  ScriptElement
		want1 string
	}{
		{"extract regular literal", args{data: fmt.Sprintf("%c%s", PubKey, "pubkey1")}, PubKey, "pubkey1"},
		{"extract literal with less than 2 elements", args{data: fmt.Sprintf("%d", PubKey)}, Undefined, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tryExtractTokenLiteral(tt.args.data)
			if got != tt.want {
				t.Errorf("tryExtractTokenLiteral() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("tryExtractTokenLiteral() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
