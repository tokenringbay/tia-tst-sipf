# \FabricApi

All URIs are relative to *http://localhost:8081/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateFabric**](FabricApi.md#CreateFabric) | **Post** /fabric | Create a Fabric
[**DeleteFabric**](FabricApi.md#DeleteFabric) | **Delete** /fabric | deleteFabric
[**GetFabric**](FabricApi.md#GetFabric) | **Get** /fabric | getFabric
[**GetFabrics**](FabricApi.md#GetFabrics) | **Get** /fabrics | getFabrics
[**UpdateFabric**](FabricApi.md#UpdateFabric) | **Put** /fabric | Update a Fabric settings


# **CreateFabric**
> FabricdataResponse CreateFabric(ctx, optional)
Create a Fabric

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
 **optional** | **map[string]interface{}** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a map[string]interface{}.

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fabric** | [**NewFabric**](NewFabric.md)| Add a new Fabric. | 

### Return type

[**FabricdataResponse**](FabricdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteFabric**
> FabricdataResponse DeleteFabric(ctx, name)
deleteFabric

Delete the specified fabric.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **name** | **string**| Name of the fabric to be deleted | 

### Return type

[**FabricdataResponse**](FabricdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetFabric**
> FabricdataResponse GetFabric(ctx, name)
getFabric

Get only specified fabric details. The fabric can be identified by id or name

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
  **name** | **string**| Name of the fabric to retrieve | 

### Return type

[**FabricdataResponse**](FabricdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetFabrics**
> FabricsdataResponse GetFabrics(ctx, )
getFabrics

Get All fabric details configured in the application

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**FabricsdataResponse**](FabricsdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateFabric**
> FabricdataResponse UpdateFabric(ctx, optional)
Update a Fabric settings

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for logging, tracing, authentication, etc.
 **optional** | **map[string]interface{}** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a map[string]interface{}.

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fabricSettings** | [**FabricSettings**](FabricSettings.md)| Update Fabric Settings. | 

### Return type

[**FabricdataResponse**](FabricdataResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

