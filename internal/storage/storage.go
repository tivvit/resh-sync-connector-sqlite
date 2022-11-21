package storage

import (
	"database/sql"
	"github.com/curusarn/resh/record"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/tivvit/resh-sync-connector-sqlite/internal/reshtime"
	"math"
	"strconv"
	"strings"
	"time"
)

const numberOfFields = 20       // it is actually 16 this is a safe value
const oneBatchInsertLimit = 999 // limitation of SQLite https://stackoverflow.com/questions/7106016/too-many-sql-variables-error-in-django-with-sqlite3

func initDb(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?parseTime=True&loc=Local")
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

func LatestEntryPerDeviceId(db *sql.DB, devices map[string]struct{}) (map[string]string, error) {
	rows, err := db.Query("select `deviceId`,  max(`time`) from `records` GROUP BY `deviceId`")
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msg("closing latestTime query failed")
		}
	}(rows)

	latest := map[string]string{}

	for rows.Next() {
		var deviceId string
		var ts string
		err = rows.Scan(&deviceId, &ts)
		if err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", ts)
		if err != nil {
			return nil, err
		}
		_, ok := devices[deviceId]
		if len(devices) == 0 || ok {
			latest[deviceId] = reshtime.TimeToReshString(t)
		}
	}

	return latest, nil
}

func ReadEntries(db *sql.DB, latestFromDevice map[string]float64) ([]record.V1, error) {
	rows, err := db.Query("select `recordId`, `deviceId`, `sessionId`, `cmdLine`, `exitCode`, `time`, `flags`, " +
		"`home`, `pwd`, `realPwd`, `device`, `gitOriginRemote`, `duration`, `partOne`, `partsNotMerged`, " +
		"`sessionExit` from `records`")
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msg("closing read query failed")
		}
	}(rows)

	var records []record.V1
	for rows.Next() {
		var r record.V1
		var exitCode, flags sql.NullInt32
		var home, pwd, realPwd, device, gitOriginRemote, duration sql.NullString
		var partOne, partsNotMerged, sessionExit sql.NullBool
		var t time.Time
		err = rows.Scan(&r.RecordID, &r.DeviceID, &r.SessionID, &r.CmdLine, &exitCode, &t, &flags, &home, &pwd,
			&realPwd, &device, &gitOriginRemote, &duration, &partOne, &partsNotMerged, &sessionExit)
		if err != nil {
			return nil, err
		}
		// Filter out old records from known devices.
		// There is a more optimal solution using conditions in the SQL query but is not trivial to build it.
		if l, ok := latestFromDevice[r.DeviceID]; ok {
			if float64(t.Unix()) <= l {
				continue
			}
		}
		r.Time = reshtime.TimeToReshString(t)
		if exitCode.Valid {
			r.ExitCode = int(exitCode.Int32)
		}
		if flags.Valid {
			r.Flags = int(flags.Int32)
		}
		if home.Valid {
			r.Home = home.String
		}
		if pwd.Valid {
			r.Pwd = pwd.String
		}
		if realPwd.Valid {
			r.RealPwd = realPwd.String
		}
		if device.Valid {
			r.Device = device.String
		}
		if gitOriginRemote.Valid {
			r.GitOriginRemote = gitOriginRemote.String
		}
		if duration.Valid {
			r.Duration = duration.String
		}
		if partOne.Valid {
			r.PartOne = partOne.Bool
		}
		if partsNotMerged.Valid {
			r.PartsNotMerged = partsNotMerged.Bool
		}
		if sessionExit.Valid {
			r.SessionExit = sessionExit.Bool
		}
		records = append(records, r)
	}
	return records, nil
}

func storeRecords(db *sql.DB, records []record.V1) error {
	const insertQuery = "INSERT INTO `records`(`recordId`, `deviceId`, `sessionId`, `cmdLine`, `exitCode`, `time`, " +
		"`flags`, `home`, `pwd`, `realPwd`, `device`, `gitOriginRemote`, `duration`, `partOne`, " +
		"`partsNotMerged`,`sessionExit`) VALUES "
	const row = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	var inserts []string
	var values []interface{}

	for _, r := range records {
		inserts = append(inserts, row)
		tf, err := strconv.ParseFloat(r.Time, 64)
		if err != nil {
			return err
		}
		sec, nsec := math.Modf(tf)
		t := time.Unix(int64(sec), int64(nsec*(1e9)))
		values = append(values, r.RecordID, r.DeviceID, r.SessionID, r.CmdLine, r.ExitCode, t, r.Flags, r.Home,
			r.Pwd, r.RealPwd, r.Device, r.GitOriginRemote, r.Duration, r.PartOne, r.PartsNotMerged, r.SessionExit)
	}
	sqlStr := insertQuery + strings.Join(inserts, ",")

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return err
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Error().Err(err).Msg("closing write statement failed")
		}
	}(stmt)

	_, err = stmt.Exec(values...)
	return err
}

func StoreRecords(db *sql.DB, records []record.V1) error {
	batch := make([]record.V1, 0, oneBatchInsertLimit/numberOfFields) // Ensure to not go over internal SQLite limitations
	for _, r := range records {
		if (len(batch)+1)*numberOfFields > oneBatchInsertLimit {
			err := storeRecords(db, batch)
			if err != nil {
				return err
			}
			batch = batch[:0]
		}
		batch = append(batch, r)
	}
	err := storeRecords(db, batch)
	return err
}
