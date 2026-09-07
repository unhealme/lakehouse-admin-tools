package hive

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"strings"
	"sync"

	_ "github.com/beltran/gohive/v2"
)

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type NullString struct{ sql.NullString }

func (s NullString) ValueOrNull() string {
	if s.Valid {
		return s.String
	} else {
		return "NULL"
	}
}

type HiveServerClient struct {
	Db *sql.DB
}

func (c *HiveServerClient) Close() (err error) {
	err = c.Db.Close()
	c.Db = nil
	return
}

func (c HiveServerClient) DescribeFormattedTable(db string, tbl string) (string, error) {
	table := escapeName(tbl)
	if db != "" {
		table = escapeName(db) + "." + table
	}
	rows, err := c.Db.Query(fmt.Sprintf("DESCRIBE FORMATTED %s", table))
	if err != nil {
		return "", err
	}

	desc := bufPool.Get().(*bytes.Buffer)
	desc.Reset()
	defer bufPool.Put(desc)

	writer := csv.NewWriter(desc)
	writer.Comma = '|'
	writer.UseCRLF = false
	for rows.Next() {
		var colName, dataType, comment NullString
		if err := rows.Scan(&colName, &dataType, &comment); err != nil {
			break
		}
		writer.Write([]string{
			colName.ValueOrNull(),
			dataType.ValueOrNull(),
			comment.ValueOrNull(),
		})
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	writer.Flush()
	return strings.TrimSuffix(desc.String(), "\n"), err
}

func (c *HiveServerClient) SetMaxConnections(n int) *HiveServerClient {
	if n > 1 {
		c.Db.SetMaxOpenConns(n)
	}
	return c
}

func (c HiveServerClient) ShowDatabasesLike(pattern ...string) (dbs []string, err error) {
	var rows *sql.Rows
	iterRows := func() {
		for rows.Next() {
			var db string
			if err := rows.Scan(&db); err != nil {
				break
			}
			dbs = append(dbs, db)
		}
		err = rows.Err()
	}

	var con *sql.Conn
	ctx := context.Background()
	if con, err = c.Db.Conn(ctx); err != nil {
		return
	}
	defer con.Close()

	if len(pattern) < 1 {
		if rows, err = con.QueryContext(ctx, "SHOW DATABASES"); err != nil {
			return
		}
		iterRows()
	} else {
		for _, p := range pattern {
			q := fmt.Sprintf("SHOW DATABASES LIKE %s", escapeValue(p))
			if rows, err = con.QueryContext(ctx, q); err != nil {
				return
			}
			iterRows()
			if err != nil {
				break
			}
		}
	}
	return
}

func (c HiveServerClient) ShowTablesLike(db string, pattern ...string) (tbls []string, err error) {
	var rows *sql.Rows
	iterRows := func() {
		for rows.Next() {
			var tbl string
			if err := rows.Scan(&tbl); err != nil {
				break
			}
			tbls = append(tbls, tbl)
		}
		err = rows.Err()
	}

	var con *sql.Conn
	ctx := context.Background()
	if con, err = c.Db.Conn(ctx); err != nil {
		return
	}
	defer con.Close()

	if len(pattern) < 1 {
		q := fmt.Sprintf("SHOW TABLES IN %s", escapeName(db))
		if rows, err = con.QueryContext(ctx, q); err != nil {
			return
		}
		iterRows()
	} else {
		for _, p := range pattern {
			q := fmt.Sprintf("SHOW TABLES IN %s LIKE %s", escapeName(db), escapeValue(p))
			if rows, err = con.QueryContext(ctx, q); err != nil {
				return
			}
			iterRows()
			if err != nil {
				break
			}
		}
	}
	return
}

// db can be empty to use default database
func (c HiveServerClient) ShowCreateTable(db string, tbl string) (string, error) {
	table := escapeName(tbl)
	if db != "" {
		table = escapeName(db) + "." + table
	}
	rows, err := c.Db.Query(fmt.Sprintf("SHOW CREATE TABLE %s", table))
	if err != nil {
		return "", err
	}

	cts := bufPool.Get().(*bytes.Buffer)
	cts.Reset()
	defer bufPool.Put(cts)

	for rows.Next() {
		var stmt string
		if err := rows.Scan(&stmt); err != nil {
			break
		}
		cts.WriteString(stmt)
		cts.WriteString("\n")
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return strings.TrimSuffix(cts.String(), "\n"), err
}

func NewHSClient(hiveUrl string) (*HiveServerClient, error) {
	db, err := sql.Open("hive", hiveUrl)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(2)
	return &HiveServerClient{db}, nil
}

func escapeName(name string) string {
	if strings.ContainsRune(name, '`') {
		name = strings.ReplaceAll(name, "`", "``")
	}
	return "`" + name + "`"
}

func escapeValue(value any) any {
	switch v := value.(type) {
	case string:
		if strings.ContainsRune(v, '\'') {
			v = strings.ReplaceAll(v, "'", "\\'")
		}
		return "'" + v + "'"
	default:
		return v
	}
}
