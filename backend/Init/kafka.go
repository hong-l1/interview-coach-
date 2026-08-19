package Init

import "github.com/IBM/sarama"

func NewKafka() sarama.Client {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	client, err := sarama.NewClient([]string{"43.142.57.35:9094"}, config)
	if err != nil {
		panic(err)
	}
	return client
}
func NewKafkaProducer(client sarama.Client) sarama.SyncProducer {
	syncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return syncProducer
}
func NewConsumer(groupId string, client sarama.Client) sarama.ConsumerGroup {
	saramaConsumer, err := sarama.NewConsumerGroupFromClient(groupId, client)
	if err != nil {
		panic(err)
	}
	return saramaConsumer
}
