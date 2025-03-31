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

const (
	logDir    = "logs"
	activeLog = "app.log" // 현재 쓰고 있는 로그 파일 이름
)

var logFile *os.File

// 초기 로그 파일 생성 (프로그램 시작 시 한 번 호출)
func InitLog() {
	// logs 디렉토리가 없는 경우 생성
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("로그 디렉토리 생성 실패: %v", err)
	}

	// 이미 열려있는 logFile이 있다면 닫기
	if logFile != nil {
		logFile.Close()
	}

	// logs/app.log 파일을 열거나 생성 (추가 모드)
	filePath := filepath.Join(logDir, activeLog)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("로그 파일을 열 수 없습니다: %v", err)
	}

	logFile = f
	// 표준 라이브러리 log의 출력 대상을 이 파일로 지정
	log.SetOutput(logFile)
	log.Println("=== 로그 초기화 완료 ===")
}

// 로그 로테이션: 하루가 지난 활성 로그 파일을 날짜 기반 이름으로 변경 후 압축
func rotateLog() {
	// 1) 현재 열려있는 로그 파일 닫기
	if logFile != nil {
		logFile.Close()
	}

	// 2) 활성 로그 파일을 날짜 기반 이름으로 변경
	//    예: logs/app_2025-02-28.log
	today := time.Now().Format("2006-01-02")
	oldPath := filepath.Join(logDir, activeLog)
	newPath := filepath.Join(logDir, fmt.Sprintf("app_%s.log", today))
	if err := os.Rename(oldPath, newPath); err != nil {
		// 만약 Rename 실패 시 다시 logFile을 열어 로그가 누락되지 않도록 복구
		// (예외 상황에 따라 에러 처리를 충분히 해주는 것이 좋습니다)
		log.Println("로그 파일 이름 변경 실패:", err)
		// 새 파일이라도 열어줌
		InitLog()
		return
	}

	// 3) 새 로그 파일 다시 연다 (logs/app.log)
	InitLog()

	// 4) 날짜별 로그 파일을 압축 후 원본 삭제
	if err := ArchiveAndDeleteLogFile(newPath); err != nil {
		log.Printf("로그 파일 압축 및 삭제 실패: %v", err)
	}
}

// tar.gz로 압축 후 원본 삭제
func ArchiveAndDeleteLogFile(filePath string) error {
	archivePath := filePath + ".tar.gz"

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
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("로그 파일 열기 실패: %v", err)
	}
	defer f.Close()

	// 파일 정보
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("로그 파일 정보 읽기 실패: %v", err)
	}

	// tar 헤더
	header := &tar.Header{
		Name: filepath.Base(filePath),
		Size: fi.Size(),
		Mode: int64(fi.Mode()),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("tar 헤더 작성 실패: %v", err)
	}

	// 실제 파일 데이터를 tar에 기록
	if _, err := io.Copy(tarWriter, f); err != nil {
		return fmt.Errorf("로그 파일 압축 실패: %v", err)
	}

	// 원본 로그 삭제
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("로그 파일 삭제 실패: %v", err)
	}
	return nil
}

// 일정 주기로 (24시간 등) 로그 로테이션 수행
func ScheduleLogRotation() {
	// 처음 시작 시 남아있던 로그 파일 없으면 새로 생성
	InitLog()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		<-ticker.C
		// 매 24시간마다 rotateLog() 수행
		rotateLog()
	}
}
