package plugin

import (
	"encoding/json"
	"github.com/battlesnakeio/engine/controller/pb"
	"net/http"
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
	err := postJSON(url, jsonReq, response)
	if err != nil {
		return nil, err
	}

	return response.ModifiedFrame, nil
}

func postJSON(url string, jsonReq string, target interface{}) error {
	resp, err := HTTPClient.Post(url, "application/json", bytes.NewBufer(jsonReq))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
