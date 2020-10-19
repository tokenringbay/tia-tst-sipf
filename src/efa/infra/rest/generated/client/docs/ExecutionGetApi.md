# \ExecutionGetApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ExecutionGet**](ExecutionGetApi.md#ExecutionGet) | **Get** /execution | getExecutionDetail


# **ExecutionGet**
> DetailedExecutionResponse ExecutionGet(ctx, id)
getExecutionDetail

Get the detailed output of the given execution ID

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **id** | **string**| Detailed output of the given execution ID | 

### Return type

[**DetailedExecutionResponse**](DetailedExecutionResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

