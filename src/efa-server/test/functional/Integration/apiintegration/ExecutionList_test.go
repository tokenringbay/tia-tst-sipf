package apiintegration

import (
	"efa-server/domain"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/database"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func cleanupDB(Database *database.Database) {
	Database.Close()
	os.Remove(constants.TESTDBLocation + "Integration")
}

func TestExecutionList_Limit(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()

	var it uint
	for it = 1; it <= 20; it++ {
		if it%2 == 0 {
			devUC.Db.CreateExecutionLog(&domain.ExecutionLog{ID: it, UUID: "1001", Status: "Completed", Command: "test"})
		} else {
			devUC.Db.CreateExecutionLog(&domain.ExecutionLog{ID: it, UUID: "1001", Status: "Failed", Command: "test"})
		}

	}

	ExecutionList, properr := infra.GetUseCaseInteractor().Db.GetExecutionLogList(0, "all")
	assert.NoError(t, properr)
	assert.Equal(t, 20, len(ExecutionList))

	ExecutionList, properr = infra.GetUseCaseInteractor().Db.GetExecutionLogList(5, "all")
	assert.NoError(t, properr)
	assert.Equal(t, 5, len(ExecutionList))

	ExecutionList, properr = infra.GetUseCaseInteractor().Db.GetExecutionLogList(10, "all")
	assert.NoError(t, properr)
	assert.Equal(t, 10, len(ExecutionList))

	ExecutionList, properr = infra.GetUseCaseInteractor().Db.GetExecutionLogList(0, "failed")
	assert.NoError(t, properr)
	assert.Equal(t, 10, len(ExecutionList))

	ExecutionList, properr = infra.GetUseCaseInteractor().Db.GetExecutionLogList(3, "failed")
	assert.NoError(t, properr)
	assert.Equal(t, 3, len(ExecutionList))

	ExecutionList, properr = infra.GetUseCaseInteractor().Db.GetExecutionLogList(0, "succeeded")
	assert.NoError(t, properr)
	assert.Equal(t, 10, len(ExecutionList))

	ExecutionList, properr = infra.GetUseCaseInteractor().Db.GetExecutionLogList(8, "succeeded")
	assert.NoError(t, properr)
	assert.Equal(t, 8, len(ExecutionList))
}
