package dovecot

import (
	"io"
	"log"
	"net/http"
	"strings"
)

const dovecotDoveadmAPIKey = "AjshkdaEkjaad8sA6261aSdsj"
const apiURL = "http://modsoussi.com:7000/doveadm/v1"

func ListCommands() {
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		log.Panic(err)
	}
	req.Header.Add("Authorization", getAPIAuthorizationHeader())
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	log.Println(string(body))
}

func FetchMail() {
	jsonBody := `
[
	[
		"fetch",
		{
			"user": "mo@modsoussi.com",
			"field": [
				"mailbox"
			],
			"query": [
				"mailbox"
			]
		},
		"tag1"
	]
]
	`

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(jsonBody))
	if err != nil {
		log.Panic(err)
	}
	req.Header.Add("Authorization", getAPIAuthorizationHeader())
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	log.Println(string(body))
}
