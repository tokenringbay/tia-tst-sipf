package utils

import (
	"fmt"
	"net"
	"net/url"
)

//IsServerConnectionError identifies if its a server connection error
func IsServerConnectionError(error error) bool {
	urlError, _ := error.(*url.Error)
	if urlError != nil {
		op, _ := urlError.Err.(*net.OpError)
		if op != nil {
			if op.Op == "dial" {
				fmt.Println("Error: Server unreachable. Please make sure the server is operational and reachable.")
				return true
			}
		}
	}
	return false
}
