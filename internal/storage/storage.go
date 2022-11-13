package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func initDb(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// create DB schema
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS `records` (`recordId` VARCHAR(255) NOT NULL PRIMARY KEY, `deviceId` VARCHAR(255) NOT NULL, `sessionId` VARCHAR(255) NOT NULL, `cmdLine` TEXT NOT NULL, `exitCode` INT, `time` DATETIME, `flags` INT, `home` TEXT, `pwd` TEXT, `realPwd` TEXT, `device` TEXT, `gitOriginRemote` TEXT, `duration` TEXT, `partOne` BOOL, `partsNotMerged` BOOL, `sessionExit` BOOL);"); err != nil {
		return nil, err
	}
	// create indexes
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS `deviceId` ON `records` (`deviceId`);"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS `time` ON `records` (`time`);"); err != nil {
		return nil, err
	}
	return db, nil

}

func ConnectDb(path string) (*sql.DB, error) {
	return initDb(path)
}

func LatestEntryPerDeviceId(db *sql.DB) (map[string]string, error) {
	// TODO
	return nil, nil
}
