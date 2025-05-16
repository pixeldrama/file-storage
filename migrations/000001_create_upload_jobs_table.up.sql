CREATE TABLE IF NOT EXISTS upload_jobs (
    id VARCHAR(36) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    file_id VARCHAR(36),
    error TEXT
); 