package summary

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"org.donghyuns.com/reference/scheduler/configs"
	"org.donghyuns.com/reference/scheduler/database"
	"org.donghyuns.com/reference/scheduler/utils"
)

func Summary(requestId string, referenceSummarySeq int) error {
	const (
		maxAttempts  = 30              // 최대 시도 횟수 (필요에 따라 조정)
		pollInterval = 2 * time.Minute // polling 간격 (필요에 따라 조정)
	)

	var summaryResponse SummaryResultAPIBlogPostTypeResponse
	var attempt int

	for {
		attempt++
		log.Printf("[SUMMARY] Request ID: %s, Reference Seq : %d", requestId, referenceSummarySeq)
		response, responseErr := GetSummary(requestId, referenceSummarySeq)
		if responseErr != nil {
			return responseErr
		}

		// HTTP 상태 코드가 400인 경우 에러 응답 처리
		if response.StatusCode == http.StatusBadRequest {
			var errorResponse SummaryRequestErrorResponse
			// utils.ParseBody 대신 body를 한 번 읽어서 사용
			bodyBytes, err := io.ReadAll(response.Body)
			if err != nil {
				return err
			}
			response.Body.Close()
			if err := json.Unmarshal(bodyBytes, &errorResponse); err != nil {
				log.Printf("[SUMMARY] Parse Error Response Err: %v", err)
				return err
			}
			log.Printf("[SUMMARY] Got Error Response: %s", errorResponse.Error)
			return fmt.Errorf("%s", errorResponse.Error)
		}

		// 응답 Body를 한 번 읽어서 버퍼에 저장
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		response.Body.Close()

		// 먼저 status 필드만 파싱
		var statusResp SummaryResultAPIPollingResponse
		if err := json.Unmarshal(bodyBytes, &statusResp); err != nil {
			log.Printf("[SUMMARY] Parse Status Error: %v", err)
			return err
		}

		// "pending" 상태이면 polling 처리
		if statusResp.Status == "pending" {
			log.Printf("[SUMMARY] Request %s is still pending (attempt %d/%d)", requestId, attempt, maxAttempts)
			if attempt >= maxAttempts {
				return fmt.Errorf("summary request %s still pending after %d attempts", requestId, maxAttempts)
			}
			time.Sleep(pollInterval)
			continue
		}

		// 최종 응답 파싱 (bodyBytes는 이미 메모리에 있으므로 재사용 가능)
		if err := json.Unmarshal(bodyBytes, &summaryResponse); err != nil {
			log.Printf("[SUMMARY] Parse Final Response Error: %v", err)
			return err
		}

		break
	}

	// 최종 응답 처리: 데이터 삽입
	if err := InsertSummaryData(fmt.Sprintf("%d", referenceSummarySeq), summaryResponse.Data.Data.BlogPost); err != nil {
		return err
	}

	return nil
}
func GetSummary(requestId string, referenceSummarySeq int) (*http.Response, error) {
	lilysConfig := configs.LilyConfig

	// 요청 생성
	request, requestErr := http.NewRequest("GET", fmt.Sprintf("%s/%s?resultType=blogPost", lilysConfig.LilysUrl, requestId), nil)

	if requestErr != nil {
		log.Printf("[SUMMARY] Request Chat Completion Error: %v", requestErr)
		return nil, requestErr
	}

	// 헤더 추가 및 auth코드 입력
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", lilysConfig.LilysApi))

	// 요청 클라이언트 인스턴스 생성
	client := &http.Client{}

	// 요청 실행
	response, responseErr := client.Do(request)
	if responseErr != nil {
		log.Printf("[SUMMARY] Send Request Err: %v", responseErr)

		// 에러 내용 update
		// updateRequestErr := UpdateSummaryStatus(referenceSummarySeq, "3")

		// if updateRequestErr != nil {
		// 	log.Printf("[CHAT] Update Request Error Response: %v", updateRequestErr)
		// 	return nil, updateRequestErr
		// }

		return nil, responseErr
	}

	return response, responseErr
}

func InsertSummaryData(referenceSeq string, summaryData []SummaryResultAPIBlogPostItem) error {
	dbCon, dbErr := database.InitPostgresConnection()

	if dbErr != nil {
		return dbErr
	}

	var queryString []string

	for _, data := range summaryData {
		query, genErr := utils.GenerateQueryString(InsertSummaryDataSql, referenceSeq, data.Title, data.Content, "1")

		if genErr != nil {
			log.Printf("[SUMMARY] Generate Updating Summary Error: %v", genErr)
			continue
		}

		log.Printf("[DEBUGGING] Query String: %s", query)

		queryString = append(queryString, query)
	}

	updateErr := dbCon.InsertMultiple(queryString)

	if updateErr != nil {
		return updateErr
	}

	return nil
}

func UpdateSummaryData(referenceSummarySeq int, summaryData []SummaryResultAPIBlogPostItem) error {
	dbCon, dbErr := database.InitPostgresConnection()

	if dbErr != nil {
		return dbErr
	}

	var queryString []string

	for _, data := range summaryData {
		query, genErr := utils.GenerateQueryString(UpdateSummaryDataSql, data.Title, data.Content, "1", fmt.Sprintf("%d", referenceSummarySeq))

		if genErr != nil {
			log.Printf("[SUMMARY] Generate Updating Summary Error: %v", genErr)
			continue
		}

		queryString = append(queryString, query)
	}

	updateErr := dbCon.InsertMultiple(queryString)

	if updateErr != nil {
		return updateErr
	}

	return nil
}
