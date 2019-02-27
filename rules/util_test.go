package rules

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupSnakeServer(t *testing.T, move MoveResponse, start StartResponse) string {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(writer http.ResponseWriter, request *http.Request) {
		data, err := json.Marshal(&start)
		require.NoError(t, err)
		_, err = writer.Write(data)
		require.NoError(t, err)
	})
	mux.HandleFunc("/move", func(writer http.ResponseWriter, request *http.Request) {
		data, err := json.Marshal(&move)
		require.NoError(t, err)
		_, err = writer.Write(data)
		require.NoError(t, err)
	})
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		require.NoError(t, err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	go func() {
		err = http.Serve(listener, mux)
		require.NoError(t, err)
	}()
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}
