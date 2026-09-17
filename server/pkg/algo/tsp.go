package algo

func UseFarthestInsertion(distanceMtrx [][]float64) []int {
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

	return order
}

func Use2Opt(tour []int, distanceMtrx [][]int) []int {
	n := len(tour)
	if n < 4 {
		return append([]int(nil), tour...)
	}

	best := append([]int(nil), tour...)

	for {
		improved := false

		// try all pairs of edges
		for i := 0; i < n-1; i++ {
			for j := i + 1; j < n; j++ {
				// reversing the whole tour is useless
				if i == 0 && j == n-1 {
					continue
				}

				prev := (i - 1 + n) % n
				next := (j + 1) % n

				a, b := best[prev], best[i]
				c, d := best[j], best[next]

				// old edges: (a,b) and (c,d)
				// new edges: (a,c) and (b,d)
				delta := distanceMtrx[a][c] + distanceMtrx[b][d] -
					distanceMtrx[a][b] - distanceMtrx[c][d]

				if delta < 0 {
					// reverse segment best[i:j+1]
					for l, r := i, j; l < r; l, r = l+1, r-1 {
						best[l], best[r] = best[r], best[l]
					}

					improved = true
					break
				}
			}

			if improved {
				break
			}
		}

		if !improved {
			break
		}
	}

	return best
}
