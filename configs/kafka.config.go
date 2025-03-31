package configs

import "os"

type KafkaConf struct {
	Broker     string
	Topic      string
	GroupId    string
	ConsumerId string
}

var KafkaConfig KafkaConf

func SetKafkaConfig() {
	KafkaConfig.Broker = os.Getenv("KAFKA_BROKER")
	KafkaConfig.Topic = os.Getenv("KAFKA_TOPIC")
	KafkaConfig.GroupId = os.Getenv("KAFKA_GROUP")
	KafkaConfig.ConsumerId = os.Getenv("KAFKA_CONSUMER")
}
