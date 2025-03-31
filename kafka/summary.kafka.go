package summary

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/donghquinn/hls_converter/biz/converter"
	"github.com/donghquinn/hls_converter/configs"
	"github.com/segmentio/kafka-go"
)

type KafkaInterface struct {
	Con        *kafka.Conn
	Broker     string
	Topic      string
	GroupId    string
	ConsumerId string
}

func KafkaInstance() KafkaInterface {
	kafkaConfig := configs.KafkaConfig

	conn2, err2 := kafka.Dial("tcp", kafkaConfig.Broker)
	if err2 != nil {
		log.Printf("failed to dial: %v", err2)
		return KafkaInterface{}
	}
	if err := conn2.Close(); err != nil {
		log.Printf("failed to close writer: %v", err)
		return KafkaInterface{}
	}
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaConfig.Broker, kafkaConfig.Topic, 0)
	if err != nil {
		log.Printf("failed to dial leader: %v", err)
		return KafkaInterface{}
	}

	// conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// _, err = conn.WriteMessages(
	// 	kafka.Message{Value: []byte("one!")},
	// 	kafka.Message{Value: []byte("two!")},
	// 	kafka.Message{Value: []byte("three!")},
	// )
	// if err != nil {
	// 	log.Printf("failed to write messages: %v", err)
	// 	return KafkaInterface{}
	// }

	conn.CreateTopics(kafka.TopicConfig{
		Topic:             kafkaConfig.Topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})

	// if err := conn.Close(); err != nil {
	// 	log.Printf("failed to close writer: %v", err)
	// 	return KafkaInterface{}
	// }

	instance := KafkaInterface{
		Con:        conn,
		Broker:     kafkaConfig.Broker,
		Topic:      kafkaConfig.Topic,
		GroupId:    kafkaConfig.GroupId,
		ConsumerId: kafkaConfig.ConsumerId,
	}

	return instance
}

func (k *KafkaInterface) Consume() {
	// make a new reader that consumes from topic-A
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{k.Broker},
		GroupID:        k.GroupId,
		Topic:          k.Topic,
		MaxBytes:       10e6, // 10MB
		CommitInterval: 0,    // 자동 커밋 비활성화
	})

	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("[KAFKA] failed to close reader: %v", err)
		}
	}()

	for {
		ctx := context.Background()
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("[KAFKA] ReadMessage Error: %v", err)
			// 에러가 지속되면 break 혹은 재시도 로직 추가
			continue
		}

		if string(m.Key) == "" || string(m.Value) == "" {
			log.Println("[KAFKA] No Key or Value found in message")
			// 잘못된 메시지는 건너뛰고 커밋할 수도 있음
			if commitErr := r.CommitMessages(ctx, m); commitErr != nil {
				log.Printf("[KAFKA] Commit failed for invalid message: %v", commitErr)
			}
			continue
		}

		// 요청 처리를 위한 딜레이 (예시로 2분 딜레이)
		log.Printf("[KAFKA] Received message: %s, sleeping for 2 minutes", string(m.Value))
		time.Sleep(2 * time.Minute)

		// key: requestId, value: referenceSummarySeq (예시)
		requestId := string(m.Key)
		referenceSummarySeq, convErr := strconv.Atoi(string(m.Value))
		if convErr != nil {
			log.Printf("[KAFKA] Convert String to Integer Error: %v", convErr)
			// 형변환 실패 시, 메시지 커밋 여부를 결정: 재처리할 것인지 DLQ에 넣을 것인지 선택
			continue
		}

		// 요약 처리 호출
		if err := converter.ConvertToHLS(&converter.ConversionJob{
			ID:          fmt.Sprintf("%d", referenceSummarySeq),
			InputFile:   requestId,
			OutputDir:   requestId + ".m3u8",
			Status:      "completed",
			CreatedAt:   time.Now(),
			CompletedAt: time.Now(),
			Error:       "",
		}); err != nil {
			log.Printf("[KAFKA] Summary Run Error: %v", err)
			// 에러 발생 시, 해당 메시지를 재처리하거나 DLQ로 보내는 로직 추가 가능
			// 여기서는 커밋하지 않으므로 재처리 대상이 됩니다.
			continue
		}

		// 요약 처리 성공 시, 오프셋 수동 커밋
		if commitErr := r.CommitMessages(ctx, m); commitErr != nil {
			log.Printf("[KAFKA] Commit Message Error: %v", commitErr)
			// 커밋 실패 시 재처리 로직 고려
		} else {
			log.Printf("[KAFKA] Message processed and committed at topic/partition/offset %v/%v/%v: %s = %s",
				m.Topic, m.Partition, m.Offset, requestId, string(m.Value))
		}
	}
}
