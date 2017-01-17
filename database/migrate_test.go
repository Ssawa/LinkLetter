package database

import (
	"os"
	"testing"

	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows(outputColumns).AddRow("nottablename"))
	assert.False(t, doesMigrationTableExist(db))
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}

	mock.ExpectQuery(listTablesQuery).WillReturnRows(sqlmock.NewRows(outputColumns).AddRow(migrationTable))
	assert.True(t, doesMigrationTableExist(db))
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCreateMigrationTable(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectExec(regexp.QuoteMeta(createMigrationTableQuery)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(createFirstEntryQuery)).WillReturnResult(sqlmock.NewResult(0, 0))
	createMigrationTable(db)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetCurrentMigration(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(getCurrentMigrationQuery)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("1_test.sql"))
	assert.Equal(t, "1_test.sql", getCurrentMigration(db))
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
