package script //nolint:testpackage // don't create separate package for tests

import "testing"

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
		{"check regular script generation for P2PK", args{scriptType: P2PK, pubKey: []byte("public-key")}, "7KTAvebKjNUpoi OP_CHECKSIG"},
		{"generation with empty public key", args{scriptType: P2PK, pubKey: []byte{}}, Undefined.String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScript(tt.args.scriptType, tt.args.pubKey); got != tt.want {
				t.Errorf("NewScript() = %v, want %v", got, tt.want)
			}
		})
	}
}
