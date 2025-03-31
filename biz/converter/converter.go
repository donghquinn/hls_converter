package converter

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 설정 구조체
type Config struct {
	Port            string `json:"port"`
	UploadDir       string `json:"upload_dir"`
	OutputDir       string `json:"output_dir"`
	SegmentDuration int    `json:"segment_duration"`
}

// 변환 작업 상태 구조체
type ConversionJob struct {
	ID          string    `json:"id"`
	InputFile   string    `json:"input_file"`
	OutputDir   string    `json:"output_dir"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
	OutputFile  string    `json:"output_file,omitempty"` // 추가: 생성된 m3u8 파일 경로
}

// 응답 구조체
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var (
	config Config
	jobs   = make(map[string]*ConversionJob)
)

// 비디오 파일 형식 검증
func isVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".mp4", ".mov", ".avi", ".mkv"}
	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// 파일명 인코딩 함수 - MD5 해시 사용
func encodeFileName(filename string) string {
	// 파일 확장자 제외한 기본 이름 추출
	baseName := filepath.Base(filename)
	baseNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// 타임스탬프 추가하여 고유성 보장
	nameToEncode := fmt.Sprintf("%s_%d", baseNameWithoutExt, time.Now().UnixNano())

	// MD5 해시 생성
	hasher := md5.New()
	hasher.Write([]byte(nameToEncode))
	encodedName := hex.EncodeToString(hasher.Sum(nil))

	return encodedName
}

// FFmpeg를 사용하여 HLS로 변환
func ConvertToHLS(job *ConversionJob) error {
	job.Status = "processing"

	// 원본 파일명에서 인코딩된 이름 생성
	encodedFileName := encodeFileName(job.InputFile)

	// 출력 파일 이름 구성
	m3u8FileName := fmt.Sprintf("%s.m3u8", encodedFileName)
	tsFilePattern := fmt.Sprintf("%s_%%03d.ts", encodedFileName)

	// 전체 경로 설정
	playlistPath := filepath.Join(job.OutputDir, m3u8FileName)
	segmentPath := filepath.Join(job.OutputDir, tsFilePattern)

	// 작업에 출력 파일 경로 저장
	job.OutputFile = playlistPath

	// 환경 변수에서 FFmpeg 경로 가져오기 또는 기본값 사용
	ffmpegPath := os.Getenv("FFMPEG_PATH")
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg" // 기본값
	}

	log.Printf("변환 시작 (Job %s): %s -> %s", job.ID, job.InputFile, playlistPath)

	// FFmpeg 명령 구성
	cmd := exec.Command(
		ffmpegPath,
		"-i", job.InputFile,
		"-profile:v", "baseline", // 호환성을 위한 프로파일
		"-level", "3.0",
		"-start_number", "0",
		"-hls_time", fmt.Sprintf("%d", config.SegmentDuration),
		"-hls_list_size", "0", // 모든 세그먼트를 플레이리스트에 유지
		"-f", "hls",
		"-hls_segment_filename", segmentPath,
		playlistPath,
	)

	// FFmpeg 명령 로깅
	log.Printf("FFmpeg 명령: %v", cmd.Args)

	// 명령 실행 및 오류 처리
	output, err := cmd.CombinedOutput()
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("FFmpeg 오류: %v\n%s", err, string(output))
		job.CompletedAt = time.Now()
		log.Printf("변환 실패 (Job %s): %v\n%s", job.ID, err, string(output))
		return err
	}

	// 변환 성공 처리
	job.Status = "completed"
	job.CompletedAt = time.Now()
	log.Printf("변환 완료 (Job %s): %s -> %s", job.ID, job.InputFile, playlistPath)

	// 업로드된 원본 파일 삭제 (선택적)
	// os.Remove(job.InputFile)
	return nil
}

// 설정 로드 함수
func LoadConfig(cfg Config) {
	config = cfg

	// 설정값 확인 로깅
	log.Printf("HLS 변환기 설정 로드: 세그먼트 길이 %d초", config.SegmentDuration)

	// 출력 디렉터리가 없으면 생성
	if config.OutputDir != "" {
		if _, err := os.Stat(config.OutputDir); os.IsNotExist(err) {
			os.MkdirAll(config.OutputDir, 0755)
			log.Printf("출력 디렉터리 생성: %s", config.OutputDir)
		}
	}
}
