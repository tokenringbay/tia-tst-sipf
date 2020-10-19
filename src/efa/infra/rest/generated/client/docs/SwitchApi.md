# \SwitchApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetSwitch**](SwitchApi.md#GetSwitch) | **Get** /switch | getSwitch
[**UpdateSwitch**](SwitchApi.md#UpdateSwitch) | **Put** /switch | updateSwitch


# **GetSwitch**
> SwitchdataResponse GetSwitch(ctx, ipAddress)
getSwitch

Get only specified Switch details. The Switch can be identified by IP Address

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **ipAddress** | **string**| IP Address of the Device | 

### Return type

[**SwitchdataResponse**](SwitchdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateSwitch**
> SwitchdataResponse UpdateSwitch(ctx, ipAddress)
updateSwitch

Update the specified device from the fabric.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **ipAddress** | **string**| IP Address of the device to be updated. | 

### Return type

[**SwitchdataResponse**](SwitchdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

