package command

import "ncp/backend/utils"

func getTag() string {
	return utils.RandStr(7)
}
