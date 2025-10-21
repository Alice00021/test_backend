package utils

import (
	"errors"
	"strings"
)

const (
	serverName      = "server"
	clientName      = "client"
	routingKeyParts = 4
)

func ConvertRoutingKey(key string) (string, error) {
	parts := strings.Split(key, ".")
	if len(parts) < routingKeyParts {
		return "", errors.New("wrong routing key format or empty")
	}

	rearranged := []string{parts[2], parts[3], parts[0], parts[1]}
	return strings.Join(rearranged, "."), nil
}

func ConstructRoutingKey(senderName, receiverName string) (string, error) {
	if receiverName == "" || senderName == "" {
		return "", errors.New("empty senderName or receiverName")
	}
	keyArray := []string{senderName, senderName + "_" + clientName, receiverName, receiverName + "_" + serverName}
	return strings.Join(keyArray, "."), nil
}

func GetListenerQueueName(exchange, queueName string) string {
	if strings.Contains(exchange, serverName) {
		return queueName + "_" + serverName
	}
	if strings.Contains(exchange, clientName) {
		return queueName + "_" + clientName
	}

	return ""
}
