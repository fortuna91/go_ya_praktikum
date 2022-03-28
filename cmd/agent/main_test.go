package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSendRequest(t *testing.T) {
	type args struct {
		metricType string
		metricName string
		value      string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "case 1. Alloc 200",
			args: args{metricType: "gauge", metricName: "Alloc", value: "123"},
			want: 200,
		},
		{
			name: "case 2. BuckHashSys 200",
			args: args{metricType: "gauge", metricName: "BuckHashSys", value: "123.0"},
			want: 200,
		},
		{
			name: "case 3. Wrong metric",
			args: args{metricType: "gauge", metricName: "MyMetric", value: "123.0"},
			want: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// Test request parameters
				// equals(t, req.URL.String(), "/some/path")
				// Send response to be tested
				w.WriteHeader(http.StatusOK)
			}))
			// Close the server when test finishes
			defer server.Close()
			// client := http.Client{}
			request, _ := http.NewRequest(http.MethodPost, server.URL+"/update/"+tt.args.metricType+"/"+tt.args.metricName+"/"+tt.args.value, nil)

			responseCode := SendRequest(server.Client(), request)
			if responseCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", responseCode, tt.want)
			}
		})
	}
}

func TestGetMetrics(t *testing.T) {
	tests := []struct {
		name  string
		count int64
		want  int
	}{
		{
			name:  "common case",
			count: 1,
			want:  29,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMetrics(tt.count); !reflect.DeepEqual(len(got), tt.want) {
				t.Errorf("GetMetrics() = len %d, want %d", len(got), tt.want)
			}
		})
	}
}
