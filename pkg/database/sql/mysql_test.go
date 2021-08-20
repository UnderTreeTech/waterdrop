/*
 *
 * Copyright 2021 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package sql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/stretchr/testify/assert"
)

func TestMySQL(t *testing.T) {
	defer log.New(nil).Sync()
	dsn := "root@tcp(127.0.0.1:3306)/test?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8"
	db := NewMySQL(&Config{
		DriverName:        "mysql",
		DBName:            "test",
		DSN:               dsn,
		ReadDSN:           []string{dsn},
		Active:            10,
		Idle:              5,
		IdleTimeout:       time.Minute,
		QueryTimeout:      time.Minute,
		ExecTimeout:       time.Minute,
		TranTimeout:       time.Minute,
		SlowQueryDuration: time.Minute,
	})

	testPing(t, db)
	testTable(t, db)
	testExec(t, db)
	testQuery(t, db)
	testQueryRow(t, db)
	testPrepare(t, db)
	testPrepared(t, db)
	testTransaction(t, db)
	testMaster(t, db)
}

func testPing(t *testing.T, db *DB) {
	err := db.Ping(context.Background())
	assert.Nil(t, err)
}

func testTable(t *testing.T, db *DB) {
	drop := "DROP TABLE IF EXISTS `test`;"
	_, err := db.Exec(context.Background(), drop)
	assert.Nil(t, err)
	table := "CREATE TABLE IF NOT EXISTS `test` (`id` int(11) NOT NULL AUTO_INCREMENT, `name` varchar(32) NOT NULL DEFAULT '', PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"
	result, err := db.Exec(context.Background(), table)
	assert.Nil(t, err)
	id, err := result.LastInsertId()
	assert.Zero(t, id)
	assert.Nil(t, err)
	ra, err := result.RowsAffected()
	assert.Zero(t, ra)
	assert.Nil(t, err)
}

func testExec(t *testing.T, db *DB) {
	sql := "INSERT INTO test(name) VALUES(?)"
	result, err := db.Exec(context.Background(), sql, "test")
	assert.Nil(t, err)
	id, err := result.LastInsertId()
	assert.Equal(t, id, int64(1))
	assert.Nil(t, err)
	ra, err := result.RowsAffected()
	assert.Equal(t, ra, int64(1))
	assert.Nil(t, err)
}

func testQuery(t *testing.T, db *DB) {
	sql := "SELECT name FROM test WHERE name=?"
	rows, err := db.Query(context.Background(), sql, "test")
	assert.Nil(t, err)
	defer rows.Close()
	for rows.Next() {
		name := ""
		if err := rows.Scan(&name); err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, name, "test")
	}
}

func testQueryRow(t *testing.T, db *DB) {
	sql := "SELECT name FROM test WHERE name=?"
	name := ""
	row := db.QueryRow(context.Background(), sql, "test")
	err := row.Scan(&name)
	assert.Nil(t, err)
	assert.Equal(t, name, "test")
}

func testPrepare(t *testing.T, db *DB) {
	var (
		selsql  = "SELECT name FROM test WHERE name=?"
		execsql = "INSERT INTO test(name) VALUES(?)"
		name    = ""
	)
	selstmt, err := db.Prepare(context.Background(), selsql)
	assert.Nil(t, err)
	row := selstmt.QueryRow(context.Background(), "test1")
	err = row.Scan(&name)
	assert.NotNil(t, err)
	assert.Equal(t, err, sql.ErrNoRows)
	rows, err := selstmt.Query(context.Background(), "test")
	assert.Nil(t, err)
	defer rows.Close()
	for rows.Next() {
		rname := ""
		if err = rows.Scan(&rname); err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, rname, "test")
	}
	execstmt, err := db.Prepare(context.Background(), execsql)
	assert.Nil(t, err)
	_, err = execstmt.Exec(context.Background(), "test")
	assert.Nil(t, err)
}

func testPrepared(t *testing.T, db *DB) {
	sql := "SELECT name FROM test WHERE name=?"
	name := ""
	stmt := db.Prepared(context.Background(), sql)
	row := stmt.QueryRow(context.Background(), "test")
	err := row.Scan(&name)
	assert.Nil(t, err)
	assert.Equal(t, name, "test")
	err = stmt.Close()
	assert.Nil(t, err)
}

func testTransaction(t *testing.T, db *DB) {
	var (
		tx      *Tx
		err     error
		execSQL = "INSERT INTO test(name) VALUES(?)"
		selSQL  = "SELECT name FROM test WHERE name=?"
		txstmt  *Stmt
	)
	tx, err = db.Begin(context.TODO())
	defer func() {
		tx.Rollback()
	}()
	assert.Nil(t, err)
	txstmt, err = tx.Prepare(execSQL)
	assert.Nil(t, err)
	stmt := tx.Stmt(txstmt)
	assert.NotNil(t, stmt)
	_, err = tx.Exec(execSQL, "tx1")
	assert.Nil(t, err)
	_, err = tx.Exec(execSQL, "tx1")
	assert.Nil(t, err)
	// query
	rows, err := tx.Query(selSQL, "tx2")
	assert.Nil(t, err)
	rows.Close()
	// queryrow
	var name string
	row := tx.QueryRow(selSQL, "noexist")
	err = row.Scan(&name)
	assert.NotNil(t, err)
	assert.Equal(t, err, sql.ErrNoRows)
	err = tx.Commit()
	assert.Nil(t, err)
}

func testMaster(t *testing.T, db *DB) {
	master := db.Master()
	assert.NotNil(t, master)
	assert.Zero(t, len(master.read))
	assert.Equal(t, master.write, db.write)
}

// BenchmarkMySQLQuery bench mysql query
func BenchmarkMySQLQuery(b *testing.B) {
	cfg := &Config{
		DriverName:        "mysql",
		DBName:            "test",
		DSN:               "root@tcp(127.0.0.1:3306)/test?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8",
		Active:            10,
		Idle:              5,
		IdleTimeout:       time.Minute,
		QueryTimeout:      time.Minute,
		ExecTimeout:       time.Minute,
		TranTimeout:       time.Minute,
		SlowQueryDuration: time.Minute,
	}
	db := NewMySQL(cfg)
	defer db.Close()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sql := "SELECT name FROM test WHERE name=?"
			rows, err := db.Query(context.TODO(), sql, "test")
			if err == nil {
				for rows.Next() {
					var name string
					if err = rows.Scan(&name); err != nil {
						break
					}
				}
				rows.Close()
			}
		}
	})
}

// BenchmarkMySQLInsert bench mysql insert
func BenchmarkMySQLInsert(b *testing.B) {
	cfg := &Config{
		DriverName:        "mysql",
		DBName:            "test",
		DSN:               "root@tcp(127.0.0.1:3306)/test?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8",
		Active:            10,
		Idle:              5,
		IdleTimeout:       time.Minute,
		QueryTimeout:      time.Minute,
		ExecTimeout:       time.Minute,
		TranTimeout:       time.Minute,
		SlowQueryDuration: time.Minute,
	}
	db := NewMySQL(cfg)
	defer db.Close()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sql := "INSERT INTO test(name) VALUES(?)"
			_, err := db.Exec(context.Background(), sql, xstring.RandomString(16))
			if err != nil {
				break
			}
		}
	})
}
