package appcontext

import (
	"context"
	"efa-server/infra/constants"
	nlog "github.com/sirupsen/logrus"
	"runtime"
)

//ContextType represents type of the context, used to in the application
type ContextType int

var httpContext = context.Background()

const (
	//RequestIDKey represents the identifier of each REST request
	RequestIDKey ContextType = iota

	//UseCaseName represents the identifier of the EFA usecase e.g: configure fabric, deconfigure fabric etc
	UseCaseName

	//DeviceName represents the identifier of the switching device participating in the fabric
	DeviceName

	//FabricName represents the identifier of fabric
	FabricName

	//FabricType represent the fabric type
	FabricType

	//IPPair depicts the pair of IP Address used for NON-CLOS fabric
	IPPair
)

func getContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

//DecorateRuntimeContext represents a logger containing the "functionName" and "lineNumber" as fields
func DecorateRuntimeContext(logger *nlog.Entry) *nlog.Entry {
	if pc, _, line, ok := runtime.Caller(1); ok {
		fName := runtime.FuncForPC(pc).Name()
		return logger.WithField("line", line).WithField("func", fName)
	}
	return logger
}

// Logger returns a zap logger with as much context as possible
func Logger(ctx context.Context) *nlog.Entry {
	newLogger := nlog.WithFields(nlog.Fields{
		"App": constants.ApplicationName,
	})
	if ctx != nil {
		if ctxReqID, ok := ctx.Value(RequestIDKey).(string); ok {
			newLogger = newLogger.WithFields(nlog.Fields{
				"rqId": ctxReqID,
			})
		}
		if useCase, ok := ctx.Value(UseCaseName).(string); ok {
			newLogger = newLogger.WithFields(nlog.Fields{
				"UseCase": useCase,
			})
		}
		if device, ok := ctx.Value(DeviceName).(string); ok {
			newLogger = newLogger.WithFields(nlog.Fields{
				"Device": device,
			})
		}
		if fabric, ok := ctx.Value(FabricName).(string); ok {
			newLogger = newLogger.WithFields(nlog.Fields{
				"Fabric": fabric,
			})
		}
		if fabricType, ok := ctx.Value(FabricType).(string); ok {
			newLogger = newLogger.WithFields(nlog.Fields{
				"FabricType": fabricType,
			})
		}
	}
	return newLogger
}

//LoggerAndContext returns a zap logger and context
func LoggerAndContext(reqID string) (*nlog.Entry, context.Context) {
	reqCtx := getContext(httpContext, reqID)
	logger := Logger(reqCtx)
	logger.Info("Request ID Created")
	return logger, reqCtx
}
