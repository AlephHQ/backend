package dovecot

import (
	"encoding/base64"
	"fmt"
)

func getBase64EncodedDoveadmAPIKey() string {
	return base64.StdEncoding.EncodeToString([]byte(dovecotDoveadmAPIKey))
}

func getAPIAuthorizationHeader() string {
	return fmt.Sprintf("X-Dovecot-API %s", getBase64EncodedDoveadmAPIKey())
}
