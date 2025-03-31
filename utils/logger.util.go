package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const logDir = "logs"

var logFile *os.File

// 새로운 로그 파일을 생성하고 로그 출력 대상을 변경
func RotateLogFile() {
	CreateDir(logDir)
	// 기존 로그 파일 닫기
	if logFile != nil {
		logFile.Close()
	}

	// 로그 파일 경로 생성 (날짜별)
	fileName := fmt.Sprintf("log_%s.log", time.Now().Format("2006-01-02"))
	filePath := filepath.Join(logDir, fileName)

	// 로그 파일 생성
	var err error
	logFile, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("로그 파일을 열 수 없습니다: %v", err)
	}

	// 로그 출력 대상을 파일로 설정
	log.SetOutput(logFile)
}

// 로그 파일을 압축한 후 기존 로그 파일 삭제
func ArchiveAndDeleteLogFile(filePath string) error {
	// 압축 파일 이름 설정
	archivePath := fmt.Sprintf("%s.tar.gz", filePath)

	// tar.gz 파일 생성
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("tar.gz 파일 생성 실패: %v", err)
	}
	defer archiveFile.Close()

	gzipWriter := gzip.NewWriter(archiveFile)
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	defer gzipWriter.Close()

	// 로그 파일 열기
	logFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("로그 파일 열기 실패: %v", err)
	}
	defer logFile.Close()

	// 로그 파일 정보 읽기
	fileInfo, err := logFile.Stat()
	if err != nil {
		return fmt.Errorf("로그 파일 정보 읽기 실패: %v", err)
	}

	// tar 헤더 작성
	header := &tar.Header{
		Name: filepath.Base(filePath),
		Size: fileInfo.Size(),
		Mode: int64(fileInfo.Mode()),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("tar 헤더 작성 실패: %v", err)
	}

	// 파일 내용 복사
	if _, err := io.Copy(tarWriter, logFile); err != nil {
		return fmt.Errorf("로그 파일 압축 실패: %v", err)
	}

	// 로그 파일 삭제
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("로그 파일 삭제 실패: %v", err)
	}

	return nil
}

// 매일 자정마다 로그 파일을 교체하고 압축하는 함수
func ScheduleLogRotation() {
	ticker := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-ticker.C:
			// 자정에 기존 로그 파일 압축 및 삭제
			currentLogFilePath := logFile.Name()
			if err := ArchiveAndDeleteLogFile(currentLogFilePath); err != nil {
				log.Printf("로그 파일 압축 및 삭제 중 오류 발생: %v", err)
			}

			// 새로운 로그 파일 생성
			RotateLogFile()
		}
	}
}
