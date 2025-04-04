package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/donghquinn/hls_converter/configs"
	_ "github.com/lib/pq"
)

type DataBaseConnector struct {
	*sql.DB
}

// DB 연결 인스턴스
func InitPostgresConnection() (*DataBaseConnector, error) {
	dbConfig := configs.DatabaseConfiguration

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbConfig.DbUser,
		dbConfig.DbPasswd,
		dbConfig.DbHost,
		dbConfig.DbPort,
		dbConfig.DbName,
	)

	// driver := sql.Open("mysql", )

	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		log.Printf("[DATABASE] Start Database Connection Error: %v", err)

		return nil, err
	}

	db.SetConnMaxLifetime(time.Second * 60)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)

	connect := &DataBaseConnector{db}

	return connect, nil
}

// 테이블 생성
func (connect *DataBaseConnector) CheckPostgresConnection() error {
	log.Printf("Waiting for Database Connection,,,")
	time.Sleep(time.Second * 10)

	defer connect.Close()

	pingErr := connect.Ping()

	if pingErr != nil {
		log.Printf("[DATABASE] Database Ping Error: %v", pingErr)
		return pingErr
	}

	return nil
}

func (connect *DataBaseConnector) CreateTable(queryList []string) error {
	ctx := context.Background()

	tx, txErr := connect.Begin()

	if txErr != nil {
		log.Printf("[DATABASE] Begin Transaction Error: %v", txErr)
		return txErr
	}

	defer tx.Rollback()

	for _, queryString := range queryList {
		_, execErr := tx.ExecContext(ctx, queryString)

		if execErr != nil {
			tx.Rollback()
			log.Printf("[DATABASE] Create Table Querystring Transaction Exec Error: %v", execErr)
			return execErr
		}
	}

	commitErr := tx.Commit()

	if commitErr != nil {
		log.Printf("[DATABASE] Commit Transaction Error: %v", commitErr)
		return commitErr
	}

	return nil
}

// func (connect *DataBaseConnector) QueryRowMultiple(queryString string, args ...string) (*sql.Rows, *sql.Tx, error) {
// 	// 1) 트랜잭션 시작
// 	tx, err := connect.DB.Begin()
// 	if err != nil {
// 		log.Printf("[QUERY TX] Begin transaction failed: %v\n", err)
// 		return nil, nil, err
// 	}

// 	// 2) 쿼리 인자 변환
// 	var arguments []interface{}
// 	for _, arg := range args {
// 		arguments = append(arguments, arg)
// 	}

// 	// 3) 트랜잭션에서 SELECT 실행
// 	rows, err := tx.Query(queryString, arguments...)
// 	if err != nil {
// 		log.Printf("[QUERY TX] Query error: %v\n", err)
// 		// 쿼리 에러 발생 시 즉시 롤백
// 		_ = tx.Rollback()
// 		return nil, nil, err
// 	}

// 	// 4) rows와 트랜잭션을 함께 반환
// 	return rows, tx, nil
// }

// 쿼리
func (connect *DataBaseConnector) QueryRows(queryString string, args ...string) (*sql.Rows, error) {
	var arguments []interface{}

	for _, arg := range args {
		arguments = append(arguments, arg)
	}

	result, err := connect.Query(queryString, arguments...)

	if err != nil {
		log.Printf("[QUERY] Query Error: %v\n", err)

		return nil, err
	}

	defer connect.Close()

	return result, nil
}

func (connect *DataBaseConnector) QueryBuilderRows(queryString string, args []interface{}) (*sql.Rows, error) {
	result, err := connect.Query(queryString, args...)

	if err != nil {
		log.Printf("[QUERY] Query Error: %v\n", err)

		return nil, err
	}

	defer connect.Close()

	return result, nil
}

// 쿼리
func (connect *DataBaseConnector) QueryOne(queryString string, args ...string) (*sql.Row, error) {
	var arguments []interface{}

	for _, arg := range args {
		arguments = append(arguments, arg)
	}

	result := connect.QueryRow(queryString, arguments...)

	if result.Err() != nil {
		log.Printf("[QUERY] Query Error: %v\n", result.Err())

		return nil, result.Err()
	}

	defer connect.Close()

	return result, nil
}

func (connect *DataBaseConnector) QueryBuilderOneRow(queryString string, args []interface{}) (*sql.Row, error) {
	result := connect.QueryRow(queryString, args...)

	if result.Err() != nil {
		log.Printf("[QUERY] Query Error: %v\n", result.Err())

		return nil, result.Err()
	}

	defer connect.Close()

	return result, nil
}

// 인서트 쿼리
func (connect *DataBaseConnector) InsertQuery(queryString string, returns []interface{}, args ...string) error {
	var arguments []interface{}

	for _, arg := range args {
		arguments = append(arguments, arg)
	}

	queryResult := connect.QueryRow(queryString, arguments...)

	defer connect.Close()

	if returns != nil {
		// Insert ID
		if scanErr := queryResult.Scan(returns...); scanErr != nil {
			log.Printf("[INSERT] Get Insert Result Scan Error: %v", scanErr)
			return scanErr
		}
	}

	return nil
}

// 인서트 쿼리
func (connect *DataBaseConnector) UpdateQuery(queryString string, returns []interface{}, args ...string) error {
	var arguments []interface{}

	for _, arg := range args {
		arguments = append(arguments, arg)
	}

	_, queryErr := connect.Exec(queryString, arguments...)

	defer connect.Close()

	if queryErr != nil {
		// Insert ID

		log.Printf("[UPDATE] Get Update Error: %v", queryErr)
		return queryErr

	}

	return nil
}

func (connect *DataBaseConnector) InsertMultiple(queryList []string) error {
	ctx := context.Background()

	tx, txErr := connect.Begin()

	if txErr != nil {
		log.Printf("[INSERT_MULTIPLE] Begin Transaction Error: %v", txErr)
		return txErr
	}

	defer tx.Rollback()

	for _, queryString := range queryList {
		_, execErr := tx.ExecContext(ctx, queryString)

		if execErr != nil {
			tx.Rollback()
			log.Printf("[INSERT_MULTIPLE] Insert Querystring Transaction Exec Error: %v", execErr)
			return execErr
		}
	}

	commitErr := tx.Commit()

	if commitErr != nil {
		log.Printf("[INSERT_MULTIPLE] Commit Transaction Error: %v", commitErr)
		return commitErr
	}

	return nil
}

func (connect *DataBaseConnector) UpdateMultiple(queryList []string) error {
	ctx := context.Background()

	tx, txErr := connect.Begin()

	if txErr != nil {
		log.Printf("[UPDATE_MULTIPLE] Begin Transaction Error: %v", txErr)
		return txErr
	}

	defer tx.Rollback()

	for _, queryString := range queryList {
		_, execErr := tx.ExecContext(ctx, queryString)

		if execErr != nil {
			tx.Rollback()
			log.Printf("[UPDATE_MULTIPLE] Update Querystring Transaction Exec Error: %v", execErr)
			return execErr
		}
	}

	commitErr := tx.Commit()

	if commitErr != nil {
		log.Printf("[UPDATE_MULTIPLE] Commit Transaction Error: %v", commitErr)
		return commitErr
	}

	return nil
}
