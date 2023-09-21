package db

import (
	"errors"
	"github.com/NeverStopDreamingWang/hgee"
	"strings"
)

// 连接 Mysql 数据库
func MysqlConnect(UseDataBases ...string) (*MysqlDB, error) {
	if len(UseDataBases) == 0 { // 使用默认数据库
		UseDataBases = append(UseDataBases, "default")
	}

	for _, Name := range UseDataBases {
		database, ok := hgee.Settings.DATABASES[Name]
		if ok == true && strings.ToLower(database.ENGINE) == "mysql" {
			databaseObject, err := MetaMysqlConnect(database)
			if err != nil {
				continue
			}
			databaseObject.Name = Name
			return databaseObject, err
		}
	}
	return nil, errors.New("连接 Mysql 错误！")
}

// 连接 Sqlite3 数据库
func Sqlite3Connect(UseDataBases ...string) (*Sqlite3DB, error) {
	if len(UseDataBases) == 0 { // 使用默认数据库
		UseDataBases = append(UseDataBases, "default")
	}

	for _, Name := range UseDataBases {
		Database, ok := hgee.Settings.DATABASES[Name]
		if ok == true && strings.ToLower(Database.ENGINE) == "sqlite3" {
			databaseObject, err := MetaSqlite3Connect(Database)
			if err != nil {
				continue
			}
			databaseObject.Name = Name
			return databaseObject, err
		}
	}
	return nil, errors.New("连接 Sqlite3 错误！")
}