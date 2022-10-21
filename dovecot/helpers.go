package dovecot

import "encoding/base64"

func getBase64EncodedDoveadmAPIKey() string {
	return base64.StdEncoding.EncodeToString([]byte(dovecotDoveadmAPIKey))
}
