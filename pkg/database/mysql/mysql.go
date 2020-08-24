package mysql

import (
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	_ "github.com/go-sql-driver/mysql"
)

// Config mysql config.
type Config struct {
	DBName            string        //db name
	DSN               string        // write data source name.
	ReadDSN           []string      // read data source name.
	Active            int           // pool
	Idle              int           // pool
	IdleTimeout       time.Duration // connect max life time.
	QueryTimeout      time.Duration // query sql timeout
	ExecTimeout       time.Duration // execute sql timeout
	TranTimeout       time.Duration // transaction sql timeout
	SlowQueryDuration time.Duration // slow query duration
}

// NewMySQL new db and retry connection when has error.
func New(c *Config) (db *DB) {
	if c.QueryTimeout == 0 || c.ExecTimeout == 0 || c.TranTimeout == 0 {
		panic("mysql must be set query/execute/transction timeout")
	}
	db, err := Open(c)
	if err != nil {
		log.Errorf("open mysql error", log.Any("err_msg", err))
		panic(err)
	}
	return
}
