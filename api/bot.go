package api

import (
	"bytes"
	"discocheckbot/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf16"
)

// telegram bot API parameters
const (
	CommandEntity     string = "bot_command"
	CrossedEntity     string = "strikethrough"
	BoldEntity        string = "bold"
	apiMethodTemplate string = "https://api.telegram.org/bot<TOKEN>/<METHOD>"
	apiFileTemplate   string = "https://api.telegram.org/file/bot<TOKEN>/<PATH>"
	UpdatesLimit      int    = 100 //todo move in config
	WaitForRetry      int    = 3
)

type BotImplementation interface {
	OnMessage(bot *Bot, msg *Message) error
	OnCallbackQuery(bot *Bot, cbq *CallbackQuery) error
}

type Bot struct {
	token          string
	updatesOffset  int
	httpTimeout    int //seconds
	log            *log.Logger
	implementation BotImplementation
}

func NewBot(cfg *config.ConfigReader, log *log.Logger, impl BotImplementation) *Bot {
	var token string
	var httpTimeout float64

	//reading config
	err := cfg.GetParameter("bot_token", &token)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = cfg.GetParameter("long_polling_timeout", &httpTimeout) //json number interprets as float64!
	if err != nil {
		log.Println(err)
		return nil
	}
	//initializing bot
	bot := Bot{
		token,
		0,
		int(httpTimeout),
		log,
		impl,
	}
	//checking existence of such bot
	_, err = makeApiRequest(bot.prepareApiUrl("getMe", ""),
		"GET",
		"",
		nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &bot
}

func (this *Bot) ListenForUpdates() {
	for {
		requestBody := RequestUpdates{
			this.updatesOffset,
			UpdatesLimit,
			this.httpTimeout,
			[]string{"message", "callback_query"},
		}
		log.Printf("requesting updates from %d", this.updatesOffset)
		updates, err := callApiMethod[RequestUpdates, []Update](this.prepareApiUrl("getUpdates", ""), requestBody)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * time.Duration(WaitForRetry))
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= this.updatesOffset {
				if update.CallbackQuery != nil {
					if err = this.implementation.OnCallbackQuery(this, update.CallbackQuery); err != nil {
						this.log.Printf("BOT ERROR: update %+v\n%s", update.CallbackQuery, err.Error())
					} else {
						this.log.Printf("BOT INFO: update %+v", update.CallbackQuery)
					}
				} else if update.Message != nil {
					if err = this.implementation.OnMessage(this, update.Message); err != nil {
						this.log.Printf("BOT ERROR: update %+v\n%s", update.Message, err.Error())
					} else {
						this.log.Printf("BOT INFO: update %+v", update.Message)
					}
				}
				this.updatesOffset = update.UpdateID + 1
			}
		}
	}
}

func makeApiRequest(url string, httpMethod string, contentType string, body []byte) (*ApiResponse, error) {
	request, err := http.NewRequest(httpMethod, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with http status code %d", response.StatusCode)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var apiResponse ApiResponse
	err = json.Unmarshal(responseBody, &apiResponse)
	if err != nil {
		return nil, err
	}
	if !apiResponse.Ok {
		return nil, fmt.Errorf("request failed with telegram error code %d", apiResponse.ErrorCode)
	}
	return &apiResponse, nil
}

func (this *Bot) prepareApiUrl(apiMethod string, filePath string) string {
	var url string
	if apiMethod != "" {
		url = strings.Replace(apiMethodTemplate, "<METHOD>", apiMethod, 1)
	} else {
		url = strings.Replace(apiFileTemplate, "<PATH>", filePath, 1)
	}
	url = strings.Replace(url, "<TOKEN>", this.token, 1)
	return url
}

func (this *Bot) SendMessage(msg SendMessage) (*Message, error) {
	this.log.Printf("sending %+v", msg)
	retMsg, err := callApiMethod[SendMessage, *Message](this.prepareApiUrl("sendMessage", ""), msg)
	if err != nil {
		this.log.Println(err)
	}
	return retMsg, err
}

func (this *Bot) EditMessageText(msg EditMessageText) (*Message, error) {
	this.log.Printf("editing %+v", msg)
	retMsg, err := callApiMethod[EditMessageText, *Message](this.prepareApiUrl("editMessageText", ""), msg)
	if err != nil {
		this.log.Println(err)
	}
	return retMsg, err
}

func (this *Bot) AnswerCallbackQuery(answer AnswerCallbackQuery) (*bool, error) {
	this.log.Printf("answering %+v", answer)
	retOk, err := callApiMethod[AnswerCallbackQuery, *bool](this.prepareApiUrl("answerCallbackQuery", ""), answer)
	if err != nil {
		this.log.Println(err)
	}
	return retOk, err
}

type allowedIn interface {
	EditMessageText | SendMessage | RequestUpdates | AnswerCallbackQuery
}

type allowedOut interface {
	*Message | []Update | *bool
}

func callApiMethod[I allowedIn, O allowedOut](url string, requestBody I) (O, error) {
	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	apiResponse, err := makeApiRequest(url,
		"POST",
		"application/json",
		requestBodyJson)
	if err != nil {
		return nil, err
	}
	var responseBody O
	err = json.Unmarshal(apiResponse.Result, &responseBody)
	if err != nil {
		return nil, err
	}
	return responseBody, nil
}

func ParseCommand(message Message) (string, error) {
	var command string
	for _, entity := range message.Entities {
		if entity.Type == CommandEntity {
			msgText16 := utf16.Encode([]rune(message.Text))
			substrTo := entity.Offset + entity.Length
			if substrTo > len(msgText16) {
				return "", fmt.Errorf("chat %d message %d bad command: text too short", message.Chat.ID, message.MessageID)
			}
			command16 := msgText16[entity.Offset+1 : substrTo] // omit slash
			command = string(utf16.Decode(command16))
		}
	}
	return command, nil
}
