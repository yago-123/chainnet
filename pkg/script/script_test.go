package script //nolint:testpackage // don't create separate package for tests

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewScript(t *testing.T) {
	type args struct {
		scriptType ScriptType
		pubKey     []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"regular script generation for P2PK", args{scriptType: P2PK, pubKey: []byte("public-key")}, fmt.Sprintf("%c7KTAvebKjNUpoi OP_CHECKSIG", PubKey)},
		{"generation with empty public key", args{scriptType: P2PK, pubKey: []byte{}}, Undefined.String()},
		{"P2PK with pubkey equal to pubkey token identifier", args{scriptType: P2PK, pubKey: []byte(fmt.Sprintf("%d", 0))}, fmt.Sprintf("%cq OP_CHECKSIG", PubKey)},
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
