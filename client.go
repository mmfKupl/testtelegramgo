package testtelegramgo

import (
	"fmt"

	"github.com/Arman92/go-tdlib"
)

type AppClient struct {
	// TODO: make fields private
	TdClient  *tdlib.Client
	AppConfig *AppConfig
}

func GetAppClient(configName string) (*AppClient, error) {
	config, err := GetAppConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("Fail to get config: %s. ", err)
	}

	tdlib.SetLogVerbosityLevel(1)
	tdlib.SetFilePath("./errors.txt")

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

	appClient := &AppClient{
		TdClient:  client,
		AppConfig: config,
	}

	return appClient, nil
}

func (appClient *AppClient) StartAppClient() error {
	err := appClient.initAppClient()
	if err != nil {
		return fmt.Errorf("Fail to init appClient: %v. ", err)
	}

	fmt.Println("Authorization Ready! Let's rock")
	return nil
}

func (appClient *AppClient) initAppClient() error {
	tdClient := appClient.TdClient
	appConfig := appClient.AppConfig

	for {
		currentState, _ := tdClient.Authorize()
		if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPhoneNumberType {
			_, err := tdClient.SendPhoneNumber(appConfig.Tel)
			if err != nil {
				return fmt.Errorf("Fail to send phone number: %v. ", err)
			}
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitCodeType {
			fmt.Print("Enter code: ")
			var code string
			_, err := fmt.Scanln(&code)
			if err != nil {
				return fmt.Errorf("Fail to enter code: %v. ", err)
			}
			_, err = tdClient.SendAuthCode(code)
			if err != nil {
				fmt.Printf("Fail to send auth code : %v", err)
			}
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateWaitPasswordType {
			_, err := tdClient.SendAuthPassword(appConfig.Password)
			if err != nil {
				fmt.Printf("Fail to send auth password : %v", err)
			}
		} else if currentState.GetAuthorizationStateEnum() == tdlib.AuthorizationStateReadyType {
			break
		}
	}

	return nil
}
