package plugin

import (
	"bytes"
	"encoding/json"
	"github.com/battlesnakeio/engine/controller/pb"
	"net/http"
	"time"
)

var HTTPClient = &http.Client{Timeout: 10 * time.Second}

type TickInterceptRequest struct {
	Game      *pb.Game
	LastFrame *pb.GameFrame
	NextFrame *pb.GameFrame
}

type TickInterceptResponse struct {
	ModifiedFrame *pb.GameFrame
}

func UpdateState(url string, game *pb.Game, lastFrame *pb.GameFrame, nextFrame *pb.GameFrame) (*pb.GameFrame, error) {
	req := TickInterceptRequest{LastFrame: lastFrame, NextFrame: nextFrame}
	jsonReq, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}

	var response *TickInterceptResponse
	err = postJSON(url, jsonReq, response)
	if err != nil {
		return nil, err
	}

	return response.ModifiedFrame, nil
}

func postJSON(url string, jsonReq []byte, target interface{}) error {
	resp, err := HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}
