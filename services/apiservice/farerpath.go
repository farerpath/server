package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/farerpath/server/services/apiservice/server"

	"github.com/farerpath/server/services/apiservice/route"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	"google.golang.org/grpc"
)

const (
	AUTHSERVICE = "authservice-service:17080"
	ALBUMSERVICE = "albumservice-service:17080"
	SESSIONSERVICE = "sessionservice-service:17080"
)

// App function
// Main handler group
func App() http.Handler {
	r := mux.NewRouter()

	// API Version 1
	apiv1 := r.PathPrefix("/v1").Subrouter()

	// Ping for test
	// Response: "{Request Method} : Pong"
	// if this not respond, it means server's death :(
	//apiv1.HandleFunc("/Ping", route.Ping)
	apiv1.HandleFunc("/Ping", route.Ping)

	// Login, Logout, Register
	apiv1.HandleFunc("/Login", route.Login)
	apiv1.HandleFunc("/Logout", route.Logout)
	apiv1.HandleFunc("/Register", route.Register)

	// User information
	// GET - Response: Account model in json string
	apiv1.HandleFunc("/User/{userId}", route.UserHandler).Methods(http.MethodGet, http.MethodPatch)
	apiv1.HandleFunc("/User/{userId}/password", route.UserPasswordHandler).Methods(http.MethodPatch)

	// Picture
	// GET - Response: Get Picture list of userId's
	// POST - Upload Picture
	apiv1.HandleFunc("/Pictures/{userId}", route.PictureListHandler).Methods(http.MethodPost, http.MethodGet)

	// Download(GET), Delete(DELETE)...
	apiv1.HandleFunc("/Pictures/{userId}/{pictureId}", route.PictureHandler).Methods(http.MethodGet, http.MethodDelete)

	apiv1.HandleFunc("/Pictures/{userId}/file", route.PictureFileHandler)

	// GET - Response: Album lists of userId
	// POST - Response: Make album
	apiv1.HandleFunc("/Album/{userId}", route.AlbumHandler).Methods(http.MethodGet, http.MethodPost)

	// Album of user
	// Delete(DELETE), Update(PATCH), Get(GET)
	apiv1.HandleFunc("/Album/{userId}/{albumId}", route.UserAlbumHandler).Methods(http.MethodDelete, http.MethodPatch, http.MethodGet)

	// Add, Delete picture to album
	apiv1.HandleFunc("/Album/{userId}/{albumId}/{pictureId}", route.AlbumPictureHandler)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "HEAD"},
		AllowedHeaders:   []string{"X-Farerpath-Token", "Content-Type"},
		AllowCredentials: true,
	})

	n := negroni.New(negroni.NewLogger(), negroni.NewRecovery(), c)

	n.UseHandler(r)

	return n
}

func main() {
	initGrpcConn()
	fmt.Println("Server starts at 0.0.0.0:80")

	http.ListenAndServe(":80", App())
}

func initGrpcConn() {
	conn1, err := grpc.Dial(AUTHSERVICE, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect grpc AUTHSERVICE 1 %v", err)
	}

	conn2, err := grpc.Dial(ALBUMSERVICE, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect grpc ALBUMSERVICE 1 %v", err)
	}

	conn3, err := grpc.Dial(SESSIONSERVICE, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect grpc SESSIONSERVICE 1 %v", err)
	}

	server.InitGrpcConn(conn1, conn2, conn3)
	route.InitGrpcConn(conn3)
}
