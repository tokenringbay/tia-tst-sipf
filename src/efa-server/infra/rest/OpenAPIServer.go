package rest

import (
	"context"
	"net/http"

	"efa-server/infra/rest/openapi"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"efa-server/gateway/appcontext"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/database"
	"github.com/google/uuid"
	"sync"
)

//OpenAPIServer Receiver object for OpenAPI Rest Server
type OpenAPIServer struct {
	http.Server
	shutdownReq chan bool
	wg          *sync.WaitGroup
	reqCount    uint32
}

func (s *OpenAPIServer) waitShutdown() {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	//Wait interrupt or shutdown request through /shutdown
	select {
	case sig := <-irqSig:
		log.Printf("Shutdown request (signal: %v)", sig)
	case sig := <-s.shutdownReq:
		log.Printf("Shutdown request (/shutdown %v)", sig)
	}

	log.Printf("Stoping http server ...")

	//Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	//shutdown the server
	err := s.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown request error: %v", err)
	}
}

//OpenAPIShutdownHandler provides shutdown handler for closing the REST Server
//It also closes the Database.
func (s *OpenAPIServer) OpenAPIShutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Shutdown server"))

	//TODO Remove this when Daemonized
	//First shut down the DB
	if err := database.GetWorkingInstance().Close(); err != nil {
		log.Errorln("Failed to shutdown the database engine", err)

	}
	//Do nothing if shutdown request already issued
	//if s.reqCount == 0 then set to 1, return true otherwise false
	if !atomic.CompareAndSwapUint32(&s.reqCount, 0, 1) {
		log.Printf("Shutdown through API call in progress...")
		return
	}

	go func() {
		s.shutdownReq <- true
	}()

	s.wg.Done()
}

//NewOpenAPIServer provides an instance of OpenAPIServer
func NewOpenAPIServer(wg *sync.WaitGroup) *OpenAPIServer {
	//create server
	s := &OpenAPIServer{
		Server: http.Server{
			Addr:         ":8081",
			ReadTimeout:  1000 * time.Second,
			WriteTimeout: 1000 * time.Second,
		},
		wg:          wg,
		shutdownReq: make(chan bool),
	}

	router := openapi.NewRouter()

	//set http server handler
	s.Handler = router
	router.HandleFunc("/shutdown", s.OpenAPIShutdownHandler)

	return s
}

//RunOpenAPIServer starts the Database and Rest Server
func (s *OpenAPIServer) RunOpenAPIServer() {
	//Start the server
	server := s
	//TODO move Database initialization out to the main When Server and Client are seperated
	//Initialize the DB
	database.Setup(constants.DBLocation)
	//Add Default Fabric from here
	rqID := uuid.New().String()
	_, ctx := appcontext.LoggerAndContext(rqID)
	infra.GetUseCaseInteractor().AddFabric(ctx, constants.DefaultFabric)
	infra.GetUseCaseInteractor().DatabaseUpgrade(ctx, constants.DefaultFabric)

	done := make(chan bool)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("Listen and serve: %v", err)
		}
		done <- true
	}()

	//wait shutdown
	server.waitShutdown()

	<-done
	log.Printf("DONE!")
}
