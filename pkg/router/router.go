package router

import (
	"github.com/Insulince/triio-api/pkg/api"
	"github.com/Insulince/triio-api/pkg/configuration"
	"github.com/gorilla/mux"
	"net/http"
)

type Router struct {
	*mux.Router
}

func CreateRouter(config *configuration.Config) (router *Router) {
	router = &Router{
		mux.NewRouter().StrictSlash(true),
	}

	router.HandleFunc("/", api.Home).Methods("GET")

	router.HandleFunc("/health", api.HealthCheck).Methods("GET")

	router.HandleFunc("/register", api.Register(config)).Methods("POST")
	router.HandleFunc("/login", api.Login(config)).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(api.NotFound)

	return router
}
