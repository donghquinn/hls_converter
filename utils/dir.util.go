package utils

import (
	"log"
	"os"
)

func CreateDir(dirPath string) {
	// 디렉토리가 존재하는지 확인
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 디렉토리가 존재하지 않으면 생성
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			log.Fatalf("Create Directory Error: %v", err)
		}
		return
	} else {
		return
	}
}
