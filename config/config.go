package config

import (
	"flag"
	"os"
	"strconv"
)

// Config contains all the configurable data needed to run the
// appliation
type Config struct {
	WebPort     int
	SqlPort     int
	SqlHost     string
	SqlDB       string
	SqlUser     string
	SqlPassword string
	SecretKey   string
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
		// WebPort doesn't follow the LINKLETTER namespacing to be complacint with Heroku
		WebPort:     GetEnvIntDefault("PORT", 8080),
		SqlPort:     GetEnvIntDefault("LINKLETTER_SQLPORT", 9753),
		SqlHost:     GetEnvStringDefault("LINKLETTER_SQLHOST", "127.0.0.1"),
		SqlDB:       GetEnvStringDefault("LINKLETTER_SQLDB", "linkletter"),
		SqlUser:     GetEnvStringDefault("LINKLETTER_SQLUSER", "linkletter"),
		SqlPassword: GetEnvStringDefault("LINKLETTER_SQLPASSWORD", "pass"),
		SecretKey:   GetEnvStringDefault("LINKLETTER_SECRETKEY", "secret123"),
	}

	flag.IntVar(&conf.WebPort, "webPort", conf.WebPort, "The port to run the web application on")
	flag.IntVar(&conf.SqlPort, "sqlPort", conf.SqlPort, "The port the SQL server is running on")
	flag.StringVar(&conf.SqlHost, "sqlHost", conf.SqlHost, "The SQL server host")
	flag.StringVar(&conf.SqlDB, "sqlDB", conf.SqlDB, "The SQL database name to connect to")
	flag.StringVar(&conf.SqlUser, "sqlUser", conf.SqlUser, "The username to use when connecting to SQL database")
	flag.StringVar(&conf.SqlPassword, "sqlPassword", conf.SqlPassword, "The password to use when conneting to SQL database")
	flag.StringVar(&conf.SecretKey, "secretKey", conf.SecretKey, "The secret key to use to sign cookies")

	flag.Parse()
	return conf
}
