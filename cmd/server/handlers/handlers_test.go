package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetGaгgeMetric(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want int
	}{
		{
			name: "case 1. Alloc 200",
			url:  "gauge/Alloc/123",
			want: 200,
		},
		{
			name: "case 2. BuckHashSys 200",
			url:  "gauge/BuckHashSys/123.0",
			want: 200,
		},
		{
			name: "case 3. Other metric",
			url:  "gauge/MyMetric/123.0",
			want: 200,
		},
		{
			name: "case 4. Wrong value",
			url:  "gauge/BuckHashSys/abc",
			want: 400,
		},
		{
			name: "case 5. Wrong value",
			url:  "gauge/BuckHashSys/",
			want: 404,
		},
		{
			name: "case 6. Wrong value",
			url:  "gauge/BuckHashSys",
			want: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/update/"+tt.url, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(SetGaugeMetric)
			h.ServeHTTP(w, request)
			response := w.Result()
			if response.StatusCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", response.StatusCode, tt.want)
			}
			defer response.Body.Close()
		})
	}
}

func TestSetCountMetric(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		count           int64
		currentCountVal int64
		want            int
	}{
		{
			name:            "case 1. Positive",
			url:             "counter/PollCount/1",
			count:           1,
			currentCountVal: 0,
			want:            200,
		},
		{
			name:            "case 2. Other metric",
			url:             "counter/MyMetric/123",
			want:            200,
			count:           123,
			currentCountVal: 0,
		},
		{
			name: "case 3. Wrong value, empty",
			url:  "counter/BuckHashSys/",
			want: 404,
		},
		{
			name: "case 4. Wrong value, str",
			url:  "counter/PollCount/abc",
			want: 400,
		},
		{
			name: "case 5. Wrong value, float",
			url:  "counter/PollCount/123.0",
			want: 400,
		},
		{
			name: "case 6. Wrong value, empty",
			url:  "counter/BuckHashSys",
			want: 404,
		},
		{
			name:            "case 7. CurrCount",
			url:             "counter/PollCount/5",
			count:           5,
			currentCountVal: 2,
			want:            200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// go Counter(CountChannel)
			if tt.want == 200 {
				currCount = tt.currentCountVal
			}
			request := httptest.NewRequest(http.MethodPost, "/update/"+tt.url, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(SetCounterMetric)
			h.ServeHTTP(w, request)
			response := w.Result()
			if response.StatusCode != tt.want {
				t.Errorf("SendRequest() = %v, want %v", response.StatusCode, tt.want)
			}
			defer response.Body.Close()
			if tt.want == 200 {
				if currCount != tt.count+tt.currentCountVal {
					t.Errorf("Wrong currCount = %v, want %v", currCount, tt.count+tt.currentCountVal)
				}
			}
		})
	}
}

func TestNotImplemented(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/update/something", nil)
	w := httptest.NewRecorder()
	h := http.HandlerFunc(NotImplemented)
	h.ServeHTTP(w, request)
	response := w.Result()
	defer response.Body.Close()
	if response.StatusCode != 501 {
		t.Errorf("SendRequest() = %v, want %v", response.StatusCode, 501)
	}
}
