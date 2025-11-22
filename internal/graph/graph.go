package graph

import (
	"math"
)

// Node represents a file in the dependency graph
type Node struct {
	ID    int64
	Path  string
	Type  string // "file", "package"
	Score float64
}

// Graph represents the dependency graph
type Graph struct {
	Nodes    map[int64]*Node
	Paths    map[string]int64
	OutEdges map[int64][]int64
	InEdges  map[int64][]int64
	NextID   int64
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		Nodes:    make(map[int64]*Node),
		Paths:    make(map[string]int64),
		OutEdges: make(map[int64][]int64),
		InEdges:  make(map[int64][]int64),
		NextID:   1,
	}
}

// AddNode adds a node to the graph if it doesn't exist
func (g *Graph) AddNode(path string, nodeType string) int64 {
	if id, exists := g.Paths[path]; exists {
		return id
	}
	id := g.NextID
	g.NextID++
	g.Nodes[id] = &Node{
		ID:    id,
		Path:  path,
		Type:  nodeType,
		Score: 1.0, // Initial score
	}
	g.Paths[path] = id
	return id
}

// AddEdge adds a directed edge from source to target
func (g *Graph) AddEdge(fromPath, toPath string) {
	fromID := g.AddNode(fromPath, "file")
	toID := g.AddNode(toPath, "file")

	// Check if edge already exists
	for _, id := range g.OutEdges[fromID] {
		if id == toID {
			return
		}
	}

	g.OutEdges[fromID] = append(g.OutEdges[fromID], toID)
	g.InEdges[toID] = append(g.InEdges[toID], fromID)
}

// PageRank computes the importance of each node
// damping: usually 0.85
// iterations: usually 10-20
func (g *Graph) PageRank(damping float64, iterations int) {
	N := float64(len(g.Nodes))
	if N == 0 {
		return
	}

	// Initialize scores uniformly
	for _, node := range g.Nodes {
		node.Score = 1.0 / N
	}

	newScores := make(map[int64]float64)

	for i := 0; i < iterations; i++ {
		diff := 0.0
		sum := 0.0

		for id := range g.Nodes {
			rank := (1.0 - damping) / N

			// Sum of incoming edges
			for _, inID := range g.InEdges[id] {
				outDegree := len(g.OutEdges[inID])
				if outDegree > 0 {
					rank += damping * (g.Nodes[inID].Score / float64(outDegree))
				}
			}

			newScores[id] = rank
			sum += rank
		}

		// Normalize to ensure sum = 1.0
		if sum > 0 {
			for id := range newScores {
				newScores[id] /= sum
			}
		}

		// Update scores and calculate diff
		for id, score := range newScores {
			diff += math.Abs(g.Nodes[id].Score - score)
			g.Nodes[id].Score = score
		}

		// Check convergence
		if diff < 0.0001 {
			break
		}
	}
}
