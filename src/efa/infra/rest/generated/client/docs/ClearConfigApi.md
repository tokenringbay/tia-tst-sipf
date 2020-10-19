# \ClearConfigApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ClearConfig**](ClearConfigApi.md#ClearConfig) | **Post** /debug/clear | Clear Config


# **ClearConfig**
> DebugClearResponse ClearConfig(ctx, optional)
Clear Config

Clear Configs from a collection of switches

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
 **optional** | **map[string]interface{}** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a map[string]interface{}.

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **switches** | [**DebugClearRequest**](DebugClearRequest.md)| One or More switches where configs are to be cleared. | 

### Return type

[**DebugClearResponse**](DebugClearResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

