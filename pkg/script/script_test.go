package script //nolint:testpackage // don't create separate package for tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	util_p2pkh "github.com/yago-123/chainnet/pkg/util/p2pkh"

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
	addressP2PKH, err := util_p2pkh.GenerateP2PKHAddrFromPubKey([]byte("public-key"), 1)
	require.NoError(t, err)

	pubKeyHash, _, err := util_p2pkh.ExtractPubKeyHashedFromP2PKHAddr(addressP2PKH)
	require.NoError(t, err)

	type args struct {
		scriptType   ScriptType
		addressP2PKH []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"regular script generation for P2PKH", args{scriptType: P2PKH, addressP2PKH: addressP2PKH}, fmt.Sprintf("OP_DUP OP_HASH160 %c%s OP_EQUALVERIFY OP_CHECKSIG", PubKeyHash, base58.Encode(pubKeyHash))},
		{"generation of P2PKH with empty public key", args{scriptType: P2PKH, addressP2PKH: []byte{}}, Undefined.String()},
		{"generation of P2PKH with short P2PKH address (trim 1 character)", args{scriptType: P2PKH, addressP2PKH: addressP2PKH[:24]}, Undefined.String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScript(tt.args.scriptType, tt.args.addressP2PKH); got != tt.want {
				t.Errorf("NewScript() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanBeUnlockedWithForP2PK(t *testing.T) {
	type args struct {
		scriptPubKey string
		address      []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"correct scriptPubKey and address", args{NewScript(P2PK, []byte("pubkey-1")), []byte("pubkey-1")}, true},
		{"incorrect scriptPubKey and address", args{NewScript(P2PK, []byte("pubkey-1")), []byte("pubkey-2")}, false},
		{"empty scriptPubKey", args{NewScript(P2PK, []byte("")), []byte("pubkey-1")}, false},
		{"empty address", args{NewScript(P2PK, []byte("pubkey-1")), []byte("")}, false},
		{"empty scriptPubKey and address", args{NewScript(P2PK, []byte("")), []byte("")}, false},
		{"random scriptPubKey", args{"random script pub key", []byte("")}, false},
		{"random scriptPubKey", args{"random script pub key", []byte("pubkey-1")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CanBeUnlockedWith(tt.args.scriptPubKey, tt.args.address, 1), "CanBeUnlockedWith(%v, %v)", tt.args.scriptPubKey, tt.args.address)
		})
	}
}

func TestCanBeUnlockedWithForP2PKH(t *testing.T) {
	type args struct {
		scriptPubKey string
		address      []byte
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"correct pubkey and address", args{NewScript(P2PKH, base58.Decode("XsyV9e37GnmCsqsKucyfe5scFpPRFTKth")), base58.Decode("aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTFiUHFDpSgR5ZaWVxcEgVr3C7VoLSwpFsZYKiG1QGdcWvoVdA9FrNvRQieoMUfS73arccNJ7GiuFpZVTBxWrH3ZU6")}, true},
		{"incorrect pubkey and address", args{NewScript(P2PKH, base58.Decode("hVb8js1bpmyYsQrKQFfWyaUski2wPrqew")), base58.Decode("hVb8js1bpmyYsQrKQFfWyaUs111111111")}, false},
		{"empty pubkey", args{NewScript(P2PKH, []byte("")), base58.Decode("hVb8js1bpmyYsQrKQFfWyaUski2wPrqew")}, false},
		{"empty address", args{NewScript(P2PKH, base58.Decode("hVb8js1bpmyYsQrKQFfWyaUski2wPrqew")), []byte{}}, false},
		{"empty pubkey and address", args{NewScript(P2PKH, []byte("")), []byte("")}, false},
		{"random script pub key", args{"random script pub key", []byte("")}, false},
		{"random script pub key", args{"random script pub key", base58.Decode("hVb8js1bpmyYsQrKQFfWyaUski2wPrqew")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CanBeUnlockedWith(tt.args.scriptPubKey, tt.args.address, 1), "CanBeUnlockedWith(%v, %v)", tt.args.scriptPubKey, tt.args.address)
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
