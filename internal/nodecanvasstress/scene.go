// Package nodecanvasstress generates deterministic, render-free NodeCanvas scenes.
package nodecanvasstress

import "math"

type Point struct{ X, Y float32 }

type Color struct{ R, G, B, A float32 }

type Layout int

const (
	Grid Layout = iota
	Clusters
)

type Settings struct {
	Count   int
	Layout  Layout
	Spacing float32
	Fanout  int // 0, 1, or 4
	Seed    uint64
}

type Node struct {
	ID     int
	Inputs [2]int
	Output int
	Pos    Point
	Accent Color
}

type Link struct {
	ID, From, To int
}

type Scene struct {
	Nodes []Node
	Links []Link
}

const (
	MaxNodes = 20000
	nodeBase = 1
	pinBase  = 100001
	linkBase = 200001
)

func Generate(s Settings) Scene {
	if s.Count < 0 {
		s.Count = 0
	}
	if s.Count > MaxNodes {
		s.Count = MaxNodes
	}
	if s.Spacing <= 0 {
		s.Spacing = 1
	}
	if s.Fanout != 1 && s.Fanout != 4 {
		s.Fanout = 0
	}
	scene := Scene{Nodes: make([]Node, s.Count), Links: make([]Link, 0, s.Count*s.Fanout)}
	cols := int(math.Ceil(math.Sqrt(float64(s.Count))))
	if cols == 0 {
		cols = 1
	}
	for i := range scene.Nodes {
		// a fixed integer mix gives each ID stable jitter and color without
		// advancing a random stream when the requested count changes.
		h := mix(s.Seed + uint64(i+1))
		x, y := i%cols, i/cols
		if s.Layout == Clusters {
			const perCluster = 100
			cluster, within := i/perCluster, i%perCluster
			x = (cluster%10)*15 + within%10
			y = (cluster/10)*15 + within/10
		}
		jitterX := float32(h&15) / 16
		jitterY := float32((h>>4)&15) / 16
		scene.Nodes[i] = Node{
			ID:     nodeBase + i,
			Inputs: [2]int{pinBase + 3*i, pinBase + 3*i + 1},
			Output: pinBase + 3*i + 2,
			Pos:    Point{X: (float32(x)*240 + jitterX*12) * s.Spacing, Y: (float32(y)*180 + jitterY*12) * s.Spacing},
			Accent: Color{R: 0.25 + float32(h>>8&255)/1024, G: 0.4 + float32(h>>16&255)/1024, B: 0.55 + float32(h>>24&255)/1024, A: 1},
		}
	}
	for i := range scene.Nodes {
		for j := 1; j <= s.Fanout && j < s.Count; j++ {
			// the modulo makes the tail connect to existing nodes. j<n
			// excludes self-links and keeps the four endpoints distinct.
			target := (i + j) % s.Count
			scene.Links = append(scene.Links, Link{
				ID:   linkBase + 4*i + j - 1,
				From: scene.Nodes[i].Output,
				To:   scene.Nodes[target].Inputs[(j-1)%2],
			})
		}
	}
	return scene
}

func mix(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ x>>30) * 0xbf58476d1ce4e5b9
	x = (x ^ x>>27) * 0x94d049bb133111eb
	return x ^ x>>31
}
