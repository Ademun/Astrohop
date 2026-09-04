package algo

type Tour struct {
	Order     []int
	Distances [][]float64
}

func UseFarthestInsertion(distanceMtrx [][]float64) *Tour {
	// list of distanceMtrx already in the route
	seenNodes := make(map[int]struct{}, len(distanceMtrx))

	nextNodeIdx := make([]int, len(distanceMtrx))

	// determine the farthest node from the fist one
	maxDist := 0.0
	nextNode := 0
	for i := 1; i < len(distanceMtrx); i++ {
		if distanceMtrx[0][i] > maxDist {
			maxDist = distanceMtrx[0][i]
			nextNode = i
		}
	}

	// build the initial tour
	nextNodeIdx[0] = nextNode
	nextNodeIdx[nextNode] = 0
	seenNodes[0] = struct{}{}
	seenNodes[nextNode] = struct{}{}

	// finish building the tour
	for len(seenNodes) < len(distanceMtrx) {
		maxDist = 0.0
		nextNode = 0
		// select farthest node from current tour
		for i := range distanceMtrx {
			// since all objects are located on a sphere, angular distance can't exceed 360 degrees
			minDist := 360.0
			if _, ok := seenNodes[i]; ok {
				continue
			}
			for j := range seenNodes {
				dist := distanceMtrx[i][j]
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
			j := nextNodeIdx[i]
			dist := distanceMtrx[i][nextNode] + distanceMtrx[nextNode][j] - distanceMtrx[i][j]
			if dist < minDist {
				minDist = dist
				insertEdge = i
			}
		}
		j := nextNodeIdx[insertEdge]
		nextNodeIdx[nextNode] = j
		nextNodeIdx[insertEdge] = nextNode
		seenNodes[nextNode] = struct{}{}
	}

	order := make([]int, len(distanceMtrx))
	cur := 0
	for i := 1; i < len(order); i++ {
		cur = nextNodeIdx[cur]
		order[i] = cur
	}

	tour := &Tour{
		Distances: distanceMtrx,
		Order:     order,
	}

	return tour
}
