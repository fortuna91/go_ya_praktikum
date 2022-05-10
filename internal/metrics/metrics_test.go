package metrics

import (
	"reflect"
	"testing"
)

func TestCalcHash(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metric
		delta    int64
		value    float64
		key      string
		wantHash string
	}{
		{
			name:     "Counter",
			metric:   Metric{ID: "GetSetZip245", MType: Counter},
			delta:    int64(984847306),
			value:    float64(98484),
			key:      "key",
			wantHash: "f3ec2b4cacd0273d5fbe5360c2de2e2d8bed7a0d01f4a2b76d58fcf25c5c8bb7",
		},
		{
			name:     "Gauge",
			metric:   Metric{ID: "GetSetZip245", MType: Gauge},
			delta:    int64(984847306),
			value:    float64(98484),
			key:      "key",
			wantHash: "30cfa69e8d64adc8a1c979346c8806f94d32e38f016ad232404f897241c17314",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.metric.Delta = &tt.delta
			tt.metric.Value = &tt.value
			if gotHash := CalcHash(&tt.metric, tt.key); !reflect.DeepEqual(gotHash, tt.wantHash) {
				t.Errorf("CalcHash() = %v, want %v", gotHash, tt.wantHash)
			}
		})
	}
}
