package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/myideascope/otherside/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ExportHandler handles HTTP requests for data export
type ExportHandler struct {
	exportService *service.ExportService
	tracer        trace.Tracer
}

// NewExportHandler creates a new export handler
func NewExportHandler(exportService *service.ExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
		tracer:        otel.Tracer("otherside/export"),
	}
}

// ExportSessions exports session data
func (h *ExportHandler) ExportSessions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ExportHandler.ExportSessions")
	defer span.End()

	var req service.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.SessionIDs) == 0 {
		http.Error(w, "At least one session ID must be specified", http.StatusBadRequest)
		return
	}

	// Set defaults if not specified
	if req.Format == "" {
		req.Format = service.ExportFormatJSON
	}

	span.SetAttributes(
		attribute.Int("export.session_count", len(req.SessionIDs)),
		attribute.String("export.format", string(req.Format)),
		attribute.Bool("export.include_audio", req.IncludeAudio),
	)

	// Perform export
	result, err := h.exportService.ExportSessions(ctx, req)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return export result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// DownloadExport downloads an export file
func (h *ExportHandler) DownloadExport(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "ExportHandler.DownloadExport")
	defer span.End()

	vars := mux.Vars(r)
	filename := vars["filename"]

	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("export.filename", filename))

	// Security: Validate filename to prevent path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// This would typically serve the file from storage
	// For now, return a placeholder response
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// In a real implementation, you would:
	// 1. Check if file exists in storage
	// 2. Stream the file to the response
	// 3. Handle appropriate content types

	http.Error(w, "Export download not yet implemented", http.StatusNotImplemented)
}

// ListExports lists available export files
func (h *ExportHandler) ListExports(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "ExportHandler.ListExports")
	defer span.End()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	span.SetAttributes(
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	// This would typically list export files from storage
	// For now, return a placeholder response
	exports := []map[string]interface{}{
		{
			"filename":      "otherside_export_20240312_143022.json",
			"size":          1024000,
			"format":        "json",
			"session_count": 3,
			"generated_at":  time.Now().Add(-24 * time.Hour),
		},
		{
			"filename":      "otherside_export_20240311_092145.zip",
			"size":          2048000,
			"format":        "zip",
			"session_count": 5,
			"generated_at":  time.Now().Add(-48 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exports": exports,
		"limit":   limit,
		"offset":  offset,
		"total":   len(exports),
	})
}

// DeleteExport deletes an export file
func (h *ExportHandler) DeleteExport(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "ExportHandler.DeleteExport")
	defer span.End()

	vars := mux.Vars(r)
	filename := vars["filename"]

	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("export.filename", filename))

	// Security: Validate filename
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// This would typically delete the file from storage
	// For now, return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Export deleted successfully",
	})
}

// RegisterRoutes registers export-related routes
func (h *ExportHandler) RegisterRoutes(r *mux.Router) {
	// Export operations
	r.HandleFunc("/api/v1/export/sessions", h.ExportSessions).Methods("POST")
	r.HandleFunc("/api/v1/export/list", h.ListExports).Methods("GET")
	r.HandleFunc("/api/v1/export/download/{filename}", h.DownloadExport).Methods("GET")
	r.HandleFunc("/api/v1/export/delete/{filename}", h.DeleteExport).Methods("DELETE")
}
