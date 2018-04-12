package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Server this is the api server
type Server struct {
	hs *http.Server
}

// ClientHandle is a function that handles an http route and accepts a ControllerClient
// in addition to the normal httprouter.Handle parameters.
type ClientHandle func(http.ResponseWriter, *http.Request, httprouter.Params, pb.ControllerClient)

// New creates a new api server
func New(addr string, c pb.ControllerClient) *Server {
	router := httprouter.New()
	router.POST("/game/create", clientHandle(c, createGame))
	router.POST("/game/start/:id", clientHandle(c, startGame))
	router.GET("/game/status/:id", clientHandle(c, getStatus))

	return &Server{
		hs: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func clientHandle(c pb.ControllerClient, innerHandle ClientHandle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		innerHandle(w, r, p, c)
	}
}

func createGame(w http.ResponseWriter, r *http.Request, _ httprouter.Params, c pb.ControllerClient) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("Unable to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	req := &pb.CreateRequest{}

	err = json.Unmarshal(body, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON: " + err.Error()))
		return
	}

	// TODO: use a context with timeout
	resp, err := c.Create(context.Background(), req)
	if err != nil {
		log.WithError(err).Error("Error creating game")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	j, err := json.Marshal(resp)
	if err != nil {
		log.WithError(err).WithField("resp", resp).Error("Error serializing to JSON")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(j)
}

func startGame(w http.ResponseWriter, r *http.Request, ps httprouter.Params, c pb.ControllerClient) {
	id := ps.ByName("id")
	req := &pb.StartRequest{
		ID: id,
	}
	// TODO: use a context with timeout
	_, err := c.Start(context.Background(), req)
	if err != nil {
		log.WithError(err).WithField("req", req).Error("Error while calling controller start")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params, c pb.ControllerClient) {
	id := ps.ByName("id")
	req := &pb.StatusRequest{
		ID: id,
	}
	// TODO: use a context with timeout
	resp, err := c.Status(context.Background(), req)
	if err != nil {
		log.WithError(err).WithField("req", req).Error("Error while calling controller status")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	j, err := json.Marshal(resp)
	if err != nil {
		log.WithError(err).WithField("resp", resp).Error("Error serializing response to JSON")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(j)
}

// WaitForExit starts up the server and blocks until the server shuts down.
func (s *Server) WaitForExit() {
	log.Infof("Battlesnake engine api listening on %s", s.hs.Addr)
	err := s.hs.ListenAndServe()
	if err != nil {
		log.Errorf("Error while listening: %v", err)
	}
}
