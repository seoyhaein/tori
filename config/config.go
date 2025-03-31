package config

import (
	"encoding/json"
	"fmt"
	globallog "github.com/seoyhaein/tori/log"
	"os"
)

type Config struct {
	RootDir    string   `json:"rootDir"`    // lustre-client 마운트된 폴더로 사용할 예정.
	Exclusions []string `json:"exclusions"` // 예: ["*.json", "invalid_files", "*.csv", "*.pb"]
}

var (
	GlobalConfig *Config
	logger       = globallog.Log
)

func init() {
	// config 설정
	config, err := LoadConfig("config.json")
	// Important 기억하자. os.Exit(1) 로만 하지 말고 Log.Fatalf 를 써서 오류 사항을 명확히 하자. 자체적으로 os.Exit(1) 처리됨.
	if err != nil {
		logger.Fatalf("failed to load config file %v", err)
	}
	GlobalConfig = config
}

func LoadConfig(filename string) (cfg *Config, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	// defer 내에서 err 가 이미 설정되어 있지 않은 경우에만 파일 닫기 에러를 처리
	defer func() {
		if cErr := file.Close(); cErr != nil && err == nil {
			logger.Warnf("failed to close file: %v", cErr)
		}
	}()

	decoder := json.NewDecoder(file)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	// 필수 항목 검증
	if config.RootDir == "" {
		return nil, fmt.Errorf("missing 'rootDir' in configuration")
	}

	// Exclusions 가 비어있으면 기본값 설정
	if len(config.Exclusions) == 0 {
		config.Exclusions = []string{"*.json", "invalid_files", "*.csv", "*.pb"}
	}

	return &config, nil
}
