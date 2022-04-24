package metrics

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestCalcHash(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metric
		delta    int64
		key      string
		wantHash string
	}{
		{
			name:     "Common",
			metric:   Metric{ID: "GetSetZip245", MType: Counter},
			delta:    int64(984847306),
			key:      "key",
			wantHash: "63cef738a3d8967b03333150bda021bf2ba250f535bdcd323a931db0a4a47874",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.metric.Delta = &tt.delta
			if gotHash := CalcHash(&tt.metric, tt.key); !reflect.DeepEqual(hex.EncodeToString(gotHash), tt.wantHash) {
				t.Errorf("CalcHash() = %v, want %v", gotHash, tt.wantHash)
			}
		})
	}
}
