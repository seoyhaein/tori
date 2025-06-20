package db

import (
	"context"
	"database/sql"
	"fmt"
)

// SaveFolders rootPath 하위의 모든 Folder 에 대해 파일 정보를 DB에 삽입함.
func SaveFolders(ctx context.Context, db *sql.DB, rootPath string, foldersExclusions, filesExclusions []string) error {
	// rootPath 하위의 Folder 목록 조회
	folders, err := GetSubFolders(rootPath, foldersExclusions)
	if err != nil {
		return fmt.Errorf("failed to get subfolders from %s: %w", rootPath, err)
	}

	// 각 서브 Folder 에 대해 파일 정보를 DB에 삽입
	for _, folder := range folders {
		err = StoreFilesFolderInfo(ctx, db, folder.Path, filesExclusions)
		if err != nil {
			return fmt.Errorf("failed to load files info for folder %s: %w", folder.Path, err)
		}
	}

	return nil
}

// UpdateDB 폴더 변경 내역과 파일 변경 내역을 DB에 반영
func UpdateDB(ctx context.Context, db *sql.DB, diffs []FolderDiff, changes []FileChange) error {
	// 폴더 변경 업데이트
	if err := UpsertFolders(ctx, db, diffs); err != nil {
		return err
	}
	// UpsertFolders 해줘야지만, db 에 folderId 가 생겨서 검색할 수 가 있음.
	for i := range changes {
		folderId, err := getFolderID(db, changes[i].Path)
		if err != nil {
			return fmt.Errorf("failed to get folder ID for path %q: %w", changes[i].Path, err)
		}
		changes[i].FolderID = folderId
	}
	// 파일 변경 업데이트,
	if err := UpsertDelFiles(ctx, db, changes); err != nil {
		return err
	}
	return nil
}

// DiffFolders 폴더 파일 비교
func DiffFolders(db *sql.DB) ([][]string, []FolderDiff, []FileChange, error) {
	// 1. 폴더 비교: 디스크 폴더들과 db의 폴더 목록을 비교
	_, folders, folderDiffs, err := CompareFolders(db, gConfig.RootDir, nil, gConfig.Exclusions)
	if err != nil {
		return nil, nil, nil, err
	}

	var (
		folderFiles    [][]string
		allFileChanges []FileChange
	)

	// 2. 각 폴더에 대해 파일 비교
	for _, folder := range folders {
		filesMatch, files, fileChanges, err := CompareFiles(db, folder.Path, gConfig.Exclusions)
		if err != nil {
			return nil, nil, nil, err
		}
		if !filesMatch {
			allFileChanges = append(allFileChanges, fileChanges...)
		}

		fileNames := ExtractFileNames(files)
		folderFiles = append(folderFiles, append([]string{folder.Path}, fileNames...))
	}

	// 전체 동일 여부 판단: folderDiffs 와 allFileChanges 가 모두 비어 있으면 동일
	if len(folderDiffs) == 0 && len(allFileChanges) == 0 {
		return folderFiles, nil, nil, nil
	}

	return folderFiles, folderDiffs, allFileChanges, nil
}
