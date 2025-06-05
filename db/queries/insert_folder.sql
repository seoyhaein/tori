INSERT INTO folders (path, total_size, file_count)
VALUES (?, ?, ?)
    ON CONFLICT(path) DO NOTHING;
