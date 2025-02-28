package api

import (
	"bytes"
	"discocheckbot/config"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf16"
)

// telegram bot API constants
const (
	CommandEntity     string = "bot_command"
	CrossedEntity     string = "strikethrough"
	BoldEntity        string = "bold"
	apiMethodTemplate string = "https://api.telegram.org/bot<TOKEN>/<METHOD>"
	apiFileTemplate   string = "https://api.telegram.org/file/bot<TOKEN>/<PATH>"
)

type BotImplementation interface {
	OnMessage(bot *Bot, msg *Message) error
	OnCallbackQuery(bot *Bot, cbq *CallbackQuery) error
}

type Bot struct {
	token           string
	updatesOffset   int
	updatesLimit    int
	reqUpdatesRetry int //seconds
	httpTimeout     int //seconds
	log             *log.Logger
	implementation  BotImplementation
}

func NewBot(cfg *config.ConfigReader, log *log.Logger, impl BotImplementation) (*Bot, error) {
	var token string
	var httpTimeout, reqUpdatesRetry, updatesLimit float64 //json number interprets as float64!
	var err error
	//reading config
	if err = cfg.GetParameter("bot_token", &token); err != nil {
		return nil, err
	}
	if err = cfg.GetParameter("long_polling_timeout", &httpTimeout); err != nil {
		return nil, err
	}
	if err = cfg.GetParameter("request_updates_retry", &reqUpdatesRetry); err != nil {
		return nil, err
	}
	if err = cfg.GetParameter("updates_limit", &updatesLimit); err != nil {
		return nil, err
	}
	//initializing bot
	bot := Bot{
		token,
		0,
		int(updatesLimit),
		int(reqUpdatesRetry),
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
		return nil, err
	}
	return &bot, nil
}

func (this *Bot) ListenForUpdates() {
	for {
		requestBody := RequestUpdates{
			this.updatesOffset,
			this.updatesLimit,
			this.httpTimeout,
			[]string{"message", "callback_query"},
		}
		log.Printf("INFO: requesting updates from %d\n", this.updatesOffset)
		updates, err := callApiMethod[RequestUpdates, []Update](this.prepareApiUrl("getUpdates", ""), requestBody)
		if err != nil {
			log.Printf("ERROR: %v, retrying in %d seconds\n", err, this.reqUpdatesRetry)
			time.Sleep(time.Second * time.Duration(this.reqUpdatesRetry))
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= this.updatesOffset {
				if update.CallbackQuery != nil {
					if err = this.implementation.OnCallbackQuery(this, update.CallbackQuery); err != nil {
						this.log.Printf("BOT ERROR: %v: callback query %s\nfrom %+v\nchat %d\nmessage %d\nwith %s\n",
							err,
							update.CallbackQuery.ID,
							update.CallbackQuery.Sender,
							update.CallbackQuery.Message.Chat.ID,
							update.CallbackQuery.Message.MessageID,
							update.CallbackQuery.Data)
					} else {
						this.log.Printf("BOT INFO: callback query %s\nfrom %+v\nchat %d\nmessage %d\nwith %s\n",
							update.CallbackQuery.ID,
							update.CallbackQuery.Sender,
							update.CallbackQuery.Message.Chat.ID,
							update.CallbackQuery.Message.MessageID,
							update.CallbackQuery.Data)
					}
				} else if update.Message != nil {
					if err = this.implementation.OnMessage(this, update.Message); err != nil {
						this.log.Printf("BOT ERROR: %v: message %d\nfrom %+v\nchat %d\nwith %s\n",
							err,
							update.Message.MessageID,
							update.Message.Sender,
							update.Message.Chat.ID,
							update.Message.Text)
					} else {
						this.log.Printf("BOT INFO: message %d\nfrom %+v\nchat %d\nwith %s\n",
							update.Message.MessageID,
							update.Message.Sender,
							update.Message.Chat.ID,
							update.Message.Text)
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
		return nil, fmt.Errorf("request for %s failed with http status code %d", url, response.StatusCode)
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
		return nil, fmt.Errorf("request for %s failed with telegram error code %d", url, apiResponse.ErrorCode)
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
	retMsg, err := callApiMethod[SendMessage, *Message](this.prepareApiUrl("sendMessage", ""), msg)
	if err != nil {
		this.log.Printf("ERROR: %v: send message chat %d\n",
			err,
			msg.ChatID)
	} else {
		this.log.Printf("INFO: send message %d\nchat %d\n",
			retMsg.MessageID,
			retMsg.Chat.ID)
	}
	return retMsg, err
}

func (this *Bot) EditMessageText(msg EditMessageText) (*Message, error) {
	retMsg, err := callApiMethod[EditMessageText, *Message](this.prepareApiUrl("editMessageText", ""), msg)
	if err != nil {
		this.log.Printf("ERROR: %v: edit message %d\nchat %d\n",
			err,
			msg.MessageID,
			msg.ChatID)
	} else {
		this.log.Printf("INFO: edit message %d\nfrom %+v\nchat %d\n",
			retMsg.MessageID,
			retMsg.Sender,
			retMsg.Chat.ID)
	}
	return retMsg, err
}

func (this *Bot) AnswerCallbackQuery(answer AnswerCallbackQuery) (*bool, error) {
	retOk, err := callApiMethod[AnswerCallbackQuery, *bool](this.prepareApiUrl("answerCallbackQuery", ""), answer)
	if err != nil {
		this.log.Printf("ERROR %v: answer callback query %s\n",
			err,
			answer.CallbackQueryId)
	} else {
		this.log.Printf("INFO: answer callback query %s\n",
			answer.CallbackQueryId)
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
				return "", errors.New("bad command: text too short")
			}
			command16 := msgText16[entity.Offset+1 : substrTo] // omit slash
			command = string(utf16.Decode(command16))
		}
	}
	return command, nil
}
