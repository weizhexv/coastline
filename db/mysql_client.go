package db

import (
	"coastline/tlog"
	"coastline/vconfig"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var mysql = initDB()

func initDB() *sqlx.DB {
	db := sqlx.MustConnect(vconfig.DbType(), vconfig.DbUrl())
	db.SetMaxOpenConns(vconfig.DbMaxOpenConns())
	db.SetMaxIdleConns(vconfig.DbMaxIdleConns())
	tlog.Entry().Info("init mysql successfully")

	return db
}

func Mysql() *sqlx.DB {
	return mysql
}
