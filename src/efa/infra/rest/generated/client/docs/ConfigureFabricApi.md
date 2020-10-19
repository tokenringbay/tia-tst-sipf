# \ConfigureFabricApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ConfigureFabric**](ConfigureFabricApi.md#ConfigureFabric) | **Post** /configure | configureFabric


# **ConfigureFabric**
> ConfigureFabricResponse ConfigureFabric(ctx, fabricName, optional)
configureFabric

Configure IP Fabric for the specified fabric

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **fabricName** | **string**| Name of the fabric to Configure | 
 **optional** | **map[string]interface{}** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a map[string]interface{}.

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fabricName** | **string**| Name of the fabric to Configure | 
 **force** | **bool**|  | [default to false]
 **persist** | **bool**|  | [default to false]

### Return type

[**ConfigureFabricResponse**](ConfigureFabricResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

