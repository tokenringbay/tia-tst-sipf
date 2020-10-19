package constants

import (
	"sync"
	"time"
)

//Constants used by the Application
const (
	ApplicationName       = "efa"
	ApplicationServerName = "efa-server"
	DBLocation            = "/var/" + ApplicationName + "/" + ApplicationName + ".db"
	TESTDBLocation        = "/var/" + ApplicationName + "/" + ApplicationName + "_test.db"
	//InfoLogLocation   = "/var/log/" + ApplicationName + "_info.log"
	//ErrorLogLocation  = "/var/log/" + ApplicationName + "_error.log"
	LogLocation       = "/var/log/" + ApplicationName + "/" + ApplicationName + ".log"
	DefaultFabric     = "default"
	DefaultTimeFormat = time.RFC3339

	NumberOfLogFiles = 10
	LLDPSleep        = 10
	RackNameSuffix   = "Rack-"
)

//AESEncryptionKey  test
var AESEncryptionKey = []byte("Why is it that always you three?")

//RestLock locks all Rest API except execution list when API start
var RestLock sync.Mutex
