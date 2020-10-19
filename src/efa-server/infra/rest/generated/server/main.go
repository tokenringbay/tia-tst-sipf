/*
 * Simplified IP Fabric
 *
 * This is the spec that defines the API provided by the application to register devices to a fabric, configure fabric parameters, validate all the devices in the fabric and configure switches for IP Fabric with/without overlay
 *
 * API version: 1.0
 * Contact: support@extremenetworks.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package main

import (
	"log"
	"net/http"

	// WARNING!
	// Change this to a fully-qualified import path
	// once you place this file into your project.
	// For example,
	//
	//    sw "github.com/myname/myrepo/go"
	//
	sw "./go"
)

func main() {
	log.Printf("Server started")

	router := sw.NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
