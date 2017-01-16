package database

import (
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"database/sql"

	"github.com/Ssawa/LinkLetter/logger"
)

const (
	migrationTable  = "_migrations_"
	listTablesQuery = "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"
)

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

	sort.Sort(byMigrationIndex(migrations))
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

	tablename := ""
	for rows.Next() {
		rows.Scan(&tablename)
		if tablename == migrationTable {
			return true
		}
	}

	return false
}
