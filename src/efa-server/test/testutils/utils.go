package testutils

import (
	"sync"
)

//TestMutex Used to Lock Between integration and apiintegration switch access
var TestMutex sync.Mutex
