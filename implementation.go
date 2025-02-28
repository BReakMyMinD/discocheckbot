package main

import (
	"discocheckbot/api"
	"discocheckbot/config"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type dbAdapter interface {
	createCheck(chk *check) error
	createAttempt(att *attempt) error
	init() error
	listUserChecks(userId int64, offsetId int64, desc bool) ([]check, error)
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
		return this.handleNewCheckDescr(bot, msg)
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
			return this.displayListChecks(bot, msg)
		default:
			err = fmt.Errorf("unsupported command %s", command)
			bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
			return err
		}
		return nil
	}
}

func (this *DiscoCheckBot) OnCallbackQuery(bot *api.Bot, cbq *api.CallbackQuery) error {
	var ok bool
	var err error
	callbackParams := strings.Split(cbq.Data, "/")
	if len(callbackParams) > 2 {
		switch callbackParams[0] {
		case addWhite:
			fallthrough
		case addRed:
			if ok, err = this.handleNewCheckProperty(bot, cbq, callbackParams); ok {
				return err
			}
		case seeTop:
			if oper, err := strconv.Atoi(callbackParams[1]); err == nil {
				switch oper {
				case listCheckDetail:
					if ok, err = this.displayCheck(bot, cbq, callbackParams); ok {
						return err
					}
				case listCheckForward:
					fallthrough
				case listCheckBackward:
					if ok, err = this.refreshListChecks(bot, cbq, callbackParams); ok {
						return err
					}
				case listCheckAction:
					if ok, err = this.handleCheckAction(bot, cbq, callbackParams); ok {
						return err
					}
				}
			}
		}
	}
	err = fmt.Errorf("invalid callback data %s because of %v", cbq.Data, err)
	bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
	return err
}

func (this *DiscoCheckBot) handleNewCheckProperty(bot *api.Bot, cbq *api.CallbackQuery, clbkPar []string) (bool, error) {
	if len(clbkPar) == 3 {
		bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
		bot.EditMessageText(getSkillDifEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, cbq.Data))
		return true, nil
	} else if len(clbkPar) == 4 {
		var dffclt int
		var typ int
		var skill int
		var err error
		var chk check
		if typ, err = strconv.Atoi(clbkPar[1]); err != nil {
			return false, err
		}
		if skill, err = strconv.Atoi(clbkPar[2]); err != nil {
			return false, err
		}
		if dffclt, err = strconv.Atoi(clbkPar[3]); err != nil {
			return false, err
		}
		chk = check{
			Typ:        typ,
			Skill:      skill,
			Difficulty: dffclt,
		}
		this.checkBuffer[cbq.Sender.ID] = chk
		bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
		bot.EditMessageText(getSkillTxtEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, chk))
		return true, nil
	} else {
		return false, errors.New("invalid number of params")
	}
}

func (this *DiscoCheckBot) handleNewCheckDescr(bot *api.Bot, msg *api.Message) error {
	var err error
	chk, ok := this.checkBuffer[msg.Sender.ID]
	if ok && msg.Text != "" {
		delete(this.checkBuffer, msg.Sender.ID)
		chk.Description = msg.Text
		chk.CreatedByUser = msg.Sender.ID
		chk.CreatedByMessage = msg.MessageID
		chk.CreatedByChat = msg.Chat.ID
		if err = chk.validate(); err != nil {
			bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
			return err
		}
		if err = this.db.createCheck(&chk); err != nil {
			bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
			return err
		}
		if chk, err = this.db.readCheck(chk.Id); err != nil {
			bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
		} else {
			bot.SendMessage(getSingleCheckMessage(msg.Chat.ID, chk))
		}
	}
	return err
}

func (this *DiscoCheckBot) displayCheck(bot *api.Bot, cbq *api.CallbackQuery, clbkPar []string) (bool, error) {
	var chk check
	var err error
	if chk.Id, err = strconv.ParseInt(clbkPar[2], 10, 64); err != nil {
		return false, err
	}
	chk, err = this.db.readCheck(chk.Id)
	if err != nil {
		bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
	} else {
		bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
		bot.EditMessageText(getSingleCheckEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, chk))
	}
	return true, err
}

func (this *DiscoCheckBot) displayListChecks(bot *api.Bot, msg *api.Message) error {
	list, err := this.db.listUserChecks(msg.Sender.ID, 0, false)
	if err != nil {
		bot.SendMessage(getErrorMessage(msg.Chat.ID, err))
	} else {
		bot.SendMessage(getListCheckMessage(seeTop, msg.Chat.ID, list))
	}
	return err
}

func (this *DiscoCheckBot) refreshListChecks(bot *api.Bot, cbq *api.CallbackQuery, clbkPar []string) (bool, error) {
	var nextChkId int64
	var err error
	list := make([]check, 0)
	oper, _ := strconv.Atoi(clbkPar[1])
	if nextChkId, err = strconv.ParseInt(clbkPar[2], 10, 64); err != nil {
		return false, err
	}
	list, err = this.db.listUserChecks(cbq.Sender.ID, nextChkId, oper == listCheckBackward)
	if err != nil {
		bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
	} else {
		bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
		if len(list) > 0 {
			bot.EditMessageText(getListCheckEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, list))
		}
	}
	return true, err
}

func (this *DiscoCheckBot) handleCheckAction(bot *api.Bot, cbq *api.CallbackQuery, clbkPar []string) (bool, error) {
	var att attempt
	var err error
	if att.CheckId, err = strconv.ParseInt(clbkPar[2], 10, 64); err != nil {
		return false, err
	}
	if att.Result, err = strconv.Atoi(clbkPar[3]); err != nil {
		return false, err
	}
	att.CreatedByMessage = cbq.Message.MessageID
	att.CreatedByChat = cbq.Message.Chat.ID
	if err = att.validate(); err != nil {
		bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
		return true, err
	}
	err = this.db.createAttempt(&att)
	if err != nil {
		bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
	} else {
		list, err := this.db.listUserChecks(cbq.Sender.ID, 0, false)
		if err != nil {
			bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
		} else {
			bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
			if len(list) > 0 {
				bot.EditMessageText(getListCheckEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, list))
			}
		}
	}
	return true, err
}
