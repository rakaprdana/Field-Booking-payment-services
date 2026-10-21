package kafka

type KafkaRegistry struct {
	brokers []string
}

// GetKafkaProducer implements [IKafkaRegistry].
func (k *KafkaRegistry) GetKafkaProducer() IKafka {
	return NewKafkaProducer(k.brokers)
}

type IKafkaRegistry interface {
	GetKafkaProducer() IKafka
}

func NewKafkaRegistry(brokers []string) IKafkaRegistry {
	return &KafkaRegistry{
		brokers: brokers,
	}
}
