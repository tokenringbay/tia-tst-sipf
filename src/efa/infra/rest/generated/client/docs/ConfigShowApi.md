# \ConfigShowApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ConfigShow**](ConfigShowApi.md#ConfigShow) | **Get** /config | getConfigShow


# **ConfigShow**
> ConfigShowResponse ConfigShow(ctx, fabricName, role)
getConfigShow

Get the Configuration of Devices in a Fabric

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **fabricName** | **string**| Name of the fabric to retrieve | 
  **role** | **string**| role of the devices for which config needs to be fetched | [default to all]

### Return type

[**ConfigShowResponse**](ConfigShowResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

