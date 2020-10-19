package infra

import (
	"efa-server/gateway"
	"efa-server/infra/database"
	"efa-server/usecase"
	"sync"
)

//DeviceInteractor provides reciever object for UseCases
var DeviceInteractor *usecase.DeviceInteractor
var once sync.Once

//GetUseCaseInteractor gets a singleton instance of DeviceInteractor
//with DatabaseRepository,FabricAdapter initialized
func GetUseCaseInteractor() *usecase.DeviceInteractor {
	once.Do(func() {
		//Initialize the DatabaseRepository Gateway
		DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
		FabricAdapter := gateway.FabricAdapter{}
		DeviceInteractor = &usecase.DeviceInteractor{
			Db:                   &DatabaseRepository,
			DeviceAdapterFactory: gateway.DeviceAdapterFactory,
			FabricAdapter:        &FabricAdapter}

	})
	return DeviceInteractor
}
