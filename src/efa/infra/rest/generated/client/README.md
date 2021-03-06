# Go API client for swagger

This is the spec that defines the API provided by the application to register devices to a fabric, configure fabric parameters, validate all the devices in the fabric and configure switches for IP Fabric with/without overlay

## Overview
This API client was generated by the [swagger-codegen](https://github.com/swagger-api/swagger-codegen) project.  By using the [swagger-spec](https://github.com/swagger-api/swagger-spec) from a remote server, you can easily generate an API client.

- API version: 1.0
- Package version: 1.0.0
- Build package: io.swagger.codegen.languages.GoClientCodegen
For more information, please visit [http://www.extremenetworks.com](http://www.extremenetworks.com)

## Installation
Put the package under your project folder and add the following in import:
```
    "./swagger"
```

## Documentation for API Endpoints

All URIs are relative to *http://localhost:8081/v1*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*ClearConfigApi* | [**ClearConfig**](docs/ClearConfigApi.md#clearconfig) | **Post** /debug/clear | Clear Config
*ConfigShowApi* | [**ConfigShow**](docs/ConfigShowApi.md#configshow) | **Get** /config | getConfigShow
*ConfigureFabricApi* | [**ConfigureFabric**](docs/ConfigureFabricApi.md#configurefabric) | **Post** /configure | configureFabric
*ExecutionGetApi* | [**ExecutionGet**](docs/ExecutionGetApi.md#executionget) | **Get** /execution | getExecutionDetail
*ExecutionListApi* | [**ExecutionList**](docs/ExecutionListApi.md#executionlist) | **Get** /executions | getExecutionList
*FabricApi* | [**CreateFabric**](docs/FabricApi.md#createfabric) | **Post** /fabric | Create a Fabric
*FabricApi* | [**DeleteFabric**](docs/FabricApi.md#deletefabric) | **Delete** /fabric | deleteFabric
*FabricApi* | [**GetFabric**](docs/FabricApi.md#getfabric) | **Get** /fabric | getFabric
*FabricApi* | [**GetFabrics**](docs/FabricApi.md#getfabrics) | **Get** /fabrics | getFabrics
*FabricApi* | [**UpdateFabric**](docs/FabricApi.md#updatefabric) | **Put** /fabric | Update a Fabric settings
*FabricValidationApi* | [**ValidateFabric**](docs/FabricValidationApi.md#validatefabric) | **Get** /validate | validateFabric
*SupportSaveApi* | [**SupportSave**](docs/SupportSaveApi.md#supportsave) | **Get** /support | getSupport
*SwitchApi* | [**GetSwitch**](docs/SwitchApi.md#getswitch) | **Get** /switch | getSwitch
*SwitchApi* | [**UpdateSwitch**](docs/SwitchApi.md#updateswitch) | **Put** /switch | updateSwitch
*SwitchesApi* | [**CreateSwitches**](docs/SwitchesApi.md#createswitches) | **Post** /switches | Add new Devices to the specified Fabric
*SwitchesApi* | [**DeleteSwitches**](docs/SwitchesApi.md#deleteswitches) | **Delete** /switches | deleteSwitches
*SwitchesApi* | [**GetSwitches**](docs/SwitchesApi.md#getswitches) | **Get** /switches | getSwitches
*SwitchesApi* | [**UpdateSwitches**](docs/SwitchesApi.md#updateswitches) | **Put** /switches | Update One or more switch details.


## Documentation For Models

 - [ConfigShowResponse](docs/ConfigShowResponse.md)
 - [ConfigureFabricResponse](docs/ConfigureFabricResponse.md)
 - [DebugClearRequest](docs/DebugClearRequest.md)
 - [DebugClearResponse](docs/DebugClearResponse.md)
 - [DeleteSwitchesRequest](docs/DeleteSwitchesRequest.md)
 - [DetailedExecutionResponse](docs/DetailedExecutionResponse.md)
 - [DeviceStatusModel](docs/DeviceStatusModel.md)
 - [ErrorModel](docs/ErrorModel.md)
 - [ExecutionResponse](docs/ExecutionResponse.md)
 - [ExecutionsResponse](docs/ExecutionsResponse.md)
 - [FabricParameter](docs/FabricParameter.md)
 - [FabricSettings](docs/FabricSettings.md)
 - [FabricValidateResponse](docs/FabricValidateResponse.md)
 - [FabricdataErrorResponse](docs/FabricdataErrorResponse.md)
 - [FabricdataResponse](docs/FabricdataResponse.md)
 - [FabricsdataErrorResponse](docs/FabricsdataErrorResponse.md)
 - [FabricsdataResponse](docs/FabricsdataResponse.md)
 - [NewFabric](docs/NewFabric.md)
 - [NewSwitches](docs/NewSwitches.md)
 - [Rack](docs/Rack.md)
 - [SupportsaveResponse](docs/SupportsaveResponse.md)
 - [SwitchUpdateResponse](docs/SwitchUpdateResponse.md)
 - [SwitchdataResponse](docs/SwitchdataResponse.md)
 - [SwitchdataResponseFabric](docs/SwitchdataResponseFabric.md)
 - [SwitchesUpdateResponse](docs/SwitchesUpdateResponse.md)
 - [SwitchesdataResponse](docs/SwitchesdataResponse.md)
 - [UpdateSwitchParameters](docs/UpdateSwitchParameters.md)
 - [ExtendedErrorModel](docs/ExtendedErrorModel.md)


## Documentation For Authorization
 Endpoints do not require authorization.


## Author

support@extremenetworks.com

