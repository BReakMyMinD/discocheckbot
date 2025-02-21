package main

import (
	"discocheckbot/api"
	"discocheckbot/config"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type dbAdapter interface {
	createCheck(chk *check) error
	init() error
	listUserChecks(userId int64, offset int) ([]check, error)
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
	switch len(callbackParams) {
	case 3:
		switch callbackParams[0] {
		case addWhite:
			fallthrough
		case addRed:
			bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
			bot.EditMessageText(getSkillDifEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, cbq.Data))
			return nil
		case seeTop:
			chk, err := this.displayCheck(cbq, callbackParams)
			if err != nil {
				bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
			} else {
				bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
				bot.EditMessageText(api.EditMessageText(getSingleCheckEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, cbq.Data, chk)))
			}
			return err
		default:
			err := fmt.Errorf("invalid callback data")
			bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
			return err
		}
	case 4:
		chk, err := this.askNewCheckName(cbq, callbackParams)
		if err != nil {
			bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
		} else {
			bot.AnswerCallbackQuery(getCbqAnswer(cbq.ID, ""))
			bot.EditMessageText(getSkillTxtEditMessage(cbq.Message.Chat.ID, cbq.Message.MessageID, chk))
		}
		return err
	default:
		err := fmt.Errorf("invalid callback data")
		bot.AnswerCallbackQuery(getErrorCbqAnswer(cbq.ID, err))
		return err
	}
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
	return check{}, nil
}

// func (this *DiscoCheckBot) showListCheckInitialPage(userId int64) error {
// 	checks, err := this.db.listUserChecks(userId, 0)
// 	if err != nil {
// 		return err
// 	}
// 	for check, i := range checks {

// 	}
// }

func getSkillMessage(cmd string, chatId int64, chkColor int) api.SendMessage {
	smsg := api.SendMessage{
		ChatID: chatId,
		Text:   "select skill:",
		ReplyMarkup: &api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{
				{{Text: skillNames[intLogic], CallbackData: makeClbk(cmd, chkColor, intLogic)},
					{Text: skillNames[intEncyclopedia], CallbackData: makeClbk(cmd, chkColor, intEncyclopedia)}},
				{{Text: skillNames[intRhetoric], CallbackData: makeClbk(cmd, chkColor, intRhetoric)},
					{Text: skillNames[intDrama], CallbackData: makeClbk(cmd, chkColor, intDrama)}},
				{{Text: skillNames[intConcept], CallbackData: makeClbk(cmd, chkColor, intConcept)},
					{Text: skillNames[intVisual], CallbackData: makeClbk(cmd, chkColor, intEncyclopedia)}},
				{{Text: skillNames[psyVolition], CallbackData: makeClbk(cmd, chkColor, psyVolition)},
					{Text: skillNames[psyInland], CallbackData: makeClbk(cmd, chkColor, psyInland)}},
				{{Text: skillNames[psyEmpathy], CallbackData: makeClbk(cmd, chkColor, psyEmpathy)},
					{Text: skillNames[psyAuthority], CallbackData: makeClbk(cmd, chkColor, psyAuthority)}},
				{{Text: skillNames[psyEsprit], CallbackData: makeClbk(cmd, chkColor, psyEsprit)},
					{Text: skillNames[psySuggestion], CallbackData: makeClbk(cmd, chkColor, psySuggestion)}},
				{{Text: skillNames[phyEndurance], CallbackData: makeClbk(cmd, chkColor, phyEndurance)},
					{Text: skillNames[phyPain], CallbackData: makeClbk(cmd, chkColor, phyPain)}},
				{{Text: skillNames[phyInstrument], CallbackData: makeClbk(cmd, chkColor, phyInstrument)},
					{Text: skillNames[phyElectrochem], CallbackData: makeClbk(cmd, chkColor, phyElectrochem)}},
				{{Text: skillNames[phyShivers], CallbackData: makeClbk(cmd, chkColor, phyShivers)},
					{Text: skillNames[phyHalflight], CallbackData: makeClbk(cmd, chkColor, phyHalflight)}},
				{{Text: skillNames[motCoordintation], CallbackData: makeClbk(cmd, chkColor, motCoordintation)},
					{Text: skillNames[motPerception], CallbackData: makeClbk(cmd, chkColor, motPerception)}},
				{{Text: skillNames[motReaction], CallbackData: makeClbk(cmd, chkColor, motReaction)},
					{Text: skillNames[motSavoir], CallbackData: makeClbk(cmd, chkColor, motSavoir)}},
				{{Text: skillNames[motInterfacing], CallbackData: makeClbk(cmd, chkColor, motInterfacing)},
					{Text: skillNames[motComposure], CallbackData: makeClbk(cmd, chkColor, motComposure)}},
			},
		},
	}
	return smsg
}

func getSkillDifEditMessage(chatId int64, msgId int, clbk string) api.EditMessageText {
	emsg := api.EditMessageText{
		ChatID:    chatId,
		MessageID: msgId,
		Text:      "select check difficulty:",
		ReplyMarkup: &api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{
				{{Text: difficultyNames[difTrivial], CallbackData: makeClbk(clbk, difTrivial)}},
				{{Text: difficultyNames[difEasy], CallbackData: makeClbk(clbk, difEasy)}},
				{{Text: difficultyNames[difMedium], CallbackData: makeClbk(clbk, difMedium)}},
				{{Text: difficultyNames[difChallenging], CallbackData: makeClbk(clbk, difChallenging)}},
				{{Text: difficultyNames[difFormidable], CallbackData: makeClbk(clbk, difFormidable)}},
				{{Text: difficultyNames[difLegendary], CallbackData: makeClbk(clbk, difLegendary)}},
				{{Text: difficultyNames[difHeroic], CallbackData: makeClbk(clbk, difHeroic)}},
				{{Text: difficultyNames[difGodly], CallbackData: makeClbk(clbk, difGodly)}},
				{{Text: difficultyNames[difImpossible], CallbackData: makeClbk(clbk, difImpossible)}},
			},
		},
	}
	return emsg
}

func getListCheckMessage(cmd string, chatId int64, list []check) api.SendMessage {

	smsg := api.SendMessage{}
	return smsg
}

func getSingleCheckEditMessage(chatId int64, msgId int, clbk string, chk check) api.EditMessageText {

	emsg := api.EditMessageText{}
	return emsg
}

func getSingleCheckMessage(chatId int64, chk check) api.SendMessage {
	smsg := api.SendMessage{
		ChatID: chatId,
		Text: typeNames[chk.Typ] + ":\n" + skillNames[chk.Skill] + "/" + difficultyNames[chk.Difficulty] +
			"\n" + chk.Description + "\n\nCreated at: " + chk.CreatedAt.Format(time.DateTime),
	}
	return smsg
}

func getSkillTxtEditMessage(chatId int64, msgId int, chk check) api.EditMessageText {
	emsg := api.EditMessageText{
		ChatID:    chatId,
		MessageID: msgId,
		Text:      "enter descrption of the check:\n" + skillNames[chk.Skill] + "/" + difficultyNames[chk.Difficulty],
	}
	return emsg
}

func getErrorMessage(chatId int64, err error) api.SendMessage {
	smsg := api.SendMessage{
		ChatID: chatId,
		Text:   fmt.Sprintf("update was not handled due to:\n%s", err.Error()),
	}
	return smsg
}

func getCbqAnswer(cbqId string, text string) api.AnswerCallbackQuery {
	answer := api.AnswerCallbackQuery{
		CallbackQueryId: cbqId,
		Text:            text,
	}
	return answer
}

func getErrorCbqAnswer(cbqId string, err error) api.AnswerCallbackQuery {
	answer := api.AnswerCallbackQuery{
		CallbackQueryId: cbqId,
		Text:            err.Error(),
		ShowAlert:       true,
	}
	return answer
}

func makeClbk(start string, params ...int) string {
	var sb strings.Builder

	for i := 0; i < len(params); i++ {
		if i > 0 || start == "" {
			sb.WriteString("/")
		} else {
			sb.WriteString(start)
			sb.WriteString("/")
		}
		sb.WriteString(strconv.Itoa(params[i]))
	}
	return sb.String()
}
