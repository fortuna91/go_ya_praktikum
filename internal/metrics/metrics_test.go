package metrics

import (
	"reflect"
	"testing"
)

func TestCalcHash(t *testing.T) {
	type args struct {
		metric *Metric
		key    string
	}
	tests := []struct {
		name     string
		args     args
		wantHash []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotHash := CalcHash(tt.args.metric, tt.args.key); !reflect.DeepEqual(gotHash, tt.wantHash) {
				t.Errorf("CalcHash() = %v, want %v", gotHash, tt.wantHash)
			}
		})
	}
}
