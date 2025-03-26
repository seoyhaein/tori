package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// TODO 업데이트 될 예정. MaxWatchCount 이거 필요 없어짐.
// TODO 수정 필요함.

type Config struct {
	MaxWatchCount int    `json:"MaxWatchCount"`
	RootDir       string `json:"rootDir"` // lustre-client 마운트된 폴더로 사용할 예정.
}

var GlobalConfig *Config

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if cErr := file.Close(); cErr != nil {
			err = fmt.Errorf("failed to close file: %w", cErr)
		}
	}()

	decoder := json.NewDecoder(file)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	// 추가적으로 필수 항목 검증
	if config.MaxWatchCount <= 0 {
		return nil, fmt.Errorf("missing or invalid 'MaxWatchCount' in configuration")
	}
	if config.RootDir == "" {
		return nil, fmt.Errorf("missing 'rootDir' in configuration")
	}

	return &config, nil
}
