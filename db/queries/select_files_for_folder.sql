SELECT f.id, f.folder_id, f.name, f.size, f.created_time
FROM files f
         JOIN folders fo ON f.folder_id = fo.id
WHERE fo.path = ?
