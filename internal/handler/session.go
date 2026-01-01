package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/myideascope/otherside/internal/domain"
	"github.com/myideascope/otherside/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SessionHandler handles HTTP requests for paranormal investigation sessions
type SessionHandler struct {
	sessionService *service.SessionService
	tracer         trace.Tracer
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		tracer:         otel.Tracer("otherside/handler"),
	}
}

// CreateSession creates a new paranormal investigation session
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.CreateSession")
	defer span.End()

	var req service.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("session.title", req.Title),
		attribute.String("session.location", req.Location.Address),
	)

	session, err := h.sessionService.CreateSession(ctx, req)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

// GetSession retrieves a session by ID
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.GetSession")
	defer span.End()

	vars := mux.Vars(r)
	sessionID := vars["id"]

	span.SetAttributes(attribute.String("session.id", sessionID))

	session, err := h.sessionService.GetSessionSummary(ctx, sessionID)
	if err != nil {
		span.RecordError(err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// ProcessEVP processes EVP audio data
func (h *SessionHandler) ProcessEVP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.ProcessEVP")
	defer span.End()

	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	span.SetAttributes(attribute.String("session.id", sessionID))

	// Parse multipart form for audio data
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Audio file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	span.SetAttributes(
		attribute.String("file.name", header.Filename),
		attribute.Int64("file.size", header.Size),
	)

	// Read audio data (simplified - would need proper audio file parsing)
	audioData := make([]byte, header.Size)
	_, err = file.Read(audioData)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to read audio data", http.StatusInternalServerError)
		return
	}

	// Convert audio data to float64 (simplified conversion)
	floatData := make([]float64, len(audioData))
	for i, b := range audioData {
		floatData[i] = float64(int8(b)) / 128.0
	}

	// Get annotations from form
	annotations := r.FormValue("annotations")
	var annotationList []string
	if annotations != "" {
		json.Unmarshal([]byte(annotations), &annotationList)
	}

	metadata := service.EVPMetadata{
		FilePath:    header.Filename,
		Annotations: annotationList,
	}

	evp, err := h.sessionService.ProcessEVPRecording(ctx, sessionID, floatData, metadata)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to process EVP: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(evp)
}

// GenerateVOX generates VOX communication
func (h *SessionHandler) GenerateVOX(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.GenerateVOX")
	defer span.End()

	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	span.SetAttributes(attribute.String("session.id", sessionID))

	var triggerData service.VOXTriggerData
	if err := json.NewDecoder(r.Body).Decode(&triggerData); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Float64("trigger.emf_anomaly", triggerData.EMFAnomaly),
		attribute.Float64("trigger.audio_anomaly", triggerData.AudioAnomaly),
		attribute.String("trigger.language_pack", triggerData.LanguagePack),
	)

	voxEvent, err := h.sessionService.GenerateVOXCommunication(ctx, sessionID, triggerData)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to generate VOX: %v", err), http.StatusInternalServerError)
		return
	}

	if voxEvent == nil {
		// No VOX generated due to insufficient trigger strength
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(voxEvent)
}

// ProcessRadar processes radar detection data
func (h *SessionHandler) ProcessRadar(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.ProcessRadar")
	defer span.End()

	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	span.SetAttributes(attribute.String("session.id", sessionID))

	var radarData service.RadarEventData
	if err := json.NewDecoder(r.Body).Decode(&radarData); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Float64("radar.strength", radarData.Strength),
		attribute.Float64("radar.emf_reading", radarData.EMFReading),
	)

	radarEvent, err := h.sessionService.ProcessRadarEvent(ctx, sessionID, radarData)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to process radar event: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(radarEvent)
}

// ProcessSLS processes SLS detection data
func (h *SessionHandler) ProcessSLS(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.ProcessSLS")
	defer span.End()

	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	span.SetAttributes(attribute.String("session.id", sessionID))

	var slsData service.SLSDetectionData
	if err := json.NewDecoder(r.Body).Decode(&slsData); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Float64("sls.confidence", slsData.Confidence),
		attribute.Int("sls.skeletal_points", len(slsData.SkeletalPoints)),
	)

	slsDetection, err := h.sessionService.ProcessSLSDetection(ctx, sessionID, slsData)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to process SLS detection: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(slsDetection)
}

// RecordInteraction records user interaction
func (h *SessionHandler) RecordInteraction(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.RecordInteraction")
	defer span.End()

	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	span.SetAttributes(attribute.String("session.id", sessionID))

	var interactionData service.UserInteractionData
	if err := json.NewDecoder(r.Body).Decode(&interactionData); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("interaction.type", string(interactionData.Type)),
		attribute.String("interaction.content", interactionData.Content),
	)

	interaction, err := h.sessionService.RecordUserInteraction(ctx, sessionID, interactionData)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to record interaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(interaction)
}

// ListSessions lists all sessions with pagination
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.ListSessions")
	defer span.End()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	status := r.URL.Query().Get("status")

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
		attribute.String("filter.status", status),
	)

	// Get sessions from repository
	var sessions []*service.SessionSummary
	var total int

	if status != "" {
		// Filter by status
		var sessionStatus domain.SessionStatus
		switch status {
		case "active":
			sessionStatus = domain.SessionStatusActive
		case "complete":
			sessionStatus = domain.SessionStatusComplete
		case "paused":
			sessionStatus = domain.SessionStatusPaused
		case "archived":
			sessionStatus = domain.SessionStatusArchived
		default:
			http.Error(w, "Invalid status filter", http.StatusBadRequest)
			return
		}

		// Get sessions by status
		sessionList, err := h.sessionService.GetSessionsByStatus(ctx, sessionStatus, limit, offset)
		if err != nil {
			span.RecordError(err)
			http.Error(w, "Failed to list sessions", http.StatusInternalServerError)
			return
		}

		// Convert to session summaries
		for _, session := range sessionList {
			summary, err := h.sessionService.GetSessionSummary(ctx, session.ID)
			if err != nil {
				// Skip sessions that can't be summarized
				continue
			}
			sessions = append(sessions, summary)
		}

		// Get total count (this would need a CountByStatus method)
		total = len(sessions) // Placeholder
	} else {
		// Get all sessions
		sessionList, err := h.sessionService.GetAllSessions(ctx, limit, offset)
		if err != nil {
			span.RecordError(err)
			http.Error(w, "Failed to list sessions", http.StatusInternalServerError)
			return
		}

		// Convert to session summaries
		for _, session := range sessionList {
			summary, err := h.sessionService.GetSessionSummary(ctx, session.ID)
			if err != nil {
				// Skip sessions that can't be summarized
				continue
			}
			sessions = append(sessions, summary)
		}

		// Get total count (this would need a Count method)
		total = len(sessions) // Placeholder
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"limit":    limit,
		"offset":   offset,
		"total":    total,
	})
}

// GetSessionEvents gets all events for a session
func (h *SessionHandler) GetSessionEvents(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "SessionHandler.GetSessionEvents")
	defer span.End()

	vars := mux.Vars(r)
	sessionID := vars["sessionId"]
	eventType := r.URL.Query().Get("type") // evp, vox, radar, sls, interaction

	span.SetAttributes(
		attribute.String("session.id", sessionID),
		attribute.String("event.type", eventType),
	)

	summary, err := h.sessionService.GetSessionSummary(ctx, sessionID)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Failed to get session events: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter events based on type
	result := make(map[string]interface{})

	switch eventType {
	case "evp":
		result["events"] = summary.EVPs
	case "vox":
		result["events"] = summary.VOXEvents
	case "radar":
		result["events"] = summary.RadarEvents
	case "sls":
		result["events"] = summary.SLSDetections
	case "interaction":
		result["events"] = summary.Interactions
	default:
		// Return all events
		result["evps"] = summary.EVPs
		result["vox_events"] = summary.VOXEvents
		result["radar_events"] = summary.RadarEvents
		result["sls_detections"] = summary.SLSDetections
		result["interactions"] = summary.Interactions
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HealthCheck provides a health check endpoint
func (h *SessionHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "SessionHandler.HealthCheck")
	defer span.End()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": r.Context().Value("request_time"),
		"service":   "otherside-api",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// RegisterRoutes registers all session-related routes
func (h *SessionHandler) RegisterRoutes(r *mux.Router) {
	// Session management
	r.HandleFunc("/api/v1/sessions", h.CreateSession).Methods("POST")
	r.HandleFunc("/api/v1/sessions", h.ListSessions).Methods("GET")
	r.HandleFunc("/api/v1/sessions/{id}", h.GetSession).Methods("GET")

	// Paranormal investigation features
	r.HandleFunc("/api/v1/sessions/{sessionId}/evp", h.ProcessEVP).Methods("POST")
	r.HandleFunc("/api/v1/sessions/{sessionId}/vox", h.GenerateVOX).Methods("POST")
	r.HandleFunc("/api/v1/sessions/{sessionId}/radar", h.ProcessRadar).Methods("POST")
	r.HandleFunc("/api/v1/sessions/{sessionId}/sls", h.ProcessSLS).Methods("POST")
	r.HandleFunc("/api/v1/sessions/{sessionId}/interactions", h.RecordInteraction).Methods("POST")

	// Event retrieval
	r.HandleFunc("/api/v1/sessions/{sessionId}/events", h.GetSessionEvents).Methods("GET")

	// Health check
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")
}

// CORS middleware
func (h *SessionHandler) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
