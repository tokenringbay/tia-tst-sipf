package main

import (
	"efa-server/infra/constants"
	"efa-server/infra/rest"
	"fmt"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"sync"
	"time"
)

func init() {

	LogWriter := lumberjack.Logger{
		Filename:   constants.LogLocation,
		MaxSize:    1, // megabytes
		MaxBackups: constants.NumberOfLogFiles - 1,
	}
	writerMap := lfshook.WriterMap{
		log.InfoLevel:  &LogWriter,
		log.ErrorLevel: &LogWriter,
	}

	log.AddHook(lfshook.NewHook(
		writerMap,
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

	startRestServer()
}

func startRestServer() error {
	start := time.Now()
	var wg sync.WaitGroup
	//Wait Groups to start both the Rest Servers
	openAPIServer := rest.NewOpenAPIServer(&wg)
	go openAPIServer.RunOpenAPIServer()
	wg.Add(1)

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Took ", elapsed)
	return nil
}

func main() {

}
