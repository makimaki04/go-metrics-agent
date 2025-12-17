package observer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

type Observer interface {
	Notify(ctx context.Context, event AuditEvent)
}

type AuditEvent struct {
	TimeStamp int      `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type FileObserver struct {
	FilePath string
	Logger   *zap.Logger
}

func (f *FileObserver) Notify(ctx context.Context, event AuditEvent) {
	file, err := os.OpenFile(f.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		f.Logger.Error("Failed to open audit file for writing", zap.Error(err))
		return
	}
	defer file.Close()

	data, err := json.MarshalIndent(event, "", "	")
	if err != nil {
		f.Logger.Error("Failed to mershal audit event", zap.Error(err))
		return
	}

	if _, err := file.Write(data); err != nil {
		f.Logger.Error("failed to write audit event to file", zap.Error(err))
		return
	}

	f.Logger.Info("audit event successfully added to the audit local file in ./data/audit_file.json")
}

type HTTPObserver struct {
	URL    string
	Logger *zap.Logger
}

func (h *HTTPObserver) Notify(ctx context.Context, event AuditEvent) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	body, err := json.Marshal(event)
	if err != nil {
		h.Logger.Error("Failed to create POST request", zap.Error(err))
		return
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		h.URL,
		bytes.NewReader(body),
	)
	if err != nil {
		h.Logger.Error("Failed to create POST request", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.Logger.Error("Failed to send POST request", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		h.Logger.Error("HTTP POST returned non-OK status", zap.Int("status", resp.StatusCode))
		return
	}

	h.Logger.Info("audit event successfully sent to HTTP server")
}

type contextKey string

const ReqIDKey contextKey = "reqID"
