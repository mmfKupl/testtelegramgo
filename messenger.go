package testtelegramgo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Arman92/go-tdlib"
	"github.com/mmfKupl/gosse"
)

type Messenger struct {
	MessageUpdates chan tdlib.Messages
	messagesStore  *tdlib.Messages
}

type Message struct {
	ID   int64  `json:"id"`
	Date int32  `json:"date"`
	Text string `json:"text"`
}

type MessagesToSend struct {
	TotalAmount int32     `json:"totalAmount"`
	Messages    []Message `json:"messages"`
}

func (m Messenger) GetFormattedMessages(messages tdlib.Messages) gosse.Message {
	messagesToSend := MessagesToSend{
		TotalAmount: messages.TotalCount,
		Messages:    make([]Message, 0, 0),
	}

	for _, message := range messages.Messages {
		messageText := ""
		messageContent, ok := message.Content.(*tdlib.MessageText)
		if !ok {
			messageText = "Fail to get text"
		} else {
			messageText = messageContent.Text.Text
		}
		messagesToSend.Messages = append(messagesToSend.Messages, Message{
			ID:   message.ID,
			Date: message.Date,
			Text: messageText,
		})
	}

	b, err := json.Marshal(messagesToSend)
	if err != nil {
		return gosse.Message(fmt.Sprintf("{ \"error\": \"%v\" }", err))
	}

	return b
}

func (m *Messenger) Close() {
	close(m.MessageUpdates)
}

func (appClient *AppClient) initMessenger() {
	appClient.messenger.MessageUpdates = make(chan tdlib.Messages)
}

func (appClient *AppClient) StartMessenger(ctx context.Context) error {
	err := appClient.updateMessengerState()
	if err != nil {
		return err
	}

	const receiverChannelCapacity int = 5

	receiverForNewMessage := appClient.tdClient.AddEventReceiver(&tdlib.UpdateNewMessage{}, appClient.messagesFilter, receiverChannelCapacity)
	defer close(receiverForNewMessage.Chan)
	receiverForUpdatingMessageContent := appClient.tdClient.AddEventReceiver(&tdlib.UpdateMessageEdited{}, appClient.messagesFilter, receiverChannelCapacity)
	defer close(receiverForUpdatingMessageContent.Chan)

	for {
		select {
		case <-receiverForNewMessage.Chan:
			err := appClient.updateMessengerState()
			if err != nil {
				return err
			}
		case <-receiverForUpdatingMessageContent.Chan:
			err := appClient.updateMessengerState()
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (appClient *AppClient) messagesFilter(msg *tdlib.TdMessage) bool {
	messageType := tdlib.UpdateEnum((*msg).MessageType())
	var chatID int64

	if messageType == tdlib.UpdateNewMessageType {
		chatID = getCharIdFromUpdateNewMessage(msg)
	} else if messageType == tdlib.UpdateMessageEditedType {
		chatID = getChatIdFromUpdateMessageEdited(msg)
	} else {
		return false
	}

	if chatID == appClient.getChatId() {
		return true
	}
	return false

}

func getCharIdFromUpdateNewMessage(msg *tdlib.TdMessage) int64 {
	message := (*msg).(*tdlib.UpdateNewMessage)
	return message.Message.ChatID
}

func getChatIdFromUpdateMessageEdited(msg *tdlib.TdMessage) int64 {
	message := (*msg).(*tdlib.UpdateMessageEdited)
	return message.ChatID
}

func (appClient *AppClient) updateMessengerState() error {
	lastMessages, err := appClient.getLastMessages(appClient.appConfig.DisplayedMessagesAmount)
	if err != nil {
		return fmt.Errorf("Fail to get last messages: %s. ", err)
	}

	messenger := appClient.messenger

	messenger.messagesStore = lastMessages
	messenger.MessageUpdates <- *messenger.messagesStore

	return nil
}

func (appClient *AppClient) getLastMessages(amount uint) (*tdlib.Messages, error) {
	lastMessage, err := appClient.getLastMessage()
	if err != nil {
		return nil, err
	}
	return appClient.getLastMessagesFromMessage(lastMessage, amount)
}

func (appClient *AppClient) getLastMessage() (tdlib.Message, error) {
	messages, err := appClient.tdClient.GetChatHistory(appClient.getChatId(), 0, 0, 1, false)
	if err != nil {
		return tdlib.Message{}, err
	}

	if messages.TotalCount == 0 {
		return tdlib.Message{}, fmt.Errorf("Can't get message from the chat or there is no messages in the chat. ")
	}

	return messages.Messages[0], nil
}

func (appClient *AppClient) getLastMessagesFromMessage(fromMessage tdlib.Message, amount uint) (*tdlib.Messages, error) {
	if amount <= 1 {
		amount = 5
	}

	return appClient.tdClient.GetChatHistory(appClient.getChatId(), fromMessage.ID, -1, int32(amount), false)
}

func (appClient *AppClient) getChatId() int64 {
	return appClient.appConfig.ChatID
}
