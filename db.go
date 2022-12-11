package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const DefaultLimit = 50

const (
	Select      string = "select"
	Insert             = "insert"
	Update             = "update"
	Delete             = "delete"
	OrderByAsc         = "asc"
	OrderByDesc        = "desc"
)

type SlowQueryLog struct {
	Query         string `json:"query"`
	TotalExecTime string `json:"total_exec_time"`
}

type Repo interface {
	Close()
	Get(params *QueryParams) ([]SlowQueryLog, error)
	Demo() error
}

type pgxIface interface {
	Begin(context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Ping(context.Context) error
	Close()
}

type PostgresDB struct {
	conn pgxIface
}

func NewPostgresRepo(conn pgxIface) Repo {
	return &PostgresDB{
		conn: conn,
	}
}

func (db *PostgresDB) Close() {
	db.conn.Close()
}

type QueryParams struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	QueryType string `query:"query_type"`
	OrderBy   string `query:"order_by"`
}

func extractQueryParams(params *QueryParams) (int, int, string, string, error) {
	limit := params.PageSize
	if limit == 0 {
		limit = DefaultLimit
	}

	offset := 0
	if params.Page > 0 {
		offset = params.Page * limit
	}

	var err error
	var queryType string
	switch strings.ToLower(params.QueryType) {
	case "":
		queryType = ""
	case Select:
		queryType = Select
		break
	case Insert:
		queryType = Insert
		break
	case Update:
		queryType = Update
		break
	case Delete:
		queryType = Delete
		break
	default:
		err = fmt.Errorf("invalid query_type: %s", params.QueryType)
	}

	var orderBy string
	switch strings.ToLower(params.OrderBy) {
	case "":
		orderBy = OrderByDesc
		break
	case OrderByDesc:
		orderBy = OrderByDesc
		break
	case OrderByAsc:
		orderBy = OrderByAsc
		break
	default:
		err = fmt.Errorf("invalid order_by: %s", params.QueryType)
	}

	return limit, offset, queryType, orderBy, err
}

func (db *PostgresDB) Get(params *QueryParams) ([]SlowQueryLog, error) {
	limit, offset, queryType, orderBy, err := extractQueryParams(params)
	if err != nil {
		return nil, err
	}
	sqlStmt := "SELECT query, total_exec_time FROM public.pg_stat_statements"
	if queryType != "" {
		sqlStmt += fmt.Sprintf(" WHERE query ILIKE '%s%%'", queryType)
	}
	if orderBy != "" {
		sqlStmt += fmt.Sprintf(" ORDER BY total_exec_time %s", orderBy)
	}
	sqlStmt += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	rows, err := db.conn.Query(context.Background(), sqlStmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var resp []SlowQueryLog
	for rows.Next() {
		var log SlowQueryLog
		rows.Scan(&log.Query, &log.TotalExecTime)
		resp = append(resp, log)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (db *PostgresDB) Demo() error {
	return db.createDemoTable()
}

func (db *PostgresDB) createDemoTable() error {
	_, err := db.conn.Exec(context.Background(), "CREATE EXTENSION pg_stat_statements;")
	if err != nil {
		return err
	}

	_, err = db.conn.Exec(context.Background(), "DROP TABLE IF EXISTS users;")
	if err != nil {
		return err
	}

	_, err = db.conn.Exec(context.Background(), `CREATE TABLE users (
  		id SERIAL PRIMARY KEY,
  		first_name TEXT,
  		last_name TEXT,
  		email TEXT UNIQUE NOT NULL,
        phone TEXT
	);
	`)
	if err != nil {
		return err
	}

	sqlStmt := `
INSERT INTO users (first_name, last_name, email, phone)
VALUES ('Oliver', 'Andersson', 'oliver@example.com', 123456789)`
	_, err = db.conn.Exec(context.Background(), sqlStmt)
	if err != nil {
		return err
	}

	sqlStmt = "SELECT * FROM users WHERE first_name ILIKE 'O%'"
	_, err = db.conn.Exec(context.Background(), sqlStmt)
	if err != nil {
		return err
	}

	sqlStmt = "DELETE FROM users WHERE first_name ILIKE 'O%'"
	_, err = db.conn.Exec(context.Background(), sqlStmt)
	if err != nil {
		return err
	}

	return nil
}
