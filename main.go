package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Arman92/go-tdlib"
)

const channelName string = "hack_pro_dota"

type appConfig struct {
	Password string `json:"password"`
	Tel      string `json:"tel"`
	ApiId    string `json:"apiId"`
	ApiHash  string `json:"apiHash"`
	ChatID   int64  `json:"chatID"`
}

func main() {

	tdlib.SetLogVerbosityLevel(1)
	tdlib.SetFilePath("./errors.txt")

	config := getAppConfig()

	client := tdlib.NewClient(tdlib.Config{
		APIID:               config.ApiId,
		APIHash:             config.ApiHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
		UseMessageDatabase:  true,
		UseFileDatabase:     false,
		UseChatInfoDatabase: false,
		UseTestDataCenter:   false,
		DatabaseDirectory:   "./tdlib-db",
		FileDirectory:       "./tdlib-files",
		IgnoreFileNames:     false,
	})

	for {
		currentState, _ := client.Authorize()
		if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPhoneNumberType {
			client.SendPhoneNumber(config.Tel)
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitCodeType {
			fmt.Print("Enter code: ")
			var code string
			fmt.Scanln(&code)
			_, err := client.SendAuthCode(code)
			if err != nil {
				fmt.Printf("Error sending auth code : %v", err)
			}
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPasswordType {
			client.SendAuthPassword(config.Password)
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateReadyType {
			fmt.Println("Authorization Ready! Let's rock")
			break
		}
	}

	lastMessage, err := getLastMessage(client, &config)
	if err != nil {
		panic(err)
	}
	lastMessages, err := getLastMessagesFromMessage(client, lastMessage, 10)
	if err != nil {
		panic(err)
	}

	fmt.Printf("TotalCount - %v\n", lastMessages.TotalCount)

	for _, message := range lastMessages.Messages {
		fmt.Println(message.Content.(*tdlib.MessageText).Text.Text)
		// fmt.Printf("%v - %v", message.Date, message.EditDate)
		fmt.Println("____________")
	}
}

func getAppConfig() appConfig {
	if _, err := os.Stat("./config.json"); err != nil {
		panic("config.json file not found.")
	}

	bytesConfig, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	var config appConfig
	err = json.Unmarshal(bytesConfig, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func getLastMessage(client *tdlib.Client, config *appConfig) (tdlib.Message, error) {
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
