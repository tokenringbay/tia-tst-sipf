package usecase

import (
	"efa-server/gateway"
	"efa-server/gateway/appcontext"
	"efa-server/infra/constants"
	"efa-server/infra/database"
	"efa-server/infra/logging"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var AuditDBName = constants.TESTDBLocation + "audit"

func TestLogMessageInit(t *testing.T) {
	alog := logging.AuditLog{Request: &logging.Request{Command: "fabric configure:ConfigureFabric"}}
	ctx := alog.LogMessageInit()
	assert.NotNil(t, ctx.Value(appcontext.RequestIDKey).(string))
}

func TestLogMessageReceivedAndCompleted(t *testing.T) {
	database.Setup(AuditDBName)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	alog := logging.AuditLog{Request: &logging.Request{Command: "fabric configure:ConfigureFabric"}}
	ctx := alog.LogMessageInit()
	uuid, _ := ctx.Value(appcontext.RequestIDKey).(string)

	alog.LogMessageReceived()

	success := true
	statusMsg := "test status Message"
	alog.LogMessageEnd(&success, &statusMsg)

	executionLog, err := DatabaseRepository.GetExecutionLogByUUID(uuid)
	assert.Nil(t, err)
	assert.Equal(t, "fabric configure:ConfigureFabric", executionLog.Command)
	assert.Contains(t, executionLog.Status, "Completed")

}

func TestLogMessageReceivedAndFailed(t *testing.T) {
	database.Setup(AuditDBName)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	alog := logging.AuditLog{Request: &logging.Request{Command: "fabric configure:ConfigureFabric"}}
	ctx := alog.LogMessageInit()
	uuid, _ := ctx.Value(appcontext.RequestIDKey).(string)

	alog.LogMessageReceived()

	success := false
	statusMsg := "test status Message"
	alog.LogMessageEnd(&success, &statusMsg)

	executionLog, err := DatabaseRepository.GetExecutionLogByUUID(uuid)
	assert.Nil(t, err)
	assert.Equal(t, "fabric configure:ConfigureFabric", executionLog.Command)
	assert.Contains(t, executionLog.Status, "Failed")

}

func cleanupDB(Database *database.Database) {
	Database.Close()
	os.Remove(AuditDBName)
}
