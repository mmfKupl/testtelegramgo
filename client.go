package testtelegramgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/Arman92/go-tdlib"
	"github.com/mmfKupl/gosse"
)

type AppClient struct {
	tdClient  *tdlib.Client
	appConfig *AppConfig
	messenger *Messenger
	notifier  gosse.INotifier
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
		tdClient:  client,
		appConfig: config,
		messenger: &Messenger{},
		notifier:  getNotifier(),
	}

	appClient.notifier.RegisterOnConnectionRegistered(func(c gosse.IClient, conn gosse.IConnection) {
		go func() {
			if conn != nil {
				conn.Notify(appClient.GetFormattedMessages(*appClient.messenger.messagesStore))
			} else {
				fmt.Printf("Current Connection Dosen't Exist - %v", c.GetId())
			}
		}()
	})

	return appClient, nil
}

func (appClient *AppClient) StartAppClient() error {
	mainContext, closeMainContext := context.WithCancel(context.Background())
	defer closeMainContext()

	err := appClient.initAppClient()
	if err != nil {
		return fmt.Errorf("Fail to init appClient: %v. ", err)
	}

	fmt.Println("Authorization Ready! Let's rock")
	appClient.initMessenger()
	defer func() {
		closeMainContext()
		appClient.messenger.Close()
	}()

	go func() {
		for updates := range appClient.messenger.MessageUpdates {
			log.Printf("Received %v messages from messanger \n", updates.TotalCount)
			appClient.notifier.NotifyAll(appClient.GetFormattedMessages(updates))
		}
	}()

	go func() {
		err = appClient.StartMessenger(mainContext)
		if err != nil {
			closeMainContext()
			panic(err)
		}
	}()

	http.Handle("/connect", appClient.notifier)
	http.HandleFunc("/pulse", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("pw")
		if header != appClient.appConfig.AdminPassword {
			http.Error(w, "{ \"error\": \"Forbidden\" }", http.StatusForbidden)
			return
		}
		nt := appClient.notifier.(*gosse.Notifier)
		clientsAmount := len(nt.GetClients())
		connectionsAmount := 0
		for _, connections := range nt.GetClientsConnections() {
			connectionsAmount += len(connections)
		}
		clientIds := make([]string, 0, clientsAmount)
		for _, client := range nt.GetClients() {
			clientIds = append(clientIds, client.GetId())
		}

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		b, err := json.Marshal(map[string]interface{}{
			"clientsAmount":     clientsAmount,
			"clientIds":         clientIds,
			"connectionsAmount": connectionsAmount,
			"alloc":             bToMb(m.Alloc),
			"totalAlloc":        bToMb(m.TotalAlloc),
			"sys":               bToMb(m.Sys),
			"munGC":             int(m.NumGC),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%v\" }", err), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(b)
		w.WriteHeader(http.StatusOK)
	})

	port := fmt.Sprintf(":%s", appClient.appConfig.Port)
	fmt.Printf("Server started on port %s\n", port)
	err = http.ListenAndServe(port, nil)
	return err
}
func bToMb(b uint64) string {
	return fmt.Sprintf("%v Mb", b/1024/1024)
}

func (appClient *AppClient) initAppClient() error {
	tdClient := appClient.tdClient
	appConfig := appClient.appConfig

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

func getNotifier() gosse.INotifier {
	notifier := &gosse.Notifier{}
	notifier.Init()
	notifier.RegisterClientIdentifier(clientIdentifier)
	return notifier
}

func clientIdentifier(r *http.Request) (string, error) {
	clientId := r.RemoteAddr
	forwardedFor := r.Header.Get("X-FORWARDED-FOR")
	if forwardedFor != "" {
		clientId = forwardedFor
	}
	if clientId == "" {
		return "", fmt.Errorf("Empty RemoteAddr recieved. ")
	}
	return clientId, nil
}
