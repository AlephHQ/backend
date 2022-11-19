package env

import "os"

var mongoURIDev = "mongodb+srv://modsoussi:4ub5r12LIY4tmGn2@cluster0.gj2yyln.mongodb.net/?retryWrites=true&w=majority"

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
