package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
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
	ctx, span := h.tracer.Start(r.Context(), "ExportHandler.DownloadExport")
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

	// Construct file path
	filePath := fmt.Sprintf("exports/%s", filename)

	// Get the file repository from export service (we need to add a method to access it)
	fileRepo := h.exportService.GetFileRepository()

	// Check if file exists
	exists, err := fileRepo.FileExists(ctx, filePath)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to check file existence", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Get file size
	fileSize, err := fileRepo.GetFileSize(ctx, filePath)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to get file size", http.StatusInternalServerError)
		return
	}

	// Get file data
	fileData, err := fileRepo.GetFile(ctx, filePath)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Determine content type based on file extension
	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(filename), ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		contentType = "text/csv"
	} else if strings.HasSuffix(strings.ToLower(filename), ".zip") {
		contentType = "application/zip"
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))

	// Write file data to response
	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}

// ListExports lists available export files
func (h *ExportHandler) ListExports(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ExportHandler.ListExports")
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

	// List files from export directory
	fileRepo := h.exportService.GetFileRepository()
	files, err := fileRepo.ListFiles(ctx, "exports")
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to list export files", http.StatusInternalServerError)
		return
	}

	// Convert files to export information
	var exports []map[string]interface{}
	for i := offset; i < len(files) && i < offset+limit; i++ {
		filePath := fmt.Sprintf("exports/%s", files[i])
		fileSize, _ := fileRepo.GetFileSize(ctx, filePath)

		// Determine format from filename
		format := "unknown"
		if strings.HasSuffix(strings.ToLower(files[i]), ".json") {
			format = "json"
		} else if strings.HasSuffix(strings.ToLower(files[i]), ".csv") {
			format = "csv"
		} else if strings.HasSuffix(strings.ToLower(files[i]), ".zip") {
			format = "zip"
		}

		// Parse timestamp from filename (otherside_export_YYYYMMDD_HHMMSS.ext)
		var generatedAt time.Time
		if parts := strings.Split(files[i], "_"); len(parts) >= 3 {
			timestampStr := strings.TrimSuffix(parts[2], filepath.Ext(parts[2]))
			if timestamp, err := time.Parse("20060102_150405", timestampStr); err == nil {
				generatedAt = timestamp
			}
		}

		if generatedAt.IsZero() {
			// Fallback to file modification time would go here in a real implementation
			generatedAt = time.Now()
		}

		exports = append(exports, map[string]interface{}{
			"filename":      files[i],
			"size":          fileSize,
			"format":        format,
			"session_count": 0, // Would be parsed from file metadata in a real implementation
			"generated_at":  generatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exports": exports,
		"limit":   limit,
		"offset":  offset,
		"total":   len(files),
	})
}

// DeleteExport deletes an export file
func (h *ExportHandler) DeleteExport(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ExportHandler.DeleteExport")
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

	// Construct file path
	filePath := fmt.Sprintf("exports/%s", filename)

	// Get the file repository
	fileRepo := h.exportService.GetFileRepository()

	// Check if file exists
	exists, err := fileRepo.FileExists(ctx, filePath)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to check file existence", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Delete the file
	if err := fileRepo.DeleteFile(ctx, filePath); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Export deleted successfully",
		"filename": filename,
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
