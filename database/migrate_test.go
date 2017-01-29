package database

import (
	"os"
	"testing"

	"regexp"

	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Ssawa/LinkLetter/logger"
	"github.com/stretchr/testify/assert"
)

func TestGetSplitToken(t *testing.T) {
	assert.Equal(t, ";", getSplitToken("NothingToSeeHere"))

	assert.Equal(t, fullSplitToken, getSplitToken(fmt.Sprintf("blah blah %s blah blah", fullSplitToken)))
}

func TestSplitQueries(t *testing.T) {
	baseContent := "  Query one %[1]s Querytwo%[1]squerythree"
	expected := []string{"Query one", "Querytwo", "querythree"}
	assert.Equal(t, expected, splitQueries(fmt.Sprintf(baseContent, ";")))
	assert.Equal(t, expected, splitQueries(fmt.Sprintf(baseContent, fullSplitToken)))
}

func TestExecSQLFile(t *testing.T) {
	db, mock, _ := sqlmock.New()

	// Tests a multiline statement
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE Persons (.*);").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("Something with; semi-colons").WillReturnResult(sqlmock.NewResult(0, 0))
	tx, _ := db.Begin()
	err := execSQLFile(tx, "test_assets/migrations/1_first.sql")
	assert.Nil(t, err, err)

	assert.Nil(t, mock.ExpectationsWereMet())

	// Tests multiple statements in file
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing2")).WillReturnResult(sqlmock.NewResult(0, 0))
	tx, _ = db.Begin()
	err = execSQLFile(tx, "test_assets/migrations/2_second.sql")
	assert.Nil(t, err, err)
	tx.Exec("SELECT * FROM testing")
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestPosInSlice(t *testing.T) {
	array := []string{"test1", "test2", "test3"}
	assert.Equal(t, 0, posInSlice(array, "test1"))
	assert.Equal(t, -1, posInSlice(array, "test4"))
}

func TestGetMigrationIndex(t *testing.T) {
	index, err := getMigrationIndex("55_test.sql")
	assert.Equal(t, nil, err)
	assert.Equal(t, 55, index)
}

func TestGetMigrationsInOrder(t *testing.T) {
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()
	os.Chdir("test_assets")

	migrations := getMigrationsInOrder()

	assert.Len(t, migrations, 3)
	assert.Equal(t, "1_first.sql", migrations[0])
	assert.Equal(t, "2_second.sql", migrations[1])
	assert.Equal(t, "10_tenth.sql", migrations[2])
}

func TestDoesMigrationTableExist(t *testing.T) {
	outputColumns := []string{"table_name"}
	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows(outputColumns))
	assert.False(t, doesMigrationTableExist(db))
	assert.Nil(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows(outputColumns).AddRow("nottablename"))
	assert.False(t, doesMigrationTableExist(db))
	assert.Nil(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows(outputColumns).AddRow(migrationTable))
	assert.True(t, doesMigrationTableExist(db))
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestCreateMigrationTable(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectExec(regexp.QuoteMeta(createMigrationTableQuery)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(createFirstEntryQuery)).WillReturnResult(sqlmock.NewResult(0, 0))
	createMigrationTable(db)
	assert.Nil(t, mock.ExpectationsWereMet())

}

func TestGetCurrentMigration(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(getCurrentMigrationQuery)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("1_test.sql"))
	assert.Equal(t, "1_test.sql", getCurrentMigration(db))
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestCreateMigrationTableIfNeeded(t *testing.T) {
	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows([]string{"table_name"}))
	mock.ExpectExec(regexp.QuoteMeta(createMigrationTableQuery)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(createFirstEntryQuery)).WillReturnResult(sqlmock.NewResult(0, 0))
	createMigrationTableIfNeeded(db)
	assert.Nil(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow(migrationTable))
	createMigrationTableIfNeeded(db)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestGetCurrentMigrationIndex(t *testing.T) {
	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(regexp.QuoteMeta(getCurrentMigrationQuery)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("2_test.sql"))
	assert.Equal(t, 3, getCurrentMigrationIndex(db, []string{"0_test.sql", "1_test.sql", "2_test.sql"}))

	mock.ExpectQuery(regexp.QuoteMeta(getCurrentMigrationQuery)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("4_test.sql"))
	assert.Equal(t, -1, getCurrentMigrationIndex(db, []string{"0_test.sql", "1_test.sql", "2_test.sql"}))

	mock.ExpectQuery(regexp.QuoteMeta(getCurrentMigrationQuery)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(""))
	assert.Equal(t, 0, getCurrentMigrationIndex(db, []string{"0_test.sql", "1_test.sql", "2_test.sql"}))
}

func TestPerformNeededMigrations(t *testing.T) {
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()
	os.Chdir("test_assets2")

	db, mock, _ := sqlmock.New()
	mock.ExpectBegin()
	tx, _ := db.Begin()

	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing2")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing3")).WillReturnResult(sqlmock.NewResult(0, 0))
	performNeededMigrations(tx, []string{"1_first.sql", "2_second.sql", "3_third.sql"}, 0)
	assert.Nil(t, mock.ExpectationsWereMet())

	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing2")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing3")).WillReturnResult(sqlmock.NewResult(0, 0))
	performNeededMigrations(tx, []string{"1_first.sql", "2_second.sql", "3_third.sql"}, 1)
	assert.Nil(t, mock.ExpectationsWereMet())

	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing3")).WillReturnResult(sqlmock.NewResult(0, 0))
	performNeededMigrations(tx, []string{"1_first.sql", "2_second.sql", "3_third.sql"}, 2)
	assert.Nil(t, mock.ExpectationsWereMet())

	performNeededMigrations(tx, []string{"1_first.sql", "2_second.sql", "3_third.sql"}, 3)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestUpdateMigrationTable(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectBegin()
	tx, _ := db.Begin()

	mock.ExpectExec(regexp.QuoteMeta(updateMigrationQuery)).WithArgs("second_file").WillReturnResult(sqlmock.NewResult(1, 1))
	updateMigrationTable(tx, []string{"first_file", "second_file"})
	assert.Nil(t, mock.ExpectationsWereMet())
}

// Because DoMigrations is essentially a wrapper around all these other functions, there's a lot of repeated test code here.
// Is there any way we can factor some of this stuff out?

func TestDoMigrationsCouldNotFind(t *testing.T) {
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()
	os.Chdir("test_assets2")

	log := logger.CreateDummyLogger()

	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow(migrationTable))
	mock.ExpectQuery(regexp.QuoteMeta(getCurrentMigrationQuery)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("doesNotExist.sql"))

	DoMigrations(db)
	assert.Nil(t, mock.ExpectationsWereMet())
	assert.Equal(t, "WARNING: Could not find migration listed in database on filesystem\n", log.Warning.Last())
}

func TestDoMigrationsUpToDate(t *testing.T) {
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()
	os.Chdir("test_assets2")

	log := logger.CreateDummyLogger()

	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow(migrationTable))
	mock.ExpectQuery(regexp.QuoteMeta(getCurrentMigrationQuery)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3_third.sql"))

	DoMigrations(db)
	assert.Nil(t, mock.ExpectationsWereMet())
	assert.Equal(t, "DEBUG: The database seems to be up to date\n", log.Debug.Last())
}

func TestDoMigrationsExecute(t *testing.T) {
	originalCWD, _ := os.Getwd()
	defer func() { os.Chdir(originalCWD) }()
	os.Chdir("test_assets2")

	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow(migrationTable))
	mock.ExpectQuery(regexp.QuoteMeta(getCurrentMigrationQuery)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("1_first.sql"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing2")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SELECT * FROM testing3")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(updateMigrationQuery)).WithArgs("3_third.sql").WillReturnResult(sqlmock.NewResult(1, 1))

	DoMigrations(db)
	assert.Nil(t, mock.ExpectationsWereMet())
}
