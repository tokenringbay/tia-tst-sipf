package nonclos

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/infra/device/actions"
	"efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"efa-server/test/functional"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

var Host = functional.NetConfFreedomIP
var Model = adapter.FreedomType
var UserName = "admin"
var Password = functional.DeviceAdminPassword

//var Password = "godcapp123"
var client *netconf.NetconfClient
var FabricName = "test_fabric"

var vteploopbackPortnumber = "1"
var arpAgingTimeout = "400"
var macAgingTimeout = "2400"
var macAgingConversationalTimeout = "600"
var macMoveLimit = "300"
var duplicateMacTimer = "10"
var duplicateMacTimerMaxCount = "5"

func init() {
	pathMap := lfshook.PathMap{
		log.InfoLevel:  "/var/log/esa_action_test_info.log",
		log.ErrorLevel: "/var/log/esa_action_test_error.log",
	}

	log.AddHook(lfshook.NewHook(
		pathMap,
		&log.JSONFormatter{},
	))
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.TextFormatter{DisableColors: true})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	//log.SetOutput(os.Stdout)
	log.SetOutput(ioutil.Discard)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func TestMain(m *testing.M) {
	//fmt.Println("starting")
	//Iniitalize the client Once
	client = &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
	client.Login()

	code := m.Run()
	//fmt.Println("Stopping")
	client.Close()
	os.Exit(code)
}

//Initialize the test
func initializeTest() (context.Context, *sync.WaitGroup, chan actions.OperationError, []actions.OperationError) {
	//Setup for calling actions
	ctx := context.Background()
	var fabricGate sync.WaitGroup
	fabricErrors := make(chan actions.OperationError, 1)
	Errors := make([]actions.OperationError, 0)
	fabricGate.Add(1)
	//Verify using NetConf Client
	//client := &netconf.NetconfClient{Host, UserName, Password, nil}
	//client.Login()
	return ctx, &fabricGate, fabricErrors, Errors
}

func initializeTestWithoutNetconfClient(numOfNodes int) (context.Context, *sync.WaitGroup, chan actions.OperationError, []actions.OperationError) {
	ctx := context.Background()
	var fabricGate sync.WaitGroup
	fabricErrors := make(chan actions.OperationError, 2*numOfNodes)
	Errors := make([]actions.OperationError, 0)
	fabricGate.Add(1)
	return ctx, &fabricGate, fabricErrors, Errors
}

//Close the Action channel
func closeActionChannel(fabricGate *sync.WaitGroup, fabricErrors chan actions.OperationError, Errors []actions.OperationError) []actions.OperationError {
	//Closing the channel
	go func() {
		fabricGate.Wait()
		close(fabricErrors)

	}()
	for err := range fabricErrors {
		Errors = append(Errors, err)
	}
	return Errors
}

func initializeFabricTest() (context.Context, *sync.WaitGroup, chan operation.ConfigSwitchResponse, chan actions.OperationError, *netconf.NetconfClient) {
	//Setup for calling actions
	ctx := context.Background()
	var fabricGate sync.WaitGroup
	switchResponseChannel := make(chan operation.ConfigSwitchResponse, 1)
	fabricErrors := make(chan actions.OperationError, 1)
	fabricGate.Add(1)
	//Verify using NetConf Client
	//client := &netconf.NetconfClient{Host, UserName, Password, nil}
	//client.Login()
	return ctx, &fabricGate, switchResponseChannel, fabricErrors, client
}

//Close the Action channel
func closeFetchActionChannel(fabricGate *sync.WaitGroup, fabricErrors chan actions.OperationError,
	switchResponseChannel chan operation.ConfigSwitchResponse) ([]operation.ConfigSwitchResponse, []actions.OperationError) {
	//Closing the channel
	go func() {
		fabricGate.Wait()
		close(fabricErrors)
		close(switchResponseChannel)

	}()
	Errors := make([]actions.OperationError, 0)
	for err := range fabricErrors {
		Errors = append(Errors, err)
	}
	SwitchResponses := make([]operation.ConfigSwitchResponse, 0)
	for switchResponse := range switchResponseChannel {
		SwitchResponses = append(SwitchResponses, switchResponse)
	}
	return SwitchResponses, Errors
}
