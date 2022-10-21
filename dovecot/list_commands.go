package dovecot

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const dovecotDoveadmAPIKey = "AjshkdaEkjaad8sA6261aSdsj"

func ListCommands() {
	url := "http://modsoussi.com:7000/doveadm/v1"
	apiAuthorization := fmt.Sprintf("X-Dovecot-API %s", getBase64EncodedDoveadmAPIKey())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Panic(err)
	}
	req.Header.Add("Authorization", apiAuthorization)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	log.Println(string(body))
}
