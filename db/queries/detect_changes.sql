SELECT path, total_size, file_count,
       (SELECT IFNULL(SUM(size), 0) FROM files WHERE files.folder_id = folders.id) AS current_size,
       (SELECT COUNT(*) FROM files WHERE files.folder_id = folders.id) AS current_count
FROM folders
WHERE total_size <> current_size OR file_count <> current_count;
