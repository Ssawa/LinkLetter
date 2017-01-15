package config

import (
	"flag"
	"os"
	"strconv"
)

// Config contains all the configurable data needed to run the
// appliation
type Config struct {
	webPort     int
	sqlPort     int
	sqlHost     string
	sqlDB       string
	sqlUser     string
	sqlPassword string
}

// GetEnvStringDefault wraps os.Getenv to get an environment variable as a
// string and supporting a default option
func GetEnvStringDefault(env string, defaultValue string) string {
	value := os.Getenv(env)
	if value == "" {
		value = defaultValue
	}
	return value
}

// GetEnvIntDefault gets an environment variable as an int and supports
// a default option
func GetEnvIntDefault(env string, defaultInt int) int {
	i, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		i = defaultInt
	}
	return i
}

// ParseForConfig grabs required information from the program args
// and environment vairables and creates a Config object
func ParseForConfig() Config {
	conf := Config{
		webPort:     GetEnvIntDefault("LINKLETTER_WEBPORT", 8080),
		sqlPort:     GetEnvIntDefault("LINKLETTER_SQLPORT", 9753),
		sqlHost:     GetEnvStringDefault("LINKLETTER_SQLHOST", "127.0.0.1"),
		sqlDB:       GetEnvStringDefault("LINKLETTER_SQLDB", "linkletter"),
		sqlUser:     GetEnvStringDefault("LINKLETTER_SQLUSER", "linkletter"),
		sqlPassword: GetEnvStringDefault("LINKLETTER_SQLPASSWORD", "pass"),
	}

	flag.IntVar(&conf.webPort, "webPort", conf.webPort, "The port to run the web application on")
	flag.IntVar(&conf.sqlPort, "sqlPort", conf.sqlPort, "The port the SQL server is running on")
	flag.StringVar(&conf.sqlHost, "sqlHost", conf.sqlHost, "The SQL server host")
	flag.StringVar(&conf.sqlDB, "sqlDB", conf.sqlDB, "The SQL database name to connect to")
	flag.StringVar(&conf.sqlUser, "sqlUser", conf.sqlUser, "The username to use when connecting to SQL database")
	flag.StringVar(&conf.sqlPassword, "sqlPassword", conf.sqlPassword, "the password to use when conneting to SQL database")

	flag.Parse()
	return conf
}
