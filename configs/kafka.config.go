package configs

import "os"

type KafkaConf struct {
	Broker      string
	InputTopic  string
	OutputTopic string
	GroupId     string
	ConsumerId  string
}

var KafkaConfig KafkaConf

func SetKafkaConfig() {
	KafkaConfig.Broker = os.Getenv("KAFKA_BROKER")
	KafkaConfig.InputTopic = os.Getenv("KAFKA_INPUT_TOPIC")
	KafkaConfig.OutputTopic = os.Getenv("KAFKA_OUTPUT_TOPIC")
	KafkaConfig.GroupId = os.Getenv("KAFKA_GROUP")
	KafkaConfig.ConsumerId = os.Getenv("KAFKA_CONSUMER")
}
