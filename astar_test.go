package astar

import (
	"testing"
)

type graph struct {
	edges map[string]map[string]float64
	curr  string
}

var (
	Start  string
	Finish string

	estimateFunc = func(given interface{}) float64 { return 1 }
)

func (g graph) Start() interface{}      { return Start }
func (g graph) Finish() bool            { return g.curr == Finish }
func (g *graph) Move(state interface{}) { g.curr = state.(string) }
func (g graph) Successors(current StatePointer) []interface{} {
	successors := []interface{}{}
	for succ := range g.edges[g.curr] {
		successors = append(successors, succ)
	}

	return successors
}
func (g graph) Cost(given interface{}) float64 {
	return g.edges[g.curr][given.(string)]
}
func (g graph) Estimate(given interface{}) float64 {
	return estimateFunc(given)
}

var BasicTests = []struct {
	name string
	g    *graph
	out  string
}{
	{
		"trivial",
		// (A)--1--(B)
		&graph{edges: map[string]map[string]float64{
			"A": {"B": 1},
		}},
		"AB",
	},
	{
		"prefer longer",
		// (A)
		//  | \
		//  |   3
		//  1     > (C)--1--(D)
		//  |   1
		//  | /
		// (B)
		&graph{edges: map[string]map[string]float64{
			"A": {"C": 3, "B": 1},
			"B": {"A": 1, "C": 1},
			"C": {"A": 3, "B": 1, "D": 1},
			"D": {"C": 1},
		}},
		"ABCD",
	},
	{
		"prefer shorter",
		// (A)
		//  | \
		//  |   1
		//  1     > (C)-----(D)
		//  |   1
		//  | /
		// (B)
		&graph{edges: map[string]map[string]float64{
			"A": {"C": 1, "B": 1},
			"B": {"A": 1, "C": 1},
			"C": {"A": 1, "B": 1, "D": 1},
			"D": {"C": 1},
		}},
		"ACD",
	},
	{
		"directly connected, but bad",
		// (A)--1--(B)
		//  |       |
		//  4       1
		//  |       |
		// (D)--1--(C)
		&graph{edges: map[string]map[string]float64{
			"A": {"B": 1, "D": 4},
			"B": {"A": 1, "C": 1},
			"C": {"B": 1, "D": 1},
			"D": {"C": 1, "A": 4},
		}},
		"ABCD",
	},
	{
		"directly connected, but good",
		// (A)--1--(B)
		//  |       |
		//  1       1
		//  |       |
		// (D)--1--(C)
		&graph{edges: map[string]map[string]float64{
			"A": {"B": 1, "D": 1},
			"B": {"A": 1, "C": 1},
			"C": {"B": 1, "D": 1},
			"D": {"C": 1, "A": 1},
		}},
		"AD",
	},
	{
		"directly connected, equal cost",
		// (A)--1--(B)
		//  |       |
		//  3       1
		//  |       |
		// (D)--1--(C)
		&graph{edges: map[string]map[string]float64{
			"A": {"B": 1, "D": 3},
			"B": {"A": 1, "C": 1},
			"C": {"B": 1, "D": 1},
			"D": {"C": 1, "A": 3},
		}},
		"AD",
	},
	{
		"zigzag",
		// (A)--4--(B)--1--(C)--4--(D)
		//  |       |       |       |
		//  1       1       1       1
		//  |       |       |       |
		// (E)--1--(F)--4--(G)--1--(H)
		&graph{edges: map[string]map[string]float64{
			"A": {"B": 4, "E": 1},
			"B": {"A": 4, "F": 1, "C": 1},
			"C": {"B": 1, "G": 1, "D": 4},
			"D": {"C": 4, "H": 1},
			"E": {"A": 1, "F": 1},
			"F": {"E": 1, "B": 1, "G": 4},
			"G": {"F": 4, "C": 1, "H": 1},
			"H": {"G": 1, "D": 1},
		}},
		"AEFBCGHD",
	},
}

var EstimateTests = []struct {
	name     string
	g        *graph
	estimate func(interface{}) float64
	out      string
}{
	{
		"directly connected, but bad",
		// (A)--1--(B)
		//  |       |
		//  4       1
		//  |       |
		// (D)--1--(C)
		&graph{edges: map[string]map[string]float64{
			"A": {"B": 1, "D": 4},
			"B": {"A": 1, "C": 1},
			"C": {"B": 1, "D": 1},
			"D": {"C": 1, "A": 4},
		}},
		estimateFunc,
		"ABCD",
	},
	{
		"above",
		// (A)--1--(B)--1--(C)     (A)---->(B)---->(C)
		//  |       |       |                       |
		//  1       1       1                       |
		//  |       |       |                       V
		// (D)--1--(E)--1--(F)                     (F)
		&graph{edges: map[string]map[string]float64{
			"A": {"B": 1, "D": 1},
			"B": {"A": 1, "E": 1, "C": 1},
			"C": {"B": 1, "F": 1},
			"D": {"A": 1, "E": 1},
			"E": {"D": 1, "B": 1, "F": 1},
			"F": {"E": 1, "C": 1},
		}},
		func(given interface{}) float64 {
			return map[string]float64{
				"A": 3, "B": 2, "C": 1,
				"D": 4, "E": 4, "F": 0,
			}[given.(string)]
		},
		"ABCF",
	},
	{
		"below",
		// (A)--1--(B)--1--(C)     (A)
		//  |       |       |       |
		//  1       1       1       |
		//  |       |       |       V
		// (D)--1--(E)--1--(F)     (D)---->(E)---->(F)
		&graph{edges: map[string]map[string]float64{
			"A": {"B": 1, "D": 1},
			"B": {"A": 1, "E": 1, "C": 1},
			"C": {"B": 1, "F": 1},
			"D": {"A": 1, "E": 1},
			"E": {"D": 1, "B": 1, "F": 1},
			"F": {"E": 1, "C": 1},
		}},
		func(given interface{}) float64 {
			return map[string]float64{
				"A": 3, "B": 4, "C": 4,
				"D": 2, "E": 1, "F": 0,
			}[given.(string)]
		},
		"ADEF",
	},
}

func stringize(in OptionalState) (out string) {
	for step := in; step != nil; step = step.Previous {
		out = step.Pather.(string) + out
	}
	return
}

func TestTrivial(t *testing.T) {
	Start, Finish = "A", "A"
	if state := Search(&graph{edges: map[string]map[string]float64{}}); stringize(state) != "A" {
		t.Errorf("empty setup: got %v, want A", state)
	}

	Start, Finish = "", "A"
	if state := Search(&graph{edges: map[string]map[string]float64{
		"A": {},
	}}); stringize(state) != "" {
		t.Errorf("no start: got %v, want empty", state)
	}

	Start, Finish = "A", "A"
	if state := Search(&graph{edges: map[string]map[string]float64{
		"A": {},
	}}); stringize(state) != "A" {
		t.Errorf("no successors: got %v, want A", state)
	}

	Start, Finish = "A", "A"
	if state := Search(&graph{edges: map[string]map[string]float64{
		"A": {"A": 1},
	}}); stringize(state) != "A" {
		t.Errorf("same successor: got %v, want A", state)
	}

	Start, Finish = "A", "B"
	if state := Search(&graph{edges: map[string]map[string]float64{
		"A": {"A": 1},
	}}); state != nil {
		t.Errorf("unreachable finish: got %v, want error", state)
	}
}

func TestBasic(t *testing.T) {
	for _, test := range BasicTests {
		Start, Finish = test.out[:1], test.out[len(test.out)-1:]

		if actual := Search(test.g); stringize(actual) != test.out {
			t.Errorf("%q: got %v, want %v", test.name, stringize(actual), test.out)
			if actual == nil {
				t.Fail()
			}
		}
	}
}

func TestEstimate(t *testing.T) {
	for _, test := range EstimateTests {
		Start, Finish = test.out[:1], test.out[len(test.out)-1:]

		estimateFunc = test.estimate

		if actual := Search(test.g); stringize(actual) != test.out {
			t.Errorf("%q: got %v, want %v, path %v", test.name, stringize(actual), test.out, actual.Pather.(string))
			if actual == nil {
				t.Fail()
			}
		}
	}
}
