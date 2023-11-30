package env

import "os"

var mongoURIDev = "[MONGO_URI]"
var mySQLURIDev = "[SQL_URI]"
var signingKeyDev = "[SIGNING_KEY]"

func Env() string {
	if env := os.Getenv("MONGO_ENV"); env != "" {
		return env
	}

	return "development"
}

func MongoURI() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}

	return mongoURIDev
}

func MySQLURI() string {
	if uri := os.Getenv("MYSQL_URI"); uri != "" {
		return uri
	}

	return mySQLURIDev
}

func SigningKey() []byte {
	if key := os.Getenv("JWT_SIGNING_KEY"); key != "" {
		return []byte(key)
	}

	return []byte(signingKeyDev)
}
