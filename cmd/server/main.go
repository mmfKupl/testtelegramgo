package main

import (
	"fmt"

	"github.com/Arman92/go-tdlib"
	"github.com/mmfkupl/testtelegramgo"
)

func main() {
	appClient, err := testtelegramgo.GetAppClient("config.json")
	if err != nil {
		panic(err)
	}

	err = appClient.StartAppClient()
	if err != nil {
		panic(err)
	}

	// lastMessage, err := getLastMessage(client, config)
	// if err != nil {
	// 	panic(err)
	// }
	// lastMessages, err := getLastMessagesFromMessage(client, lastMessage, 10)
	// if err != nil {
	// 	panic(err)
	// }
	//
	// fmt.Printf("TotalCount - %v\n", lastMessages.TotalCount)
	//
	// for _, message := range lastMessages.Messages {
	// 	fmt.Println(message.Content.(*tdlib.MessageText).Text.Text)
	// 	// fmt.Printf("%v - %v", message.Date, message.EditDate)
	// 	fmt.Println("____________")
	// }
}

func getLastMessage(client *tdlib.Client, config *testtelegramgo.AppConfig) (tdlib.Message, error) {
	messages, err := client.GetChatHistory(config.ChatID, 0, 0, 1, false)
	if err != nil {
		return tdlib.Message{}, err
	}

	if messages.TotalCount == 0 {
		return tdlib.Message{}, fmt.Errorf("Can't get message from the chat or there is no messages in the chat. ")
	}

	return messages.Messages[0], nil
}

func getLastMessagesFromMessage(client *tdlib.Client, fromMessage tdlib.Message, amount uint) (*tdlib.Messages, error) {
	if amount == 0 {
		amount = 5
	}
	return client.GetChatHistory(fromMessage.ChatID, fromMessage.ID, -1, int32(amount), false)
}
