package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/myideascope/otherside/internal/domain"
)

// ExportService handles data export functionality
type ExportService struct {
	sessionRepo     domain.SessionRepository
	evpRepo         domain.EVPRepository
	voxRepo         domain.VOXRepository
	radarRepo       domain.RadarRepository
	slsRepo         domain.SLSRepository
	interactionRepo domain.InteractionRepository
	fileRepo        domain.FileRepository
}

// ExportFormat represents different export formats
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatZIP  ExportFormat = "zip"
)

// ExportRequest contains export parameters
type ExportRequest struct {
	SessionIDs   []string     `json:"session_ids"`
	Format       ExportFormat `json:"format"`
	IncludeAudio bool         `json:"include_audio"`
	IncludeVideo bool         `json:"include_video"`
	DateFrom     *time.Time   `json:"date_from,omitempty"`
	DateTo       *time.Time   `json:"date_to,omitempty"`
	IncludeEVPs  bool         `json:"include_evps"`
	IncludeVOX   bool         `json:"include_vox"`
	IncludeRadar bool         `json:"include_radar"`
	IncludeSLS   bool         `json:"include_sls"`
	IncludeNotes bool         `json:"include_notes"`
}

// ExportResult contains export results and metadata
type ExportResult struct {
	Filename     string    `json:"filename"`
	Size         int64     `json:"size"`
	Format       string    `json:"format"`
	SessionCount int       `json:"session_count"`
	ItemCount    int       `json:"item_count"`
	GeneratedAt  time.Time `json:"generated_at"`
	FilePath     string    `json:"file_path"`
}

// NewExportService creates a new export service
func NewExportService(
	sessionRepo domain.SessionRepository,
	evpRepo domain.EVPRepository,
	voxRepo domain.VOXRepository,
	radarRepo domain.RadarRepository,
	slsRepo domain.SLSRepository,
	interactionRepo domain.InteractionRepository,
	fileRepo domain.FileRepository,
) *ExportService {
	return &ExportService{
		sessionRepo:     sessionRepo,
		evpRepo:         evpRepo,
		voxRepo:         voxRepo,
		radarRepo:       radarRepo,
		slsRepo:         slsRepo,
		interactionRepo: interactionRepo,
		fileRepo:        fileRepo,
	}
}

// ExportSessions exports session data in the specified format
func (s *ExportService) ExportSessions(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Validate request
	if len(req.SessionIDs) == 0 {
		return nil, fmt.Errorf("no sessions specified for export")
	}

	// Collect session data
	sessionData := make(map[string]*SessionExportData)

	for _, sessionID := range req.SessionIDs {
		data, err := s.collectSessionData(ctx, sessionID, req)
		if err != nil {
			return nil, fmt.Errorf("failed to collect data for session %s: %w", sessionID, err)
		}
		sessionData[sessionID] = data
	}

	// Generate export based on format
	var exportData []byte
	var filename string
	var err error

	switch req.Format {
	case ExportFormatJSON:
		exportData, filename, err = s.exportAsJSON(sessionData)
	case ExportFormatCSV:
		exportData, filename, err = s.exportAsCSV(sessionData, req)
	case ExportFormatZIP:
		exportData, filename, err = s.exportAsZIP(ctx, sessionData, req)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("export generation failed: %w", err)
	}

	// Save export file
	filePath := filepath.Join("exports", filename)
	if err := s.fileRepo.SaveFile(ctx, filePath, exportData); err != nil {
		return nil, fmt.Errorf("failed to save export file: %w", err)
	}

	// Calculate statistics
	totalItems := 0
	for _, data := range sessionData {
		totalItems += len(data.EVPs) + len(data.VOXEvents) + len(data.RadarEvents) +
			len(data.SLSDetections) + len(data.Interactions)
	}

	return &ExportResult{
		Filename:     filename,
		Size:         int64(len(exportData)),
		Format:       string(req.Format),
		SessionCount: len(sessionData),
		ItemCount:    totalItems,
		GeneratedAt:  time.Now(),
		FilePath:     filePath,
	}, nil
}

// SessionExportData contains all data for a session export
type SessionExportData struct {
	Session       *domain.Session           `json:"session"`
	EVPs          []*domain.EVPRecording    `json:"evps,omitempty"`
	VOXEvents     []*domain.VOXEvent        `json:"vox_events,omitempty"`
	RadarEvents   []*domain.RadarEvent      `json:"radar_events,omitempty"`
	SLSDetections []*domain.SLSDetection    `json:"sls_detections,omitempty"`
	Interactions  []*domain.UserInteraction `json:"interactions,omitempty"`
}

func (s *ExportService) collectSessionData(ctx context.Context, sessionID string, req ExportRequest) (*SessionExportData, error) {
	data := &SessionExportData{}

	// Get session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	data.Session = session

	// Collect EVP data
	if req.IncludeEVPs {
		evps, err := s.evpRepo.GetBySessionID(ctx, sessionID)
		if err == nil {
			data.EVPs = evps
		}
	}

	// Collect VOX data
	if req.IncludeVOX {
		voxEvents, err := s.voxRepo.GetBySessionID(ctx, sessionID)
		if err == nil {
			data.VOXEvents = voxEvents
		}
	}

	// Collect Radar data
	if req.IncludeRadar {
		radarEvents, err := s.radarRepo.GetBySessionID(ctx, sessionID)
		if err == nil {
			data.RadarEvents = radarEvents
		}
	}

	// Collect SLS data
	if req.IncludeSLS {
		slsDetections, err := s.slsRepo.GetBySessionID(ctx, sessionID)
		if err == nil {
			data.SLSDetections = slsDetections
		}
	}

	// Collect Interactions
	if req.IncludeNotes {
		interactions, err := s.interactionRepo.GetBySessionID(ctx, sessionID)
		if err == nil {
			data.Interactions = interactions
		}
	}

	return data, nil
}

func (s *ExportService) exportAsJSON(sessionData map[string]*SessionExportData) ([]byte, string, error) {
	// Create export structure
	export := struct {
		ExportInfo struct {
			Version     string    `json:"version"`
			GeneratedAt time.Time `json:"generated_at"`
			Format      string    `json:"format"`
			Application string    `json:"application"`
		} `json:"export_info"`
		Sessions map[string]*SessionExportData `json:"sessions"`
	}{
		Sessions: sessionData,
	}

	export.ExportInfo.Version = "1.0.0"
	export.ExportInfo.GeneratedAt = time.Now()
	export.ExportInfo.Format = "json"
	export.ExportInfo.Application = "OtherSide Paranormal Investigation"

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("JSON marshaling failed: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("otherside_export_%s.json", timestamp)

	return data, filename, nil
}

func (s *ExportService) exportAsCSV(sessionData map[string]*SessionExportData, req ExportRequest) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	defer writer.Flush()

	// Write sessions CSV
	if err := s.writeSessionsCSV(writer, sessionData); err != nil {
		return nil, "", fmt.Errorf("failed to write sessions CSV: %w", err)
	}

	// Write EVPs CSV if requested
	if req.IncludeEVPs {
		writer.Write([]string{}) // Empty line
		writer.Write([]string{"EVP RECORDINGS"})
		if err := s.writeEVPsCSV(writer, sessionData); err != nil {
			return nil, "", fmt.Errorf("failed to write EVPs CSV: %w", err)
		}
	}

	// Write VOX Events CSV if requested
	if req.IncludeVOX {
		writer.Write([]string{}) // Empty line
		writer.Write([]string{"VOX COMMUNICATIONS"})
		if err := s.writeVOXCSV(writer, sessionData); err != nil {
			return nil, "", fmt.Errorf("failed to write VOX CSV: %w", err)
		}
	}

	// Write Radar Events CSV if requested
	if req.IncludeRadar {
		writer.Write([]string{}) // Empty line
		writer.Write([]string{"RADAR DETECTIONS"})
		if err := s.writeRadarCSV(writer, sessionData); err != nil {
			return nil, "", fmt.Errorf("failed to write radar CSV: %w", err)
		}
	}

	// Write SLS Detections CSV if requested
	if req.IncludeSLS {
		writer.Write([]string{}) // Empty line
		writer.Write([]string{"SLS DETECTIONS"})
		if err := s.writeSLSCSV(writer, sessionData); err != nil {
			return nil, "", fmt.Errorf("failed to write SLS CSV: %w", err)
		}
	}

	// Write Interactions CSV if requested
	if req.IncludeNotes {
		writer.Write([]string{}) // Empty line
		writer.Write([]string{"INTERACTIONS"})
		if err := s.writeInteractionsCSV(writer, sessionData); err != nil {
			return nil, "", fmt.Errorf("failed to write interactions CSV: %w", err)
		}
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("otherside_export_%s.csv", timestamp)

	return buf.Bytes(), filename, nil
}

func (s *ExportService) exportAsZIP(ctx context.Context, sessionData map[string]*SessionExportData, req ExportRequest) ([]byte, string, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	defer zipWriter.Close()

	// Add JSON export to ZIP
	jsonData, _, err := s.exportAsJSON(sessionData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JSON for ZIP: %w", err)
	}

	jsonFile, err := zipWriter.Create("sessions.json")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create JSON file in ZIP: %w", err)
	}
	jsonFile.Write(jsonData)

	// Add CSV export to ZIP
	csvData, _, err := s.exportAsCSV(sessionData, req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate CSV for ZIP: %w", err)
	}

	csvFile, err := zipWriter.Create("sessions.csv")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create CSV file in ZIP: %w", err)
	}
	csvFile.Write(csvData)

	// Add individual session files
	for sessionID, data := range sessionData {
		sessionDir := fmt.Sprintf("sessions/%s/", sessionID)

		// Session summary
		summaryData, _ := json.MarshalIndent(data.Session, "", "  ")
		summaryFile, _ := zipWriter.Create(sessionDir + "summary.json")
		summaryFile.Write(summaryData)

		// Add audio files if requested and available
		if req.IncludeAudio {
			s.addAudioFilesToZip(ctx, zipWriter, sessionDir, data.EVPs)
		}
	}

	// Add README
	readmeContent := s.generateReadme(req)
	readmeFile, _ := zipWriter.Create("README.txt")
	readmeFile.Write([]byte(readmeContent))

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("otherside_export_%s.zip", timestamp)

	return buf.Bytes(), filename, nil
}

// CSV writing helper methods
func (s *ExportService) writeSessionsCSV(writer *csv.Writer, sessionData map[string]*SessionExportData) error {
	// Write header
	header := []string{
		"Session ID", "Title", "Location", "Venue", "Start Time", "End Time",
		"Status", "Temperature", "Humidity", "Pressure", "EMF Level", "Notes",
	}
	writer.Write(header)

	// Write data
	for _, data := range sessionData {
		session := data.Session
		record := []string{
			session.ID,
			session.Title,
			session.Location.Address,
			session.Location.Venue,
			session.StartTime.Format(time.RFC3339),
			formatTimePtr(session.EndTime),
			string(session.Status),
			fmt.Sprintf("%.1f", session.Environmental.Temperature),
			fmt.Sprintf("%.1f", session.Environmental.Humidity),
			fmt.Sprintf("%.2f", session.Environmental.Pressure),
			fmt.Sprintf("%.2f", session.Environmental.EMFLevel),
			session.Notes,
		}
		writer.Write(record)
	}

	return nil
}

func (s *ExportService) writeEVPsCSV(writer *csv.Writer, sessionData map[string]*SessionExportData) error {
	// Write header
	header := []string{
		"Session ID", "EVP ID", "Timestamp", "Duration", "Quality",
		"Detection Level", "File Path", "Annotations",
	}
	writer.Write(header)

	// Write data
	for sessionID, data := range sessionData {
		for _, evp := range data.EVPs {
			annotations := ""
			if len(evp.Annotations) > 0 {
				annotationsData, _ := json.Marshal(evp.Annotations)
				annotations = string(annotationsData)
			}

			record := []string{
				sessionID,
				evp.ID,
				evp.Timestamp.Format(time.RFC3339),
				fmt.Sprintf("%.2f", evp.Duration),
				string(evp.Quality),
				fmt.Sprintf("%.3f", evp.DetectionLevel),
				evp.FilePath,
				annotations,
			}
			writer.Write(record)
		}
	}

	return nil
}

func (s *ExportService) writeVOXCSV(writer *csv.Writer, sessionData map[string]*SessionExportData) error {
	// Write header
	header := []string{
		"Session ID", "VOX ID", "Timestamp", "Generated Text", "Phonetic Bank",
		"Trigger Strength", "Language Pack", "User Response", "Response Delay",
	}
	writer.Write(header)

	// Write data
	for sessionID, data := range sessionData {
		for _, vox := range data.VOXEvents {
			record := []string{
				sessionID,
				vox.ID,
				vox.Timestamp.Format(time.RFC3339),
				vox.GeneratedText,
				vox.PhoneticBank,
				fmt.Sprintf("%.3f", vox.TriggerStrength),
				vox.LanguagePack,
				vox.UserResponse,
				fmt.Sprintf("%.2f", vox.ResponseDelay),
			}
			writer.Write(record)
		}
	}

	return nil
}

func (s *ExportService) writeRadarCSV(writer *csv.Writer, sessionData map[string]*SessionExportData) error {
	// Write header
	header := []string{
		"Session ID", "Radar ID", "Timestamp", "Position X", "Position Y",
		"Strength", "Source Type", "EMF Reading", "Audio Anomaly", "Duration",
	}
	writer.Write(header)

	// Write data
	for sessionID, data := range sessionData {
		for _, radar := range data.RadarEvents {
			record := []string{
				sessionID,
				radar.ID,
				radar.Timestamp.Format(time.RFC3339),
				fmt.Sprintf("%.2f", radar.Position.X),
				fmt.Sprintf("%.2f", radar.Position.Y),
				fmt.Sprintf("%.3f", radar.Strength),
				string(radar.SourceType),
				fmt.Sprintf("%.2f", radar.EMFReading),
				fmt.Sprintf("%.3f", radar.AudioAnomaly),
				fmt.Sprintf("%.2f", radar.Duration),
			}
			writer.Write(record)
		}
	}

	return nil
}

func (s *ExportService) writeSLSCSV(writer *csv.Writer, sessionData map[string]*SessionExportData) error {
	// Write header
	header := []string{
		"Session ID", "SLS ID", "Timestamp", "Confidence", "Skeletal Points Count",
		"Bounding Box Width", "Bounding Box Height", "Duration", "Movement Pattern",
	}
	writer.Write(header)

	// Write data
	for sessionID, data := range sessionData {
		for _, sls := range data.SLSDetections {
			record := []string{
				sessionID,
				sls.ID,
				sls.Timestamp.Format(time.RFC3339),
				fmt.Sprintf("%.3f", sls.Confidence),
				strconv.Itoa(len(sls.SkeletalPoints)),
				fmt.Sprintf("%.1f", sls.BoundingBox.Width),
				fmt.Sprintf("%.1f", sls.BoundingBox.Height),
				fmt.Sprintf("%.2f", sls.Duration),
				sls.Movement.Pattern,
			}
			writer.Write(record)
		}
	}

	return nil
}

func (s *ExportService) writeInteractionsCSV(writer *csv.Writer, sessionData map[string]*SessionExportData) error {
	// Write header
	header := []string{
		"Session ID", "Interaction ID", "Timestamp", "Type", "Content", "Response", "Response Time",
	}
	writer.Write(header)

	// Write data
	for sessionID, data := range sessionData {
		for _, interaction := range data.Interactions {
			record := []string{
				sessionID,
				interaction.ID,
				interaction.Timestamp.Format(time.RFC3339),
				string(interaction.Type),
				interaction.Content,
				interaction.Response,
				fmt.Sprintf("%.2f", interaction.ResponseTime),
			}
			writer.Write(record)
		}
	}

	return nil
}

func (s *ExportService) addAudioFilesToZip(ctx context.Context, zipWriter *zip.Writer, sessionDir string, evps []*domain.EVPRecording) {
	audioDir := sessionDir + "audio/"

	for _, evp := range evps {
		if evp.FilePath != "" {
			audioData, err := s.fileRepo.GetFile(ctx, evp.FilePath)
			if err != nil {
				// Log error but continue with other files
				continue
			}

			// Create filename based on EVP ID and original extension
			filename := fmt.Sprintf("evp_%s%s", evp.ID, filepath.Ext(evp.FilePath))
			if filename == "evp_"+evp.ID {
				filename += ".wav" // Default extension if none found
			}

			// Create file in ZIP
			audioFile, err := zipWriter.Create(audioDir + filename)
			if err != nil {
				continue
			}

			// Write audio data to ZIP
			audioFile.Write(audioData)

			// Also add processed audio if available
			if evp.ProcessedPath != "" {
				processedData, err := s.fileRepo.GetFile(ctx, evp.ProcessedPath)
				if err == nil {
					processedFilename := fmt.Sprintf("evp_%s_processed%s", evp.ID, filepath.Ext(evp.ProcessedPath))
					processedFile, _ := zipWriter.Create(audioDir + processedFilename)
					processedFile.Write(processedData)
				}
			}
		}
	}
}

func (s *ExportService) generateReadme(req ExportRequest) string {
	return fmt.Sprintf(`OtherSide Paranormal Investigation Export
Generated: %s

This export contains paranormal investigation data in multiple formats:
- sessions.json: Complete session data in JSON format
- sessions.csv: Tabular data suitable for spreadsheet applications

Export Configuration:
- Include EVPs: %t
- Include VOX: %t
- Include Radar: %t
- Include SLS: %t
- Include Notes: %t
- Include Audio Files: %t
- Include Video Files: %t

File Formats:
- JSON files contain complete structured data
- CSV files contain tabular data for analysis
- Individual session folders contain detailed breakdowns

For questions or support, visit: https://github.com/myideascope/otherside

Generated by OtherSide Paranormal Investigation App v1.0.0
`, time.Now().Format(time.RFC3339),
		req.IncludeEVPs, req.IncludeVOX, req.IncludeRadar,
		req.IncludeSLS, req.IncludeNotes, req.IncludeAudio, req.IncludeVideo)
}

// GetFileRepository returns the file repository for direct file operations
func (s *ExportService) GetFileRepository() domain.FileRepository {
	return s.fileRepo
}

// Helper functions
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
