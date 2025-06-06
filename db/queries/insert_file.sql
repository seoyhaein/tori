INSERT INTO files (folder_id, name, size)
VALUES (?, ?, ?)
    ON CONFLICT(folder_id, name) DO NOTHING;

