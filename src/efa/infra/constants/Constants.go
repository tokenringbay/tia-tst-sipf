package constants

import "time"

//Constants used by the Client Application
const (
	ApplicationName   = "efa"
	DefaultTimeFormat = time.RFC3339
	DefaultFabric     = "default"
	SSArchive         = "/var/" + ApplicationName + "/" + ApplicationName + "_SS.zip"
	AppInfoLocation   = "/var/" + ApplicationName + "/" + ApplicationName + "_appinfo.txt"
	LogPathToArchove  = "/var/log/" + ApplicationName + "/"
	DBLocation        = "/var/" + ApplicationName + "/" + ApplicationName + ".db"
	// TODO We might have to include the build number and version string here, instead of fetching from the server
)
