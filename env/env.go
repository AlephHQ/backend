package env

import "os"

var mongoURIDev = ""
var mySQLURIDev = ""
var signingKeyDev = "jkjhasd.uyjflk-ajksh_jghasd86aisdhjkg-jksdjlh.gsd96"
var domainDev = ""

func Env() string {
	if env := os.Getenv("ALEPH_ENV"); env != "" {
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

func Domain() string {
	if dom := os.Getenv("DOMAIN"); dom != "" {
		return dom
	}

	return domainDev
}
