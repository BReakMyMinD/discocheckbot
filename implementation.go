package main

import (
	"discocheckbot/api"
	"discocheckbot/config"
	"fmt"
	"strconv"
	"strings"
)

type dbAdapter interface {
	createCheck(chk *check) error
	init() error
	listUserChecks(userId int64, offsetId int64) ([]check, error)
	readCheck(checkId int64) (check, error)
}

type DiscoCheckBot struct {
	checkBuffer map[int64]check
	db          dbAdapter
}

func NewDiscoCheckBot(cfg *config.ConfigReader) (*DiscoCheckBot, error) {
	var dbHost, dbName, dbUser, dbPassword string
	var dbPort float64
	var err error
	if err = cfg.GetParameter("db_host", &dbHost); err != nil {
		return nil, err
	}
	if err = cfg.GetParameter("db_port", &dbPort); err != nil {
		return nil, err
	}
	if err = cfg.GetParameter("db_user", &dbUser); err != nil {
		return nil, err
	}
	if err = cfg.GetParameter("db_password", &dbPassword); err != nil {
		return nil, err
	}
	if err = cfg.GetParameter("db_name", &dbName); err != nil {
		return nil, err
	}
	db, err := newPsqlAdapter(dbHost, dbUser, dbPassword, dbName, int(dbPort))
	if err != nil {
		return nil, err
	}
	if err = db.init(); err != nil {
		return nil, err
	}
	dcb := DiscoCheckBot{
		make(map[int64]check),
		db,
	}
	return &dcb, nil
}

func (this *DiscoCheckBot) OnMessage(bot *api.Bot, msg *api.Message) error {
	command, err := api.ParseCommand(*msg)
	if command == "" {
		chk, err := this.receiveNewCheckName(msg)
		if err != nil {
			bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
		} else if !chk.empty() {
			if chk, err = this.db.readCheck(chk.Id); err != nil {
				bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
			}
			bot.SendMessage(getSingleCheckMessage(msg.Chat.ID, chk))
		}
		return err
	} else {
		delete(this.checkBuffer, msg.Sender.ID)
		switch command {
		case start:
			bot.SendMessage(getStartMessage(msg.Chat.ID))
		case addWhite:
			bot.SendMessage(getSkillMessage(command, msg.Chat.ID, typRetriable))
		case addRed:
			bot.SendMessage(getSkillMessage(command, msg.Chat.ID, typNonRetriable))
		case seeTop:
			list, err := this.db.listUserChecks(msg.Sender.ID, 0)
			if err != nil {
				bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
			} else {
				bot.SendMessage(getListCheckMessage(command, msg.Chat.ID, list))
			}
			return err
		default:
			err = fmt.Errorf("unsupported command %s", command)
			bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
			return err
		}
		return nil
	}
}

func (this *DiscoCheckBot) OnCallbackQuery(bot *api.Bot, cbq *api.CallbackQuery) error {
	delete(this.checkBuffer, cbq.Sender.ID)
	callbackParams := strings.Split(cbq.Data, "/")
	// switch len(callbackParams) {
	// case 3:
	if len(callbackParams) > 2 {
		switch callbackParams[0] {
		case addWhite:
			fallthrough
		case addRed:
			if len(callbackParams) == 3 {
				bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
				bot.EditMessageText(getSkillDifEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, cbq.Data))
				return nil
			} else if len(callbackParams) == 4 {
				chk, err := this.askNewCheckName(cbq, callbackParams)
				if err != nil {
					bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
				} else {
					bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
					bot.EditMessageText(getSkillTxtEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, chk))
				}
				return err
			}
		case seeTop:
			if oper, err := strconv.Atoi(callbackParams[1]); err == nil {
				switch oper {
				case listCheckDetail:
					chk, err := this.displayCheck(cbq, callbackParams)
					if err != nil {
						bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
					} else {
						bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
						bot.EditMessageText(getSingleCheckEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, cbq.Data, chk))
					}
					return err
				//case listCheckPrevious:
				//	fallthrough
				case listCheckNext:
					list, err := this.updateListPage(cbq, callbackParams)
					if err != nil {
						bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
					} else {
						bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
						bot.EditMessageText(getListCheckEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, list, callbackParams))
					}
					return err
				}
			}
		}
	}
	err := fmt.Errorf("invalid callback data")
	bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
	return err
}

func (this *DiscoCheckBot) askNewCheckName(cbq *api.CallbackQuery, clbkPar []string) (check, error) {
	var dffclt int
	var typ int
	var skill int
	var err error
	var chk check
	if typ, err = strconv.Atoi(clbkPar[1]); err != nil {
		return chk, err
	}
	if skill, err = strconv.Atoi(clbkPar[2]); err != nil {
		return chk, err
	}
	if dffclt, err = strconv.Atoi(clbkPar[3]); err != nil {
		return chk, err
	}
	chk = check{
		Typ:        typ,
		Skill:      skill,
		Difficulty: dffclt,
	}
	this.checkBuffer[cbq.Sender.ID] = chk
	return chk, nil
}

func (this *DiscoCheckBot) receiveNewCheckName(msg *api.Message) (check, error) {
	var err error
	chk, ok := this.checkBuffer[msg.Sender.ID]
	if ok && msg.Text != "" {
		delete(this.checkBuffer, msg.Sender.ID)
		chk.Description = msg.Text
		chk.CreatedByUser = msg.Sender.ID
		chk.CreatedByMessage = msg.MessageID
		chk.CreatedByChat = msg.Chat.ID
		if err = chk.validate(); err != nil {
			return chk, err
		}
		if err = this.db.createCheck(&chk); err != nil {
			return chk, err
		}
	}
	return chk, nil
}

func (this *DiscoCheckBot) displayCheck(cbq *api.CallbackQuery, clbkPar []string) (check, error) {
	var chk check
	var err error
	if chk.Id, err = strconv.ParseInt(clbkPar[2], 10, 64); err != nil {
		return chk, err
	}
	return this.db.readCheck(chk.Id)
}

func (this *DiscoCheckBot) updateListPage(cbq *api.CallbackQuery, clbkPar []string) ([]check, error) {
	var nextChkId int64
	list := make([]check, 0)
	var err error
	//test

	if nextChkId, err = strconv.ParseInt(clbkPar[2], 10, 64); err != nil {
		return list, err
	}

	return this.db.listUserChecks(cbq.Sender.ID, nextChkId)
}

// func (this *DiscoCheckBot) showListCheckInitialPage(userId int64) error {
// 	checks, err := this.db.listUserChecks(userId, 0)
// 	if err != nil {
// 		return err
// 	}
// 	for check, i := range checks {

// 	}
// }
