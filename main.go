package main

import (
	"github.com/seoyhaein/tori/cmd"
	globallog "github.com/seoyhaein/tori/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		globallog.Log.Fatalf("명령 실행 실패: %v", err)
	}
}
