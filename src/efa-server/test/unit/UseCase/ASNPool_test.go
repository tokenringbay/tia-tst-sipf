package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/test/unit/mock"
	"efa-server/usecase"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestASNPool_PopulateASN(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{Name: FabricName, ID: 1}, nil
		},
	}
	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	err := devUC.PopulateASN(context.Background(), FabricName, 100, 1000, 1, usecase.SpineRole)
	assert.Nil(t, err, "")
}
func TestASNPool_FailedPopulateASN(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{Name: FabricName, ID: 1}, nil
		},
		MockCreateASN: func(ASN *domain.ASNAllocationPool) error {
			return errors.New("Unable to fetch ASN")
		},
	}
	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	err := devUC.PopulateASN(context.Background(), FabricName, 100, 1000, 1, usecase.SpineRole)
	assert.EqualError(t, err, "ASN Pool Initialization Failed for Role Spine", "")
}

func TestASNPool_GetASN(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{Name: FabricName, ID: 1}, nil
		},
		MockGetNextASNForRole: func(FabricID uint, role string) (domain.ASNAllocationPool, error) {
			return domain.ASNAllocationPool{ASN: uint64(120)}, nil
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	asn, err := devUC.GetASN(context.Background(), 1, 1, usecase.SpineRole)
	assert.Equal(t, uint64(120), asn, "asn")
	fmt.Println(asn)
	assert.Nil(t, err, "")
}
func TestASNPool_GetASNExhausted(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{Name: FabricName, ID: 1}, nil
		},
		MockGetNextASNForRole: func(FabricID uint, role string) (domain.ASNAllocationPool, error) {
			return domain.ASNAllocationPool{}, errors.New("Exhausted")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	asn, err := devUC.GetASN(context.Background(), 1, 1, usecase.SpineRole)
	assert.Equal(t, uint64(0), asn, "asn")
	assert.EqualError(t, err, "ASN Exhausted for 1", "")
}
