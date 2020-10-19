package openapi

import (
	ohandler "efa-server/infra/rest/openapi/handler"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

//Route defines a unique route for a REST request
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	QueryPairs  []string
}

//Routes define an array of routes supported by the application
type Routes []Route

//NewRouter returns a new Router which routes the REST request to the unique Handler
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = ohandler.Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler).
			Queries(route.QueryPairs...)

	}

	return router
}

//Index returns a welcome message
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

var routes = Routes{
	Route{
		Name:        "Index",
		Method:      "GET",
		Pattern:     "/v1/",
		HandlerFunc: Index,
	},

	Route{
		Name:        "ExecutionGet",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/v1/execution",
		HandlerFunc: ohandler.ExecutionGetHandler,
		QueryPairs:  []string{"id", "{id}"},
	},

	Route{
		Name:        "ExecutionList",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/v1/executions",
		HandlerFunc: ohandler.ExecutionListHandler,
		QueryPairs:  []string{"limit", "{limit}", "status", "{status}"},
	},
	Route{
		Name:        "UpdateSwitches",
		Method:      strings.ToUpper("Put"),
		Pattern:     "/v1/switches",
		HandlerFunc: ohandler.UpdateSwitches,
	},
	Route{
		Name:        "UpdateSwitches",
		Method:      strings.ToUpper("Put"),
		Pattern:     "/v1/switches",
		HandlerFunc: ohandler.UpdateSwitches,
	},
	Route{
		Name:        "CreateSwitches",
		Method:      strings.ToUpper("Post"),
		Pattern:     "/v1/switches",
		HandlerFunc: ohandler.CreateSwitches,
	},
	Route{
		Name:        "ValidateFabric",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/v1/validate",
		HandlerFunc: ohandler.ValidateFabric,
		QueryPairs:  []string{"fabric_name", "{fabric_name}"},
	},
	Route{
		Name:        "ConfigureFabric",
		Method:      strings.ToUpper("Post"),
		Pattern:     "/v1/configure",
		HandlerFunc: ohandler.ConfigureFabric,
		QueryPairs:  []string{"fabric_name", "{fabric_name}", "persist", "{persist}", "force", "{force}"},
	},

	Route{
		Name:        "ConfigShow",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/v1/config",
		HandlerFunc: ohandler.FabricConfigShow,
		QueryPairs:  []string{"fabricName", "{fabricName}", "role", "{role}"},
	},

	Route{
		Name:        "DebugClear",
		Method:      strings.ToUpper("Post"),
		Pattern:     "/v1/debug/clear",
		HandlerFunc: ohandler.DebugClear,
	},
	Route{
		Name:        "DeleteSwitches",
		Method:      strings.ToUpper("Delete"),
		Pattern:     "/v1/switches",
		HandlerFunc: ohandler.DeleteSwitches,
	},
	Route{
		Name:        "getSwitches",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/v1/switches",
		HandlerFunc: ohandler.ShowDevicesInFabric,
		QueryPairs:  []string{"name", "{name}"},
	},
	Route{
		Name:        "updateFabric",
		Method:      strings.ToUpper("Put"),
		Pattern:     "/v1/fabric",
		HandlerFunc: ohandler.UpdateFabricSettings,
	},
	Route{
		Name:        "getFabric",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/v1/fabric",
		HandlerFunc: ohandler.ShowFabricSettings,
		QueryPairs:  []string{"name", "{name}"},
	},
}
