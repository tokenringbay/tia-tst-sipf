# \SwitchesApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateSwitches**](SwitchesApi.md#CreateSwitches) | **Post** /switches | Add new Devices to the specified Fabric
[**DeleteSwitches**](SwitchesApi.md#DeleteSwitches) | **Delete** /switches | deleteSwitches
[**GetSwitches**](SwitchesApi.md#GetSwitches) | **Get** /switches | getSwitches
[**UpdateSwitches**](SwitchesApi.md#UpdateSwitches) | **Put** /switches | Update One or more switch details.


# **CreateSwitches**
> SwitchesdataResponse CreateSwitches(ctx, optional)
Add new Devices to the specified Fabric

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
 **optional** | **map[string]interface{}** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a map[string]interface{}.

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **switches** | [**NewSwitches**](NewSwitches.md)| Add one or more switches to the fabric. | 

### Return type

[**SwitchesdataResponse**](SwitchesdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteSwitches**
> SwitchesdataResponse DeleteSwitches(ctx, optional)
deleteSwitches

Delete the specified devices from the fabric.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
 **optional** | **map[string]interface{}** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a map[string]interface{}.

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **switches** | [**DeleteSwitchesRequest**](DeleteSwitchesRequest.md)| IP Addresses of the device to be deleted. | 

### Return type

[**SwitchesdataResponse**](SwitchesdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetSwitches**
> SwitchesdataResponse GetSwitches(ctx, name)
getSwitches

Get All switches in the specified fabric.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **name** | **string**| Name of the fabric to retrieve switches for | 

### Return type

[**SwitchesdataResponse**](SwitchesdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateSwitches**
> SwitchesUpdateResponse UpdateSwitches(ctx, optional)
Update One or more switch details.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
 **optional** | **map[string]interface{}** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a map[string]interface{}.

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **switches** | [**UpdateSwitchParameters**](UpdateSwitchParameters.md)| Update One or more switch details | 

### Return type

[**SwitchesUpdateResponse**](SwitchesUpdateResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

