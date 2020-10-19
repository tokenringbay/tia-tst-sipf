# \FabricValidationApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ValidateFabric**](FabricValidationApi.md#ValidateFabric) | **Get** /validate | validateFabric


# **ValidateFabric**
> FabricValidateResponse ValidateFabric(ctx, fabricName)
validateFabric

Validate Fabric settings, cabling between switches and potentially configurations on switches for IP Fabric formation

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **fabricName** | **string**| Name of the fabric to validate | 

### Return type

[**FabricValidateResponse**](FabricValidateResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

