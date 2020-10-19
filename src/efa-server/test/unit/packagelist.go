package unit

import (
	"efa-server/domain"
	"efa-server/infra/database"
	"efa-server/infra/logging"
	"efa-server/infra/rest"
	"efa-server/infra/util"
	"efa-server/usecase"
)

var (
	x = domain.ConfigNone
	y = usecase.SpineRole
	z = logging.COMPLETED
	m = util.AesDecrypt
	b = database.GetWorkingInstance
	d = rest.NewOpenAPIServer
)
