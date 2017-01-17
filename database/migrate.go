package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"github.com/Ssawa/LinkLetter/logger"
)

const (
	migrationTable            = "_migrations_"
	listTablesQuery           = "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"
	createMigrationTableQuery = "CREATE TABLE _migrations_ (version TEXT NOT NULL)"
	createFirstEntryQuery     = "INSERT INTO _migrations_ VALUES ('')"
	getCurrentMigrationQuery  = "SELECT version FROM _migrations_ LIMIT 1"
	updateMigrationQuery      = "UPDATE _migrations_ SET version=$1"
)

func execSQLFile(tx *sql.Tx, file string) error {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	contents := string(buf)
	statements := strings.Split(contents, ";")
	for _, statement := range statements {
		_, err = tx.Exec(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

// posInSlice simply finds the index of a string in a string slice
func posInSlice(array []string, value string) int {
	for index, element := range array {
		if value == element {
			return index
		}
	}
	return -1
}

// getMigrationIndex splits the filename of a migration file and returns
// the index number as an int
func getMigrationIndex(migration string) (int, error) {
	return strconv.Atoi(strings.Split(migration, "_")[0])
}

// This type definition, and following method declorations
// are used for ordering migration file names
type byMigrationIndex []string

func (array byMigrationIndex) Len() int {
	return len(array)
}
func (array byMigrationIndex) Swap(i, j int) {
	array[i], array[j] = array[j], array[i]
}
func (array byMigrationIndex) Less(i, j int) bool {
	iLoc, _ := getMigrationIndex(array[i])
	jLoc, _ := getMigrationIndex(array[j])
	return iLoc < jLoc
}

// getMigrationsInOrder asserts a file format of "num_desc.sql"
func getMigrationsInOrder() []string {
	migrations := []string{}
	files, err := ioutil.ReadDir("migrations")
	if err != nil {
		logger.Error.Printf("Error occured getting list of database migrations")
		panic(err)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {

			_, err := getMigrationIndex(f.Name())
			if err != nil {
				continue
			}

			migrations = append(migrations, f.Name())
		}
	}

	sort.Stable(byMigrationIndex(migrations))
	return migrations
}

// doesMigrationTableExist determines if the table used to track migrations
// already exists
func doesMigrationTableExist(db *sql.DB) bool {
	rows, err := db.Query(listTablesQuery)
	if err != nil {
		logger.Error.Printf("Unable to get database table information for migrations")
		panic(err)
	}
	defer rows.Close()

	var tablename string
	for rows.Next() {
		rows.Scan(&tablename)
		if tablename == migrationTable {
			return true
		}
	}

	return false
}

// createMigrationTable creates the table used for keeping track of migrations
func createMigrationTable(db *sql.DB) {
	_, err := db.Exec(createMigrationTableQuery)
	if err != nil {
		logger.Error.Printf("Unable to create migration table")
		panic(err)
	}
	_, err = db.Exec(createFirstEntryQuery)
	if err != nil {
		logger.Error.Printf("Unable to insert first migration entry")
		panic(err)
	}

}

// getCurrentMigration retrives the current migration listed in the database
func getCurrentMigration(db *sql.DB) string {
	var migration string
	err := db.QueryRow(getCurrentMigrationQuery).Scan(&migration)
	if err != nil {
		logger.Error.Printf("Unable to retrieve current migration")
		panic(err)
	}
	return migration
}

// DoMigrations brings the supplied database up to date with current state of
// the migration files
func DoMigrations(db *sql.DB) {
	if !doesMigrationTableExist(db) {
		logger.Info.Printf("Migration table does not yet exist. Creating it")
		createMigrationTable(db)
	}

	migrations := getMigrationsInOrder()

	var startIndex int
	if currentMigration := getCurrentMigration(db); currentMigration != "" {
		startIndex = posInSlice(migrations, currentMigration)
		if startIndex == -1 {
			logger.Warning.Printf("Could not find migration listed in database on filesystem. Aborting migration")
			return
		}
		startIndex++
	} else {
		startIndex = 0
	}

	if startIndex == len(migrations) {
		logger.Debug.Printf("The database seems to be up to date")
		return
	}

	logger.Info.Printf("Perfoming database migrations")
	tx, err := db.Begin()
	if err != nil {
		logger.Error.Printf("Could not start transaction for migrations")
		panic(err)
	}
	for _, migration := range migrations {
		logger.Info.Printf("Perfoming migration: %s", migration)
		err = execSQLFile(tx, fmt.Sprintf("migrations/%s", migration))
		if err != nil {
			logger.Error.Printf("Error occured executing migration. Rolling back and panicking")
			tx.Rollback()
			panic(err)
		}
	}

	logger.Info.Printf("Finished performing migrations. Updating migration table")
	_, err = tx.Exec(updateMigrationQuery, migrations[len(migrations)-1])
	if err != nil {
		logger.Error.Printf("Unable to update migration table. Rolling back")
		tx.Rollback()
		panic(err)
	}
	tx.Commit()
}
