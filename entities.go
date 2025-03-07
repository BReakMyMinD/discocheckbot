package main

import (
	"errors"
	"fmt"
	"time"
)

type check struct {
	// core logic attributes
	Id          int64  `sql:"check_id"`
	Skill       int    `sql:"skill"`
	Difficulty  int    `sql:"difficulty"`
	Typ         int    `sql:"type"`
	Description string `sql:"description"`
	Attempts    []attempt
	// metadata attributes
	CreatedByUser    int64     `sql:"created_by_user"`
	CreatedByChat    int64     `sql:"created_by_chat"`
	CreatedByMessage int       `sql:"created_by_message"`
	CreatedAt        time.Time `sql:"created_at"`
}

func (this check) empty() bool {
	return this.Skill+this.Difficulty+this.Typ == 0
}

func (this check) closed() bool {
	if i := len(this.Attempts); i > 0 {
		switch this.Attempts[len(this.Attempts)-1].Result {
		case resCanceled:
			fallthrough
		case resSuccess:
			return true
		case resFailure:
			return this.Typ != typRetriable
		}
	}
	return false
}

func (this check) validate() error {
	if this.Typ < typNonRetriable || this.Typ > typRetriable {
		return fmt.Errorf("invalid type %d", this.Typ)
	}
	if this.Skill < intLogic || this.Skill > motComposure {
		return fmt.Errorf("invalid skill %d", this.Skill)
	}
	if this.Difficulty < difTrivial || this.Difficulty > difImpossible {
		return fmt.Errorf("invalid difficulty %d", this.Difficulty)
	}
	if this.CreatedByUser == 0 ||
		this.CreatedByChat == 0 ||
		this.CreatedByMessage == 0 {
		return errors.New("incomplete metadata")
	}
	return nil
}

type attempt struct {
	// core logic attributes
	Id      int64 `sql:"attempt_id"`
	CheckId int64 `sql:"check_id"`
	Result  int   `sql:"result"`
	//metadata attributes
	CreatedByChat    int64     `sql:"created_by_chat"`
	CreatedByMessage int       `sql:"created_by_message"`
	CreatedAt        time.Time `sql:"a_created_at"`
}

func (this attempt) validate() error {
	if this.Result < resCanceled || this.Result > resSuccess {
		return fmt.Errorf("invalid result %d", this.Result)
	}
	if this.CreatedByChat == 0 ||
		this.CreatedByMessage == 0 {
		return errors.New("incomplete metadata")
	}
	return nil
}
