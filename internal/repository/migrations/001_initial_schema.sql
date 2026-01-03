-- Database schema for OtherSide Paranormal Investigation Application

-- Sessions table - stores investigation sessions
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    location_latitude REAL,
    location_longitude REAL,
    location_address TEXT,
    location_description TEXT,
    location_venue TEXT,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    notes TEXT,
    env_temperature REAL,
    env_humidity REAL,
    env_pressure REAL,
    env_emf_level REAL,
    env_light_level REAL,
    env_noise_level REAL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- EVP Recordings table - stores EVP (Electronic Voice Phenomenon) recordings
CREATE TABLE IF NOT EXISTS evp_recordings (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    duration REAL NOT NULL,
    timestamp DATETIME NOT NULL,
    waveform_data TEXT, -- JSON array of waveform data
    processed_path TEXT,
    annotations TEXT, -- JSON array of annotations
    quality TEXT NOT NULL,
    detection_level REAL NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- VOX Events table - stores VOX communication events
CREATE TABLE IF NOT EXISTS vox_events (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    generated_text TEXT NOT NULL,
    phonetic_bank TEXT NOT NULL,
    frequency_data TEXT, -- JSON array of frequency data
    trigger_strength REAL NOT NULL,
    language_pack TEXT NOT NULL,
    modulation_type TEXT NOT NULL,
    user_response TEXT,
    response_delay REAL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Radar Events table - stores radar/presence detection events
CREATE TABLE IF NOT EXISTS radar_events (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    position_x REAL NOT NULL,
    position_y REAL NOT NULL,
    position_z REAL,
    strength REAL NOT NULL,
    source_type TEXT NOT NULL,
    emf_reading REAL NOT NULL,
    audio_anomaly REAL NOT NULL,
    duration REAL NOT NULL,
    movement_trail TEXT, -- JSON array of coordinate positions
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- SLS Detections table - stores Structured Light Sensor detections
CREATE TABLE IF NOT EXISTS sls_detections (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    skeletal_points TEXT NOT NULL, -- JSON array of skeletal points
    confidence REAL NOT NULL,
    bounding_box_top_left_x REAL NOT NULL,
    bounding_box_top_left_y REAL NOT NULL,
    bounding_box_bottom_right_x REAL NOT NULL,
    bounding_box_bottom_right_y REAL NOT NULL,
    bounding_box_width REAL NOT NULL,
    bounding_box_height REAL NOT NULL,
    video_frame TEXT,
    filter_applied TEXT, -- JSON array of applied filters
    duration REAL NOT NULL,
    movement_speed REAL,
    movement_direction REAL,
    movement_pattern TEXT,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- User Interactions table - stores user interactions during investigation
CREATE TABLE IF NOT EXISTS user_interactions (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    audio_path TEXT,
    response TEXT,
    response_time REAL,
    randomizer_type TEXT,
    randomizer_result TEXT,
    randomizer_range TEXT,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Files table - stores file metadata for session recordings and media
CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_type TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type TEXT,
    checksum TEXT,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);
CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);

CREATE INDEX IF NOT EXISTS idx_evp_session_id ON evp_recordings(session_id);
CREATE INDEX IF NOT EXISTS idx_evp_timestamp ON evp_recordings(timestamp);
CREATE INDEX IF NOT EXISTS idx_evp_quality ON evp_recordings(quality);
CREATE INDEX IF NOT EXISTS idx_evp_detection_level ON evp_recordings(detection_level);

CREATE INDEX IF NOT EXISTS idx_vox_session_id ON vox_events(session_id);
CREATE INDEX IF NOT EXISTS idx_vox_timestamp ON vox_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_vox_trigger_strength ON vox_events(trigger_strength);

CREATE INDEX IF NOT EXISTS idx_radar_session_id ON radar_events(session_id);
CREATE INDEX IF NOT EXISTS idx_radar_timestamp ON radar_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_radar_strength ON radar_events(strength);
CREATE INDEX IF NOT EXISTS idx_radar_source_type ON radar_events(source_type);

CREATE INDEX IF NOT EXISTS idx_sls_session_id ON sls_detections(session_id);
CREATE INDEX IF NOT EXISTS idx_sls_timestamp ON sls_detections(timestamp);
CREATE INDEX IF NOT EXISTS idx_sls_confidence ON sls_detections(confidence);

CREATE INDEX IF NOT EXISTS idx_interactions_session_id ON user_interactions(session_id);
CREATE INDEX IF NOT EXISTS idx_interactions_timestamp ON user_interactions(timestamp);
CREATE INDEX IF NOT EXISTS idx_interactions_type ON user_interactions(type);

CREATE INDEX IF NOT EXISTS idx_files_session_id ON files(session_id);
CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type);