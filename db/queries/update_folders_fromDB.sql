UPDATE folders
SET total_size = (SELECT IFNULL(SUM(files.size), 0) FROM files WHERE files.folder_id = folders.id),
    file_count = (SELECT COUNT(*) FROM files WHERE files.folder_id = folders.id)
WHERE id = ?;
