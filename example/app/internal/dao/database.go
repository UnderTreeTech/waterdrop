package dao

import (
	"fmt"

	"github.com/UnderTreeTech/waterdrop/pkg/database/redis"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/conf"
	"github.com/UnderTreeTech/waterdrop/pkg/database/mysql"
)

var d *dao

type dao struct {
	db    *mysql.DB
	redis *redis.Redis
}

func NewDao() *dao {
	db := NewMySQL()
	r := NewRedis()

	ds := &dao{
		db:    db,
		redis: r,
	}
	d = ds
	return ds
}

func (d *dao) Close() {
	d.db.Close()
	d.redis.Close()
}

func NewMySQL() *mysql.DB {
	config := &mysql.Config{}
	if err := conf.Unmarshal("mysql", config); err != nil {
		panic(fmt.Sprintf("unmarshal mysql config fail,err msg %s", err.Error()))
	}
	log.Infof("db config", log.Any("config", config))
	db := mysql.New(config)

	return db
}

func NewRedis() *redis.Redis {
	config := &redis.Config{}
	if err := conf.Unmarshal("redis", config); err != nil {
		panic(fmt.Sprintf("unmarshal redis config fail,err msg %s", err.Error()))
	}
	log.Infof("redis config", log.Any("config", config))

	redis := redis.New(config)
	return redis
}

func GetDao() *dao {
	return d
}

func GetRedis() *redis.Redis {
	return d.redis
}
