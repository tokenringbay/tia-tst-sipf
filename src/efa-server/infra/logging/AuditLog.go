package logging

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"efa-server/infra"
	"efa-server/infra/constants"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
)

//AuditLog Constants indicating the commands status on the Rest Server
const (
	RCVD      = "Recieved"
	COMPLETED = "Completed"
	FAILED    = "Failed"
)

//Request captures the Request Command and its parameters
type Request struct {
	Command string                 `json:"cmd"`
	Params  map[string]interface{} `json:"params"`
}

//AuditLog is the receiver object providing the Logging functionality
type AuditLog struct {
	Request   *Request
	StartTime time.Time
	Logger    *logrus.Entry
	ReqID     string
	Log       *domain.ExecutionLog
}

//LogMessageInit initializes the AuditLog and setups the logger with the Request
func (alog *AuditLog) LogMessageInit() context.Context {
	alog.ReqID = uuid.New().String()
	Logger, ctx := appcontext.LoggerAndContext(alog.ReqID)
	alog.Logger = Logger.WithFields(logrus.Fields{
		"request": alog.Request,
	})
	return ctx
}

//LogMessageReceived to be invoked when the Request messages is recieved on the server.
//It will logs the message to log file and audit log
func (alog *AuditLog) LogMessageReceived() {
	alog.Logger.Infoln(RCVD)
	alog.logStartAudit(RCVD)

}
func (alog *AuditLog) logStartAudit(status string) {
	alog.StartTime = time.Now()
	RequestBytes, _ := json.Marshal(alog.Request.Params)
	alog.Log = &domain.ExecutionLog{UUID: alog.ReqID, StartTime: alog.StartTime.Format(constants.DefaultTimeFormat),
		Status: status, Command: alog.Request.Command, Params: string(RequestBytes)}

	infra.GetUseCaseInteractor().Db.CreateExecutionLog(alog.Log)

}
func (alog *AuditLog) logEndAudit(status string) {

	alog.Log.EndTime = time.Now().Format(constants.DefaultTimeFormat)
	duration := time.Since(alog.StartTime)
	alog.Log.Status = fmt.Sprintf("%s(%s)", status, duration.String())
	infra.GetUseCaseInteractor().Db.UpdateExecutionLog(alog.Log)
}

//LogMessageEnd to indicate the completion of command.
func (alog *AuditLog) LogMessageEnd(success *bool, statusMsg *string) {
	if *statusMsg != "" {
		alog.Logger = alog.Logger.WithField("reason", statusMsg)
	}
	if *success {
		alog.Logger.Infoln(COMPLETED)
		alog.logEndAudit(COMPLETED)
	} else {
		alog.Logger.Infoln(FAILED)
		alog.Logger.Errorln(FAILED)
		alog.logEndAudit(FAILED)
	}
}
