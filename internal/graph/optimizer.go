package graph

import (
	"sort"
	"strings"
)

// Optimizer selects relevant context based on the graph
type Optimizer struct {
	graph *Graph
}

// NewOptimizer creates a new context optimizer
func NewOptimizer(g *Graph) *Optimizer {
	return &Optimizer{graph: g}
}

// OptimizeContext returns the most relevant files for a query
func (o *Optimizer) OptimizeContext(query string, limit int) []string {
	// 1. Identify entry points (nodes matching query terms)
	queryTerms := strings.Fields(strings.ToLower(query))
	entryPoints := make(map[int64]float64)

	for id, node := range o.graph.Nodes {
		score := 0.0
		pathLower := strings.ToLower(node.Path)

		for _, term := range queryTerms {
			if strings.Contains(pathLower, term) {
				score += 1.0
			}
		}

		if score > 0 {
			entryPoints[id] = score
		}
	}

	// 2. Expand to neighbors (1 hop)
	candidates := make(map[int64]float64)
	for id, score := range entryPoints {
		candidates[id] += score * 2.0 // Entry points are very important

		// Add outgoing neighbors (dependencies)
		for _, outID := range o.graph.OutEdges[id] {
			candidates[outID] += score * 0.5
		}

		// Add incoming neighbors (dependents)
		for _, inID := range o.graph.InEdges[id] {
			candidates[inID] += score * 0.5
		}
	}

	// 3. Combine with PageRank
	type Result struct {
		Path  string
		Score float64
	}
	var results []Result

	for id, relevance := range candidates {
		node := o.graph.Nodes[id]
		// Final score = Relevance * PageRank
		// PageRank helps prioritize "central" files among the relevant ones
		finalScore := relevance * (1.0 + node.Score*10.0)

		results = append(results, Result{
			Path:  node.Path,
			Score: finalScore,
		})
	}

	// 4. Sort and limit
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	var paths []string
	for i, res := range results {
		if i >= limit {
			break
		}
		paths = append(paths, res.Path)
	}

	return paths
}
