package database

// If writing your own database migration system from scratch sounds silly and dangerous
// to you, then congratulations, because you're paying attention. The only thing I can say
// in my defense is that this isn't the kind of thing you do because you think it will be a
// fun little project.
//
// I'll say it until I'm blue in the face, I'd like this project to be as simple and straight
// forward as possible. However, neither "simple" nor "straightforward" belong anywhere near
// "database schema management." In production environments, I use the sophisticated and
// tested libraries. The Alembics, the Rakes, the ORMs that do it for you automatically. They're
// great and necessary but you need to spend the time learning them.
//
// All I wanted was a simple system where you drop a SQL file into a folder and have it execute
// in the proper order when it needs to be. I didn't care about "up" migrations and "down" migrations
// or about trying to extract away the raw SQL. I couldn't find anything like it, so I had to write
// this.

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"github.com/Ssawa/LinkLetter/logger"
)

// I was at a company once where the CTO yelled at us because we were writing our SQL all throughout
// the files rather than in designated areas (relax, relax, all executed SQL was getting sanitized)
// and so it was hard to test which were our problem queries. I propably rolled my eyes at the time
// but I guess it stuck because I still try to keep my queries in special areas, like this const
// section here, instead of inlining them.
//
// You might be wondering, aptly, why we define migrationTable as a variable and then not actually
// use it in our queries? Believe me, it bothers me too. The problem is that you can't pass the table
// name in as a query argument, the Postgres driver complains. So you're stuck doing string interpolation
// in Go. Unfortunately, you can't do this and have it be a constant. So besides the query looking ugly and
// unreadable, you'd have to make the regular variables, and I don't want them to be variables, I want them
// to be constants because that's what they are. So in the end I went with keeping everything clear and
// explicit in const and just repeating "_migrations_". This is not a good resolution, as this is something
// that will come up for all our tables and we'll want to make sure we can change table names easily without
// breaking all our queries.
const (
	migrationTable            = "_migrations_"
	listTablesQuery           = "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"
	createMigrationTableQuery = "CREATE TABLE _migrations_ (version TEXT NOT NULL)"
	createFirstEntryQuery     = "INSERT INTO _migrations_ VALUES ('')"
	getCurrentMigrationQuery  = "SELECT version FROM _migrations_ LIMIT 1"
	updateMigrationQuery      = "UPDATE _migrations_ SET version=$1"
)

func execSQLFile(tx *sql.Tx, file string) error {
	// Surprisingly, Golang's SQL package has no way of actually executing a sql file. Weird right?
	// The suggested way of doing this is to spawn a subprocess, executing something like
	// "cat file > psql ..." I'd really like to not have to do that. We'd have to make an assumption
	// that psql is on the developers path and I just know it would be a headache. I mean, we already
	// have a connection to postgres, we should just use that, right? So what this does is split the
	// file into separate commands using a known token and executes each one at a time. Unfortunately
	// this token can't (or shouldn't) just be ";", because then we get into cases where what happens
	// if we have subqueries or just strings with escaped semi-colons? The standard convention is to
	// instead just use some kind of known comment string to split up statements. It's a pain and
	// requires an extra step, but it's safer. Maybe we should do something where it can try
	// semi colon by default unless it sees the marker? Something like https://bitbucket.org/liamstask/goose

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	contents := string(buf)
	statements := strings.Split(contents, ";")
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			logger.Debug.Printf("Executing: %s", statement)
			_, err = tx.Exec(statement)
			if err != nil {
				return err
			}
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

// What we're doing here is specifying a new type declaration for
// []string, so that we can add custom methods for sorting. These
// will be called by sort.Stable
//
// I actually never would have thought to do it this way. It seems
// to be the typical Golang way and I think it's kind of elegant.
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

// getMigrationsInOrder gets all sql files in the migrations folder that follow
// the convention "num_desc.sql" and orders them in numerically ascending
// order.
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

// createMigrationTableIfNeeded checks to see if the migration table exists and
// if it doesn't creates it.
func createMigrationTableIfNeeded(db *sql.DB) {
	if !doesMigrationTableExist(db) {
		logger.Info.Printf("Migration table does not yet exist. Creating it")
		createMigrationTable(db)
	}
}

// getCurrentMigrationIndex retrieves the currently last executed file
// from the database
func getCurrentMigrationIndex(db *sql.DB, migrations []string) int {
	if currentMigration := getCurrentMigration(db); currentMigration != "" {
		startIndex := posInSlice(migrations, currentMigration)
		if startIndex == -1 {
			return startIndex
		}
		return startIndex + 1
	}
	return 0
}

// performNeededMigrations determines what migrations need to be performed,
// based on the current migration index, and executes them.
func performNeededMigrations(tx *sql.Tx, migrations []string, startIndex int) {
	for _, migration := range migrations[startIndex:] {
		logger.Info.Printf("Perfoming migration: %s", migration)
		err := execSQLFile(tx, fmt.Sprintf("migrations/%s", migration))
		if err != nil {
			logger.Error.Printf("Error occured executing migration. Rolling back and panicking")
			tx.Rollback()
			panic(err)
		}
	}
}

// updateMigrationTable updates the migration table with the latest migration
func updateMigrationTable(tx *sql.Tx, migrations []string) {
	_, err := tx.Exec(updateMigrationQuery, migrations[len(migrations)-1])
	if err != nil {
		logger.Error.Printf("Unable to update migration table. Rolling back")
		tx.Rollback()
		panic(err)
	}
}

// DoMigrations brings the supplied database up to date with current state of
// the migration files
func DoMigrations(db *sql.DB) {
	createMigrationTableIfNeeded(db)

	migrations := getMigrationsInOrder()

	startIndex := getCurrentMigrationIndex(db, migrations)

	if startIndex == -1 {
		logger.Warning.Printf("Could not find migration listed in database on filesystem")
		return
	}

	if startIndex == len(migrations) {
		logger.Debug.Printf("The database seems to be up to date")
		return
	}

	logger.Info.Printf("Performing database migrations")

	tx, err := db.Begin()

	if err != nil {
		logger.Error.Printf("Could not start transaction for migrations")
		panic(err)
	}

	performNeededMigrations(tx, migrations, startIndex)

	logger.Info.Printf("Finished performing migrations. Updating migration table")
	updateMigrationTable(tx, migrations)

	tx.Commit()
}
