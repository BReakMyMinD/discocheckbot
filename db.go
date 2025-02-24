package main

import (
	"database/sql"
	"fmt"
	"reflect"

	_ "github.com/lib/pq"
)

type psqlAdapter struct {
	connStr string
}

func newPsqlAdapter(host, user, password, db string, port int) (*psqlAdapter, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, db)
	adapter := psqlAdapter{
		connStr: connStr,
	}
	if db, err := adapter.connect(); err != nil {
		return nil, err
	} else {
		db.Close()
		return &adapter, nil
	}
}

func (this *psqlAdapter) connect() (*sql.DB, error) {
	db, err := sql.Open("postgres", this.connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (this *psqlAdapter) createCheck(chk *check) error {
	conn, err := this.connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	res, err := conn.Query(
		`INSERT INTO checks (
			skill,
			type,
			difficulty,
			description,
			created_at,
			created_by_user,
			created_by_message,
			created_by_chat
			) VALUES (
			$1, $2, $3, $4, 
			now()::timestamp,
			$5, $6, $7
		) RETURNING check_id;`,
		chk.Skill,
		chk.Typ,
		chk.Difficulty,
		chk.Description,
		chk.CreatedByUser,
		chk.CreatedByMessage,
		chk.CreatedByChat)
	if err != nil {
		return err
	}
	defer res.Close()
	if !res.Next() {
		return fmt.Errorf("insert checks not successful, no id returned")
	}
	res.Scan(&chk.Id)
	return nil
}

func (this *psqlAdapter) createAttempt(att *attempt) error {
	conn, err := this.connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	res, err := conn.Query(
		`INSERT INTO attempts (
			check_id,
			result,
			created_at,
			created_by_message,
			created_by_chat
			) VALUES (
			$1, $2,
			now()::timestamp,
			$5, $6
		) RETURNING attempt_id;`,
		att.CheckId,
		att.Result,
		att.CreatedByMessage,
		att.CreatedByChat)
	if err != nil {
		return err
	}
	defer res.Close()
	if !res.Next() {
		return fmt.Errorf("insert attempts not successful, no id returned")
	}
	res.Scan(&att.Id)
	return nil
}

func (this *psqlAdapter) listUserChecks(userId int64, offsetId int64) ([]check, error) {
	conn, err := this.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.Query(
		`WITH check_updates AS (
			SELECT DISTINCT ON(check_id)
				c.check_id,
				coalesce(a.result,0) AS result,
				CASE WHEN a.created_at IS NULL THEN c.created_at ELSE a.created_at END AS updated_at
		  	FROM checks c
			LEFT JOIN attempts a
			ON c.check_id = a.check_id
			WHERE c.created_by_user = $1
			ORDER BY check_id, updated_at DESC
		)
		SELECT 
			c.check_id,
			c.skill,
			c.difficulty,
			c.type,
			c.description,
			u.result
		FROM checks c 
		JOIN check_updates u 
		ON c.check_id = u.check_id 
		WHERE u.updated_at < coalesce(
		(
			SELECT updated_at
			FROM check_updates
			WHERE check_id = $2
		), to_timestamp('9999','YYYY'))
		ORDER BY updated_at DESC
		LIMIT $3;`,
		userId,
		offsetId,
		maxChecksAtListPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]check, 0)
	for rows.Next() {
		chk := check{}
		if err := moveCorresponding(rows, &chk); err != nil {
			return nil, err
		}
		att := attempt{}
		if err := moveCorresponding(rows, &att); err != nil {
			return nil, err
		}
		if att.Result != resDefault {
			chk.Attempts = append(chk.Attempts, att)
		}
		result = append(result, chk)
	}
	return result, nil
}

func (this *psqlAdapter) readCheck(checkId int64) (check, error) {
	conn, err := this.connect()
	if err != nil {
		return check{}, err
	}
	defer conn.Close()
	rows, err := conn.Query(
		`SELECT 
			c.check_id,
			c.skill,
			c.difficulty,
			c.type,
			c.description,
			c.created_at,
			coalesce(a.result,0) AS result,
			a.created_at AS a_created_at
		 FROM checks c
		 LEFT JOIN attempts a
		 ON c.check_id = a.check_id
		 WHERE c.check_id = $1
		 ORDER BY a_created_at;`,
		checkId)
	if err != nil {
		return check{}, err
	}
	defer rows.Close()
	var result check
	for rows.Next() {
		if result.empty() {
			if err := moveCorresponding(rows, &result); err != nil {
				return check{}, err
			}
		}
		att := attempt{}
		if err := moveCorresponding(rows, &att); err != nil {
			return check{}, err
		}
		if att.Result != resDefault {
			result.Attempts = append(result.Attempts, att)
		}
	}
	if result.empty() {
		return result, fmt.Errorf("check %d not found", checkId)
	}
	return result, nil
}

func (this *psqlAdapter) init() error {
	conn, err := this.connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Exec(
		`CREATE TABLE IF NOT EXISTS checks (
			check_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			skill INTEGER,
			type INTEGER,
			difficulty INTEGER,
			description VARCHAR(100),
			created_at TIMESTAMP,
			created_by_user BIGINT,
			created_by_message BIGINT,
			created_by_chat BIGINT
		);
		CREATE TABLE IF NOT EXISTS attempts (
			attempt_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			check_id BIGINT REFERENCES checks (check_id),
			result INTEGER,
			created_at TIMESTAMP,
			created_by_message BIGINT,
			created_by_chat BIGINT
		);`)
	return err
}

func moveCorresponding(row *sql.Rows, struc interface{}) error {
	strucType := reflect.TypeOf(struc).Elem()
	strucVal := reflect.ValueOf(struc).Elem()
	sqlCol, err := row.Columns()
	if err != nil {
		return err
	}
	resVal := make([]interface{}, len(sqlCol))
	resPtr := make([]interface{}, len(sqlCol))
	for i := range sqlCol {
		resPtr[i] = &resVal[i]
	}
	if err = row.Scan(resPtr...); err != nil {
		return err
	}
	var fldVal reflect.Value
	for i, col := range sqlCol {
		match := false
		for j := 0; j < strucVal.NumField(); j++ {
			fldTag := strucType.Field(j).Tag.Get("sql")
			if fldTag == col {
				fldVal = strucVal.Field(j)
				match = true
				break
			}
		}
		if match && reflect.ValueOf(resVal[i]) != reflect.ValueOf(nil) {
			if fldVal.Addr().Elem().Type().AssignableTo(reflect.TypeOf(resVal[i])) {
				fldVal.Addr().Elem().Set(reflect.ValueOf(resVal[i]))
			} else if reflect.TypeOf(resVal[i]).ConvertibleTo(fldVal.Addr().Elem().Type()) {
				fldVal.Addr().Elem().Set(reflect.ValueOf(resVal[i]).Convert(fldVal.Addr().Elem().Type()))
			}
		}
	}
	return nil
}
