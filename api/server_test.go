package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type MockController struct {
	pb.ControllerClient

	Error          error
	CreateResponse *pb.CreateResponse
	StartResponse  *pb.StartResponse
	StatusResponse *pb.StatusResponse
}

func (mc *MockController) Create(ctx context.Context, req *pb.CreateRequest, opts ...grpc.CallOption) (*pb.CreateResponse, error) {
	return mc.CreateResponse, mc.Error
}

func (mc *MockController) Start(ctx context.Context, req *pb.StartRequest, opts ...grpc.CallOption) (*pb.StartResponse, error) {
	return mc.StartResponse, mc.Error
}

func (mc *MockController) Status(ctx context.Context, req *pb.StatusRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {
	return mc.StatusResponse, mc.Error
}

func createAPIServer() (*Server, *MockController) {
	var client = &MockController{
		CreateResponse: &pb.CreateResponse{},
		StartResponse:  &pb.StartResponse{},
		StatusResponse: &pb.StatusResponse{},
	}
	s := New(":1234", client)
	return s, client
}

func createAPIServerWithError(err error) (*Server, *MockController) {
	var client = &MockController{
		Error:          err,
		CreateResponse: &pb.CreateResponse{},
		StartResponse:  &pb.StartResponse{},
		StatusResponse: &pb.StatusResponse{},
	}
	s := New(":1234", client)
	return s, client
}

func basicCreateTest(t *testing.T, bodyJSON string, expectedStatusCode int, controllerError error) {
	s, _ := createAPIServerWithError(controllerError)

	buf := &bytes.Buffer{}
	buf.WriteString(bodyJSON)
	req, _ := http.NewRequest("POST", "/game/create", buf)
	rr := httptest.NewRecorder()

	s.hs.Handler.ServeHTTP(rr, req)
	require.Equal(t, expectedStatusCode, rr.Code)
}

func TestCreate(t *testing.T) {
	basicCreateTest(t, "{}", http.StatusOK, nil)
}

func TestCreateWithBadBodyJSON(t *testing.T) {
	basicCreateTest(t, "invalid json", http.StatusBadRequest, nil)
}

func TestCreateHandlesErrors(t *testing.T) {
	basicCreateTest(t, "{}", http.StatusInternalServerError, errors.New("fail"))
}

func TestStart(t *testing.T) {
	s, _ := createAPIServer()

	req, _ := http.NewRequest("POST", "/game/start/abc_123", nil)
	rr := httptest.NewRecorder()

	s.hs.Handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func basicStatusTest(t *testing.T, id string, expectedStatusCode int, controllerError error) {
	s, _ := createAPIServerWithError(controllerError)

	req, _ := http.NewRequest("GET", "/game/status/"+id, nil)
	rr := httptest.NewRecorder()

	s.hs.Handler.ServeHTTP(rr, req)
	require.Equal(t, expectedStatusCode, rr.Code)
}

func TestStatus(t *testing.T) {
	basicStatusTest(t, "abc_123", http.StatusOK, nil)
}

func TestStatusHandlesErrors(t *testing.T) {
	basicStatusTest(t, "12345", http.StatusInternalServerError, errors.New("fail"))
}
