package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/makimaki04/go-metrics-agent.git/internal/repository"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandler_PostMetric(t *testing.T) {
	type want struct {
		code        int
		contentType string
		response    string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "positive gauge test",
			request: "/update/gauge/CMD/20.55",
			want: want{
				code:        200,
				contentType: "text/plain",
				response:    "",
			},
		},
		{
			name:    "negative gauge test without ID",
			request: "/update/gauge//20.55",
			want: want{
				code:     404,
				response: `{"error": "metric ID is required"}`,
			},
		},
		{
			name:    "negative gauge test with wrong value",
			request: "/update/gauge/CMD/ABC",
			want: want{
				code:     400,
				response: `{"error": "invalid metric's value"}`,
			},
		},
		{
			name:    "positive counter test",
			request: "/update/counter/CMD/20",
			want: want{
				code:        200,
				contentType: "text/plain",
				response:    "",
			},
		},
		{
			name:    "negative counter test without ID",
			request: "/update/counter//20",
			want: want{
				code:     404,
				response: `{"error": "metric ID is required"}`,
			},
		},
		{
			name:    "negative counter test with wrong value",
			request: "/update/counter/CMD/10.66",
			want: want{
				code:     400,
				response: `{"error": "invalid delta value"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewStorage()
			service := service.NewService(storage, &zap.Logger{})
			handler := NewHandler(service, "")

			r := chi.NewRouter()
			r.Post("/update/{MType}/{ID}/{value}", handler.HandleReq)

			req := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestHandler_GetAllMetrics(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Get all metrics simple test",
			request: "/",
			want: want{
				code:        200,
				contentType: "text/html; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewStorage()
			service := service.NewService(storage, &zap.Logger{})
			handler := NewHandler(service, "")

			service.UpdateCounter("PollCount", 1)

			r := chi.NewRouter()
			r.Get("/", handler.GetAllMetrics)

			req := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			res := w.Result()

			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestHandler_GetMetric(t *testing.T) {
	type want struct {
		code int
		contentType string
		response string
	}
	tests := []struct {
		name string
		request string
		want want
	}{
		{
			name: "Get metric positive test",
			request: "/value/counter/PollCount",
			want: want{
				code: 200,
				contentType: "text/plain",
				response: "1",
			},
		},
		{
			name: "Get metric negative test",
			request: "/value/counter/SomeMetric",
			want: want{
				code: 404,
				response: `{"error": "invalid metric"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewStorage()
			service := service.NewService(storage, &zap.Logger{})
			handler := NewHandler(service, "")
			
			service.UpdateCounter("PollCount", 1)

			r := chi.NewRouter()
			r.Get("/value/{MType}/{ID}", handler.GetMetric)

			req := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			res := w.Result()

			require.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.response, string(body))
		})
	}
}
