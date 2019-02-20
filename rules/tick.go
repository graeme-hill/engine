package rules

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// GameTick runs the game one tick and updates the state
func GameTick(game *pb.Game, lastFrame *pb.GameFrame) (*pb.GameFrame, error) {
	if lastFrame == nil {
		return nil, fmt.Errorf("rules: invalid state, previous frame is nil")
	}
	nextFrame := &pb.GameFrame{
		Turn:   lastFrame.Turn + 1,
		Snakes: lastFrame.Snakes,
		Food:   lastFrame.Food,
	}
	duration := time.Duration(game.SnakeTimeout) * time.Millisecond
	log.WithFields(log.Fields{
		"GameID":  game.ID,
		"Turn":    nextFrame.Turn,
		"Timeout": duration,
	}).Info("GatherSnakeMoves")
	moves := GatherSnakeMoves(duration, game, lastFrame)

	// we have all the snake moves now
	// 1. update snake coords
	updateSnakes(game, nextFrame, moves)
	// 2. game update
	//    a - turn incr -- done above when the next tick is created
	//    b - reduce health points
	//    c - grow snakes, and update snake health if they ate
	//    d - shrink snakes that didn't eat
	//    e - remove eaten food
	//    f - replace eaten food
	log.WithFields(log.Fields{
		"GameID": game.ID,
		"Turn":   nextFrame.Turn,
	}).Info("reduce snake health")
	for _, s := range nextFrame.AliveSnakes() {
		s.Health = s.Health - 1
	}

	log.WithFields(log.Fields{
		"GameID": game.ID,
		"Turn":   nextFrame.Turn,
	}).Info("handle food")

	foodToRemove := checkForSnakesEating(nextFrame)
	nextFood, err := updateFood(game, lastFrame, foodToRemove)
	if err != nil {
		return nil, err
	}
	nextFrame.Food = nextFood

	// 3. check for death
	// 	  a - starvation
	//    b - wall collision
	//    c - snake collision
	log.WithFields(log.Fields{
		"GameID": game.ID,
		"Turn":   nextFrame.Turn,
	}).Info("check for death")
	deathUpdates := checkForDeath(game.Width, game.Height, nextFrame)
	for _, du := range deathUpdates {
		if du.Snake.Death == nil {
			du.Snake.Death = du.Death
		}
	}

	// If using a tick intercept plugin then let it intercept now
	if game.TickInterceptURL != "" {
		return plugin.UpdateState(game.TickInterceptURL, game, lastFrame, nextFrame)
	}

	return nextFrame, nil
}

func updateFood(game *pb.Game, gameFrame *pb.GameFrame, foodToRemove []*pb.Point) ([]*pb.Point, error) {
	var food []*pb.Point
	// discover what food was not eaten
	for _, foodPos := range gameFrame.Food {
		found := false
		for _, r := range foodToRemove {
			if foodPos.Equal(r) {
				found = true
				break
			}
		}

		if !found {
			food = append(food, foodPos)
		}
	}

	foodToAdd := 0
	if game.MaxTurnsToNextFoodSpawn <= 0 {
		foodToAdd = len(foodToRemove)
	} else {
		if game.TurnsSinceLastFoodSpawn == game.MaxTurnsToNextFoodSpawn {
			foodToAdd = int(math.Ceil(float64(len(gameFrame.AliveSnakes())) / 2.0))
		} else {
			chance := rand.Int31n(1001) // use 101 here so we get 0-100 inclusive
			calculatedChance := calculateFoodSpawnChance(game)
			log.WithFields(log.Fields{
				"GameID":            game.ID,
				"Food Spawn Chance": chance,
				"Turns Since Last":  game.TurnsSinceLastFoodSpawn,
				"Calculate Chance":  calculatedChance,
			}).Info("food spawn chance")
			fmt.Println(len(gameFrame.AliveSnakes()))
			if float64(chance) <= calculatedChance {
				foodToAdd = int(math.Ceil(float64(len(gameFrame.AliveSnakes())) / 2.0))

			}
		}
	}

	if foodToAdd > 0 {
		game.TurnsSinceLastFoodSpawn = 0
		for i := 0; i < foodToAdd; i++ {
			p := getUnoccupiedPoint(game.Width, game.Height, gameFrame.Food, gameFrame.AliveSnakes())
			if p != nil {
				food = append(food, p)
			}
		}
	} else {
		game.TurnsSinceLastFoodSpawn++
	}

	return food, nil
}

func calculateFoodSpawnChance(game *pb.Game) float64 {
	minSpawnChance := float64(0.5)

	ratio := math.Pow(1000/minSpawnChance, 1.0/float64(game.MaxTurnsToNextFoodSpawn-1))
	seqNum := float64(game.TurnsSinceLastFoodSpawn)

	spawnChance := minSpawnChance * ((1 - math.Pow(ratio, seqNum)) / (1 - ratio))
	return spawnChance
}

func getUnoccupiedPoint(width, height int32, food []*pb.Point, snakes []*pb.Snake) *pb.Point {
	openPoints := getUnoccupiedPoints(width, height, food, snakes)
	return pickRandomPoint(openPoints)
}

func getUnoccupiedPointOdd(width, height int32, food []*pb.Point, snakes []*pb.Snake) *pb.Point {
	openPoints := getUnoccupiedPoints(width, height, food, snakes)
	openPoints = filterPoints(openPoints, true)
	return pickRandomPoint(openPoints)
}

func getUnoccupiedPointEven(width, height int32, food []*pb.Point, snakes []*pb.Snake) *pb.Point {
	openPoints := getUnoccupiedPoints(width, height, food, snakes)
	openPoints = filterPoints(openPoints, false)
	return pickRandomPoint(openPoints)
}

func filterPoints(openPoints []*pb.Point, even bool) []*pb.Point {
	filteredPoints := []*pb.Point{}
	mod := int32(0)
	if !even {
		mod = int32(1)
	}
	for i := int32(0); i < int32(len(openPoints)); i++ {
		if (openPoints[i].X+openPoints[i].Y)%2 != mod {
			filteredPoints = append(filteredPoints, openPoints[i])
		}
	}
	return filteredPoints
}

func pickRandomPoint(openPoints []*pb.Point) *pb.Point {
	if len(openPoints) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(openPoints))

	return openPoints[randIndex]
}

func getUnoccupiedPoints(width, height int32, food []*pb.Point, snakes []*pb.Snake) []*pb.Point {
	occupiedPoints := getUniqOccupiedPoints(food, snakes)

	numCandidatePoints := (width * height) - int32(len(occupiedPoints))
	if numCandidatePoints <= 0 {
		return []*pb.Point{}
	}

	candidatePoints := make([]*pb.Point, 0, numCandidatePoints)

	for x := int32(0); x < width; x++ {
		for y := int32(0); y < height; y++ {
			p := &pb.Point{X: x, Y: y}
			match := false

			for _, o := range occupiedPoints {
				if o.Equal(p) {
					match = true
					break
				}
			}

			if !match {
				candidatePoints = append(candidatePoints, p)
			}
		}
	}

	return candidatePoints
}

func getUniqOccupiedPoints(food []*pb.Point, snakes []*pb.Snake) []*pb.Point {
	occupiedPoints := []*pb.Point{}
	for _, f := range food {
		candidate := true
		for _, o := range occupiedPoints {
			if o.Equal(f) {
				candidate = false
				break
			}
		}
		if candidate {
			occupiedPoints = append(occupiedPoints, f)
		}
	}

	for _, s := range snakes {
		for _, b := range s.Body {
			candidate := true
			for _, o := range occupiedPoints {
				if o.Equal(b) {
					candidate = false
					break
				}
			}
			if candidate {
				occupiedPoints = append(occupiedPoints, b)
			}
		}
	}

	return occupiedPoints
}

func updateSnakes(game *pb.Game, frame *pb.GameFrame, moves []*SnakeUpdate) {
	for _, update := range moves {
		if update.Err != nil {
			log.WithFields(log.Fields{
				"GameID":  game.ID,
				"Error":   update.Err,
				"SnakeID": update.Snake.ID,
				"Name":    update.Snake.Name,
				"Turn":    frame.Turn,
			}).Info("Default move")
			update.Snake.DefaultMove()
		} else {
			log.WithFields(log.Fields{
				"GameID":  game.ID,
				"SnakeID": update.Snake.ID,
				"Name":    update.Snake.Name,
				"Turn":    frame.Turn,
				"Move":    update.Move,
			}).Info("Move")
			update.Snake.Move(update.Move)
		}
	}
}

func checkForSnakesEating(frame *pb.GameFrame) []*pb.Point {
	foodToRemove := []*pb.Point{}
	for _, snake := range frame.AliveSnakes() {
		ate := false
		for _, foodPos := range frame.Food {
			if snake.Head().Equal(foodPos) {
				snake.Health = 100
				ate = true
				foodToRemove = append(foodToRemove, foodPos)
				log.WithFields(log.Fields{
					"SnakeID": snake.ID,
					"Name":    snake.Name,
					"Turn":    frame.Turn,
					"Food":    foodPos,
				}).Info("snake ate")
			}
		}

		// It's possible here to have 2 points at the tail, and so if we remove the second one, it
		// looks like the snake hasn't moved.
		// This shouldn't happen, but just in case
		if len(snake.Body) != 0 {
			snake.Body = snake.Body[:len(snake.Body)-1]
		}
		if ate {
			tail := snake.Tail()
			snake.Body = append(snake.Body, &pb.Point{X: tail.X, Y: tail.Y})
		}
	}
	return foodToRemove
}
