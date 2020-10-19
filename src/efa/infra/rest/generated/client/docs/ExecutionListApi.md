# \ExecutionListApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ExecutionList**](ExecutionListApi.md#ExecutionList) | **Get** /executions | getExecutionList


# **ExecutionList**
> ExecutionsResponse ExecutionList(ctx, limit, optional)
getExecutionList

Get the list of all the previous executions

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **limit** | **int32**| Limit the number of executions that will be sent in the response. Default is 10 | [default to 10]
 **optional** | **map[string]interface{}** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a map[string]interface{}.

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int32**| Limit the number of executions that will be sent in the response. Default is 10 | [default to 10]
 **status** | **string**| Filter the executions based on the status(failed/succeeded/all) | [default to all]

### Return type

[**ExecutionsResponse**](ExecutionsResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

