package planner

import (
	"astrohop/internal/astronomy/coordinates"
)

type path struct {
	Next     *path
	Id       int
	Distance float64
}

type node struct {
	Pos coordinates.Equatorial
}

func useFarthestInsertion(nodes []node) *path {
	// distance matrix for quick lookup
	distMtrx := make([][]float64, len(nodes))
	for i, ni := range nodes {
		distMtrx[i] = make([]float64, len(nodes))
		for j, nj := range nodes {
			if i == j {
				continue
			}
			dist := coordinates.DistanceEq(ni.Pos, nj.Pos)
			distMtrx[i][j] = dist
		}
	}
	// list of nodes already in the route
	seenNodes := make(map[int]struct{}, len(nodes))

	nextIdx := make([]int, len(nodes))
	prevIdx := make([]int, len(nodes))

	// determine the farthest node form the fist one
	maxDist := 0.0
	nextNode := 0
	for i := 1; i < len(nodes); i++ {
		if distMtrx[0][i] > maxDist {
			maxDist = distMtrx[0][i]
			nextNode = i
		}
	}

	// build the initial route
	nextIdx[0] = nextNode
	prevIdx[0] = nextNode
	nextIdx[nextNode] = 0
	prevIdx[nextNode] = 0

	seenNodes[0] = struct{}{}
	seenNodes[nextNode] = struct{}{}

	// finish building the route
	for len(seenNodes) < len(nodes) {
		maxDist = 0.0
		nextNode = 0
		// select farthest node from current route
		for i := range nodes {
			// since all objects are located on a sphere, angular distance can't exceed 360 degrees
			minDist := 360.0
			if _, ok := seenNodes[i]; ok {
				continue
			}
			for j := range seenNodes {
				dist := distMtrx[i][j]
				if dist < minDist {
					minDist = dist
				}
			}
			if minDist > maxDist {
				maxDist = minDist
				nextNode = i
			}
		}
		minDist := 360.0
		insertEdge := 0
		// select an edge to insert next node to
		for i := range seenNodes {
			j := nextIdx[i]
			dist := distMtrx[i][nextNode] + distMtrx[nextNode][j] - distMtrx[i][j]
			if dist < minDist {
				minDist = dist
				insertEdge = i
			}
		}
		j := nextIdx[insertEdge]
		nextIdx[nextNode] = j
		prevIdx[nextNode] = insertEdge
		prevIdx[j] = nextNode
		nextIdx[insertEdge] = nextNode
		seenNodes[nextNode] = struct{}{}
	}

	p := &path{
		Id: 0,
	}
	currPath := p
	next := nextIdx[0]
	for next != 0 {
		nextPath := &path{
			Id: next,
		}
		currPath.Distance = distMtrx[currPath.Id][next]
		currPath.Next = nextPath
		currPath = nextPath
		next = nextIdx[next]
	}
	currPath.Next = p
	currPath.Distance = distMtrx[currPath.Id][0]

	return p
}
