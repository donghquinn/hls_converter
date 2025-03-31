package converter

import (
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

// // 파일 업로드 및 변환 요청 처리
// func handleConvert(w http.ResponseWriter, r *http.Request) {
// 	// 파일 업로드 처리
// 	err := r.ParseMultipartForm(32 << 20) // 32MB 제한
// 	if err != nil {
// 		sendErrorResponse(w, "파일 파싱 오류", http.StatusBadRequest)
// 		return
// 	}

// 	file, handler, err := r.FormFile("file")
// 	if err != nil {
// 		sendErrorResponse(w, "파일을 찾을 수 없습니다", http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	// 파일 확장자 검증
// 	if !isVideoFile(handler.Filename) {
// 		sendErrorResponse(w, "지원하지 않는 파일 형식입니다. MP4, MOV, AVI 또는 MKV 파일만 지원합니다", http.StatusBadRequest)
// 		return
// 	}

// 	// 업로드 파일 저장
// 	jobID := uuid.New().String()
// 	uploadPath := filepath.Join(config.UploadDir, jobID+filepath.Ext(handler.Filename))
// 	outDir := filepath.Join(config.OutputDir, jobID)

// 	// 출력 디렉토리 생성
// 	os.MkdirAll(outDir, 0755)

// 	// 업로드된 파일 저장
// 	dst, err := os.Create(uploadPath)
// 	if err != nil {
// 		sendErrorResponse(w, "파일 저장 오류", http.StatusInternalServerError)
// 		return
// 	}
// 	defer dst.Close()

// 	// 업로드된 파일을 저장
// 	file.Seek(0, 0)
// 	_, err = file.WriteTo(dst)
// 	if err != nil {
// 		sendErrorResponse(w, "파일 저장 오류", http.StatusInternalServerError)
// 		return
// 	}

// 	// 새로운 변환 작업 생성
// 	job := &ConversionJob{
// 		ID:        jobID,
// 		InputFile: uploadPath,
// 		OutputDir: outDir,
// 		Status:    "pending",
// 		CreatedAt: time.Now(),
// 	}
// 	jobs[jobID] = job

// 	// 비동기로 변환 작업 시작
// 	go convertToHLS(job)

// 	// 응답 반환
// 	response := Response{
// 		Success: true,
// 		Message: "변환 작업이 시작되었습니다",
// 		Data: map[string]string{
// 			"job_id":     jobID,
// 			"status_url": fmt.Sprintf("/jobs/%s", jobID),
// 			"stream_url": fmt.Sprintf("/outputs/%s/playlist.m3u8", jobID),
// 		},
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusAccepted)
// 	json.NewEncoder(w).Encode(response)
// }

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

// FFmpeg를 사용하여 HLS로 변환
func ConvertToHLS(job *ConversionJob) error {
	job.Status = "processing"

	playlistPath := filepath.Join(job.OutputDir, "playlist.m3u8")
	segmentPath := filepath.Join(job.OutputDir, "segment_%03d.ts")

	// 환경 변수에서 FFmpeg 경로 가져오기 또는 기본값 사용
	ffmpegPath := os.Getenv("FFMPEG_PATH")
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg" // 기본값
	}

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
