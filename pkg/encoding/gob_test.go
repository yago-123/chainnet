package encoding //nolint:testpackage // don't create separate package for tests

import (
	"chainnet/pkg/kernel"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGobEncoder_DeserializeBlock(t *testing.T) {
	type fields struct {
		logger *logrus.Logger
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *kernel.Block
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gobenc := &GobEncoder{
				logger: tt.fields.logger,
			}
			got, err := gobenc.DeserializeBlock(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeserializeBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGobEncoder_SerializeBlock(t *testing.T) {
	type fields struct {
		logger *logrus.Logger
	}
	type args struct {
		b kernel.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gobenc := &GobEncoder{
				logger: tt.fields.logger,
			}
			got, err := gobenc.SerializeBlock(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SerializeBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGobEncoder(t *testing.T) {
	type args struct {
		logger *logrus.Logger
	}
	tests := []struct {
		name string
		args args
		want *GobEncoder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGobEncoder(tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGobEncoder() = %v, want %v", got, tt.want)
			}
		})
	}
}
