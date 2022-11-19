package env

import "os"

var mongoURIDev = "mongodb+srv://modsoussi:4ub5r12LIY4tmGn2@cluster0.gj2yyln.mongodb.net/?retryWrites=true&w=majority"
var mySQLURIDev = "admin:7Lxb85t9l8Si0ZRiiSNY@tcp(mail-auth-db-0.c3hvckhvqu2r.us-east-1.rds.amazonaws.com:3306)/auth"

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