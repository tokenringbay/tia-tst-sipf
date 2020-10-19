package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/test/unit/mock"
	"efa-server/usecase"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

var FabricName = "default"

func TestFabricAdd_FabricAlreadyExists(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{Name: FabricName}, nil
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	err := devUC.AddFabric(context.Background(), FabricName)
	assert.Equal(t, "Fabric default already exists", err.Error(), "Should be equal")
}

func TestFabricAdd_FabricCreateFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{}, errors.New("Fabric Already Exists")
		},
		MockCreateFabric: func(Fabric *domain.Fabric) error {
			return errors.New("Fabric Create Failed")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	err := devUC.AddFabric(context.Background(), FabricName)
	assert.Equal(t, "Fabric default create failed", err.Error(), "Should be equal")
}

func TestFabricAdd_FabricPropertyCreateFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{}, errors.New("Fabric Already Exists")
		},
		MockCreateFabric: func(Fabric *domain.Fabric) error {
			return nil
		},
		MockGetFabricProperties: func(FabricID uint) (domain.FabricProperties, error) {
			return domain.FabricProperties{}, errors.New("Fabric Property does not exist")
		},
		MockCreateFabricProperties: func(FabricProperties *domain.FabricProperties) error {
			return errors.New("Fabric Create Failed")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	err := devUC.AddFabric(context.Background(), FabricName)
	assert.Equal(t, "Fabric default update Fabric Property failed", err.Error(), "Should be equal")
}

func TestFabricAdd_FabricPropertyUpdateFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{}, errors.New("Fabric Already Exists")
		},
		MockCreateFabric: func(Fabric *domain.Fabric) error {
			return nil
		},

		MockUpdateFabricProperties: func(FabricProperties *domain.FabricProperties) error {
			return errors.New("Fabric Create Failed")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	err := devUC.AddFabric(context.Background(), FabricName)
	assert.Equal(t, "Fabric default update Fabric Property failed", err.Error(), "Should be equal")
}
