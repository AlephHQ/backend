package command

import "aleph/backend/utils"

func getTag() string {
	return utils.RandStr(7)
}
