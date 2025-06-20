ALTER TABLE upload_jobs 
ADD CONSTRAINT fk_upload_jobs_file_id 
FOREIGN KEY (file_id) REFERENCES file_info(id) ON DELETE CASCADE; 