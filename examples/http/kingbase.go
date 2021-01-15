/*
 *
 * Copyright 2020 waterdrop authors.
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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/server"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/database/sql"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/UnderTreeTech/waterdrop/pkg/utils/xtime"

	"github.com/Masterminds/squirrel"

	"github.com/gin-gonic/gin"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	defer log.New(nil).Sync()
	srv := server.New(nil)

	g := srv.Group("/api")
	{
		g.GET("/kingbase", kingbase)
	}
	srv.Start()
	<-c
	srv.Stop(context.Background())
}

type TSystemInfo struct {
	FsystemUsedTime string
	FsystemLicense  string
}

func kingbase(c *gin.Context) {
	cfg := &sql.Config{
		DBName:            "db_global",
		DriverName:        "postgres",
		DSN:               "user=SYSTEM password=123456 dbname=db_global host=192.168.1.163 port=54321 sslmode=disable timezone=Asia/Shanghai",
		ReadDSN:           []string{"user=SYSTEM password=123456 dbname=db_global host=10.206.83.79 port=54321 sslmode=disable timezone=Asia/Shanghai"},
		Active:            20,
		Idle:              10,
		IdleTimeout:       time.Hour * 1,
		QueryTimeout:      time.Millisecond * 200,
		ExecTimeout:       time.Millisecond * 200,
		TranTimeout:       time.Millisecond * 200,
		SlowQueryDuration: time.Millisecond * 200,
	}
	db := sql.NewPostgres(cfg)
	defer db.Close()
	err := db.Ping(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	// select
	condition := make(map[string]interface{})
	condition["Fsystem_license"] = "222"
	sql, args, err := squirrel.Select("*").From("t_system_info").Where(condition).Limit(1).PlaceholderFormat(squirrel.Dollar).ToSql()
	fmt.Println(sql, args, err)
	var license TSystemInfo
	err = db.QueryRow(context.Background(), sql, args...).Scan(&license.FsystemUsedTime, &license.FsystemLicense)
	if err != nil {
		fmt.Println("query data fail", err)
	} else {
		fmt.Println(license)
	}

	// update
	sql, args, err = squirrel.Update("t_system_info").
		Set("Fsystem_used_time", xtime.Now().Format(xtime.DateTimeFormat)).
		Where(condition).
		//Limit(1).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	fmt.Println(sql, args, err)
	res, err := db.Exec(context.Background(), sql, args...)
	fmt.Println(res, err)

	// insert
	sql, args, err = squirrel.Insert("t_system_info").Columns("Fsystem_used_time", "Fsystem_license").
		Values(xtime.Now().Format(xtime.DateTimeFormat), xstring.RandomString(32)).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	fmt.Println(sql, args, err)
	res, err = db.Exec(context.Background(), sql, args...)
	fmt.Println(res, err)

	// delete
	condition["Fsystem_license"] = "22"
	sql, args, err = squirrel.Delete("t_system_info").Where(condition).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	fmt.Println(sql, args, err)
	res, err = db.Exec(context.Background(), sql, args...)
	fmt.Println(res, err)

	// transaction
	tx, err := db.Begin(context.Background())
	if err != nil {
		fmt.Println("start transaction fail", err)
		return
	}

	condition["Fsystem_license"] = "2"
	sql, args, err = squirrel.Insert("t_system_info").Columns("Fsystem_used_time", "Fsystem_license").
		Values(xtime.Now().Format(xtime.DateTimeFormat), xstring.RandomString(32)).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	fmt.Println(sql, args, err)
	res, err = db.Exec(context.Background(), sql, args...)
	fmt.Println(res, err)
	if err != nil {
		tx.Rollback()
	}

	sql, args, err = squirrel.Update("t_system_info").
		Set("Fsystem_used_time", xtime.Now().Format(xtime.DateTimeFormat)).
		Where(condition).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	fmt.Println(sql, args, err)
	res, err = db.Exec(context.Background(), sql, args...)
	fmt.Println(res, err)
	if err != nil {
		tx.Rollback()
	}

	sql, args, err = squirrel.Delete("t_system_info").Where(condition).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	fmt.Println(sql, args, err)
	res, err = db.Exec(context.Background(), sql, args...)
	fmt.Println(res, err)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

	//squirrel.
	time.Sleep(time.Second)
	c.JSON(200, "ping")
}
