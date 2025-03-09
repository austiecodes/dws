package services

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// sendMsgToNode
func sendMsgToNode(ch *amqp.Channel, routingKey string, message string) {
	err := ch.Publish(
		"direct_nodes", // 交换机名称
		routingKey,     // 路由键（目标从节点标识）
		false,          // 是否强制
		false,          // 是否立即
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(message),
			DeliveryMode: amqp.Persistent, // 消息持久化
		},
	)
	if err != nil {
		log.Fatalf("Failed to send message to %s: %v", routingKey, err)
	}
	log.Printf("Sent to %s: %s", routingKey, message)
}
