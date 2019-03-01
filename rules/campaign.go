package rules

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/battlesnakeio/engine/controller/pb"
)

var currentID = 0

func cellIsInVacant2x2(p1 *pb.Point, f *pb.GameFrame, g *pb.Game) bool {
	p2 := &pb.Point{X: p1.X + 1, Y: p1.Y}
	p3 := &pb.Point{X: p1.X, Y: p1.Y + 1}
	p4 := &pb.Point{X: p1.X + 1, Y: p1.Y + 1}
	return cellIsVacant(p1, f, g) && cellIsVacant(p2, f, g) && cellIsVacant(p3, f, g) && cellIsVacant(p4, f, g)
}

func cellIsInVacant3x3(p *pb.Point, f *pb.GameFrame, g *pb.Game) bool {
	if !cellIsInVacant2x2(p, f, g) {
		return false
	}

	p1 := &pb.Point{X: p.X + 2, Y: p.Y}
	p2 := &pb.Point{X: p.X + 2, Y: p.Y + 1}
	p3 := &pb.Point{X: p.X + 2, Y: p.Y + 2}
	p4 := &pb.Point{X: p.X + 1, Y: p.Y + 2}
	p5 := &pb.Point{X: p.X, Y: p.Y + 2}

	return cellIsVacant(p1, f, g) && cellIsVacant(p2, f, g) && cellIsVacant(p3, f, g) && cellIsVacant(p4, f, g) && cellIsVacant(p5, f, g)
}

func allGreatCells(f *pb.GameFrame, g *pb.Game) []*pb.Point {
	result := []*pb.Point{}
	for x := int32(0); x < g.Width; x += 3 {
		for y := int32(0); y < g.Height; y += 3 {
			p := &pb.Point{X: x, Y: y}
			if cellIsInVacant3x3(p, f, g) {
				result = append(result, p)
			}
		}
	}
	return result
}

func allGoodCells(f *pb.GameFrame, g *pb.Game) []*pb.Point {
	result := []*pb.Point{}
	for x := int32(0); x < g.Width; x += 2 {
		for y := int32(0); y < g.Height; y += 2 {
			p := &pb.Point{X: x, Y: y}
			if cellIsInVacant2x2(p, f, g) {
				result = append(result, p)
			}
		}
	}
	return result
}

func allVacantCells(f *pb.GameFrame, g *pb.Game) []*pb.Point {
	result := []*pb.Point{}
	for x := int32(0); x < g.Width; x++ {
		for y := int32(0); y < g.Height; y++ {
			p := &pb.Point{X: x, Y: y}
			if cellIsVacant(p, f, g) {
				result = append(result, p)
			}
		}
	}
	return result
}

func nextID() string {
	id := currentID
	currentID++
	return strconv.Itoa(id)
}

func getSnakeCount(level int) int {
	//return int(level / 10)
	return 1
}

func makeSnake(pos *pb.Point, level int) *pb.Snake {
	name := fmt.Sprintf("~~ Level %d ~~", level)
	return &pb.Snake{
		ID:     nextID(),
		Name:   name,
		Body:   []*pb.Point{pos, pos, pos},
		URL:    "http://localhost:5000",
		Health: 100,
		Color:  "#ff00bb",
	}
}

func getCurrentLevel(f *pb.GameFrame) int {
	for _, s := range f.Snakes {
		if strings.HasPrefix(s.Name, "~~ Level ") {
			return getLevel(s.Name)
		}
	}
	return 0
}

func getLevel(name string) int {
	parts := strings.Split(name, " ")
	level, _ := strconv.Atoi(parts[2])
	return level
}

func spawnAt(pos *pb.Point, level int, f *pb.GameFrame) {
	snake := makeSnake(pos, level)
	f.Snakes = append(f.Snakes, snake)
}

func oob(p *pb.Point, g *pb.Game) bool {
	return p.X >= g.Width || p.Y >= g.Height
}

func cellIsVacant(p1 *pb.Point, f *pb.GameFrame, g *pb.Game) bool {
	if oob(p1, g) {
		return false
	}

	for _, s := range f.AliveSnakes() {
		for _, p2 := range s.Body {
			if p1.X == p2.X && p1.Y == p2.Y {
				return false
			}
		}
	}
	return true
}

func findSpawnPoints(f *pb.GameFrame, g *pb.Game, n int) []*pb.Point {
	result := []*pb.Point{}
	great := allGreatCells(f, g)
	result = append(result, randomPoints(great, n)...)
	if len(result) >= n {
		return result
	}

	good := allGoodCells(f, g)
	result = append(result, randomPoints(good, n-len(result))...)
	if len(result) >= n {
		return result
	}

	vacant := allVacantCells(f, g)
	result = append(result, randomPoints(vacant, n-len(result))...)
	return result
}

func randomPoints(points []*pb.Point, n int) []*pb.Point {
	rand.Shuffle(len(points), func(i, j int) {
		points[i], points[j] = points[j], points[i]
	})
	return points[0:n]
}

func isHero(s *pb.Snake) bool {
	return !strings.HasPrefix(s.Name, "~~ Level ")
}

func levelComplete(f *pb.GameFrame) bool {
	alive := f.AliveSnakes()
	for _, snake := range alive {
		if !isHero(snake) {
			return false
		}
	}
	return true
}

func startNextLevel(f *pb.GameFrame, g *pb.Game, level int) {
	snakeCount := getSnakeCount(level)
	spawnPoints := findSpawnPoints(f, g, snakeCount)
	for _, p := range spawnPoints {
		spawnAt(p, level, f)
	}
	wipeDeadCampaignSnakes(f)
}

func wipeDeadCampaignSnakes(f *pb.GameFrame) {
	remaining := []*pb.Snake{}
	for _, s := range f.Snakes {
		if s.Death == nil || isHero(s) {
			remaining = append(remaining, s)
		}
	}
	f.Snakes = remaining
}

func updateCampaign(f *pb.GameFrame, g *pb.Game) {
	if levelComplete(f) {
		level := getCurrentLevel(f)
		startNextLevel(f, g, level+1)
	}
}
