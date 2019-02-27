package pb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnake_Move(t *testing.T) {
	tests := []struct {
		Direction string
		Expected  *Point
	}{
		{
			Direction: "up",
			Expected:  &Point{X: 5, Y: 4},
		},
		{
			Direction: "down",
			Expected:  &Point{X: 5, Y: 6},
		},
		{
			Direction: "left",
			Expected:  &Point{X: 4, Y: 5},
		},
		{
			Direction: "right",
			Expected:  &Point{X: 6, Y: 5},
		},
		{
			Direction: "",
			Expected:  &Point{X: 5, Y: 4},
		},
	}

	for _, test := range tests {
		s := &Snake{
			Body: []*Point{
				{X: 5, Y: 5},
			},
		}
		s.Move(test.Direction)
		require.Equal(t, test.Expected, s.Head(), "Direction: %s", test.Direction)
	}
}

func TestSnake_DefaultMove(t *testing.T) {
	tests := []struct {
		Body     []*Point
		Expected *Point
	}{
		{
			Body: []*Point{
				{X: 5, Y: 5},
				{X: 5, Y: 5},
			},
			Expected: &Point{X: 5, Y: 4},
		},
		{
			Body: []*Point{
				{X: 5, Y: 5},
				{X: 5, Y: 4},
			},
			Expected: &Point{X: 5, Y: 6},
		},
		{
			Body: []*Point{
				{X: 5, Y: 4},
				{X: 5, Y: 5},
			},
			Expected: &Point{X: 5, Y: 3},
		},
		{
			Body: []*Point{
				{X: 4, Y: 5},
				{X: 5, Y: 5},
			},
			Expected: &Point{X: 3, Y: 5},
		},
		{
			Body: []*Point{
				{X: 5, Y: 5},
				{X: 4, Y: 5},
			},
			Expected: &Point{X: 6, Y: 5},
		},
	}

	for _, test := range tests {
		s := &Snake{
			Body: test.Body,
		}
		s.DefaultMove()
		require.Equal(t, test.Expected, s.Head())
	}
}

func TestSnake_Tail(t *testing.T) {
	s := &Snake{
		Body: []*Point{
			{X: 5, Y: 5},
			{X: 4, Y: 5},
		},
	}

	require.Equal(t, &Point{X: 4, Y: 5}, s.Tail())
}
