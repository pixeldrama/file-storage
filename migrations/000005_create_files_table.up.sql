CREATE TABLE file_info (
    id VARCHAR(255) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    file_type VARCHAR(100) NOT NULL,
    linked_resource_type VARCHAR(100) NOT NULL,
    linked_resource_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_file_info_linked_resource ON file_info(linked_resource_type, linked_resource_id);
CREATE INDEX idx_file_info_file_type ON file_info(file_type); 