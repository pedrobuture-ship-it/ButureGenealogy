// Package genetics holds pure, deterministic genetics algorithms. It has no
// dependency on SQL, Wails or the UI, so every function is unit-testable with
// synthetic data.
package genetics

import (
	"sort"
	"strings"
)

// Node is one animal in a pedigree graph.
type Node struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Registration string `json:"registration"`
	Sex          string `json:"sex"` // "male" | "female"
	BreedName    string `json:"breedName"`
	BirthDate    string `json:"birthDate"` // YYYY-MM-DD or ""
	FatherID     string `json:"fatherId"`
	MotherID     string `json:"motherId"`
}

// Graph maps animal ID -> node. Parents referenced by a node but absent from
// the graph are treated as "known but not loaded" (e.g. beyond the load depth).
type Graph map[string]*Node

// Occurrence is one position of an ancestor in an animal's pedigree.
// Path is the lineage from the animal: "S" = sire, "D" = dam, so "SD" is the
// sire's dam. Generation = len(Path) (1 = parents, 2 = grandparents, ...).
type Occurrence struct {
	ID         string `json:"id"`
	Generation int    `json:"generation"`
	Path       string `json:"path"`
}

// Ancestors lists every ancestor position up to maxGen generations, breadth
// first. The same animal can appear several times (pedigree collapse).
// Expansion is bounded by maxGen, so malformed cyclic data cannot loop forever.
func Ancestors(g Graph, id string, maxGen int) []Occurrence {
	type item struct {
		id   string
		path string
	}
	var out []Occurrence
	level := []item{{id, ""}}
	for gen := 1; gen <= maxGen; gen++ {
		var next []item
		for _, it := range level {
			n := g[it.id]
			if n == nil {
				continue
			}
			if n.FatherID != "" {
				p := it.path + "S"
				out = append(out, Occurrence{n.FatherID, gen, p})
				next = append(next, item{n.FatherID, p})
			}
			if n.MotherID != "" {
				p := it.path + "D"
				out = append(out, Occurrence{n.MotherID, gen, p})
				next = append(next, item{n.MotherID, p})
			}
		}
		level = next
	}
	return out
}

// CommonAncestor is an animal that is an ancestor of both subjects.
// The subjects themselves count as generation 0 of their own lineage, so a
// sire is a common ancestor of himself and his offspring.
type CommonAncestor struct {
	ID    string `json:"id"`
	GenA  []int  `json:"generationsFromA"` // distinct generations at which it appears for A
	GenB  []int  `json:"generationsFromB"`
	Depth int    `json:"depth"` // smallest genA+genB over all pairs of occurrences
}

func generationsByID(g Graph, id string, maxGen int) map[string]map[int]bool {
	m := map[string]map[int]bool{id: {0: true}}
	for _, o := range Ancestors(g, id, maxGen) {
		if m[o.ID] == nil {
			m[o.ID] = map[int]bool{}
		}
		m[o.ID][o.Generation] = true
	}
	return m
}

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// CommonAncestors returns the common ancestors of a and b within maxGen
// generations, closest first (ties broken by ID for determinism).
func CommonAncestors(g Graph, a, b string, maxGen int) []CommonAncestor {
	ga, gb := generationsByID(g, a, maxGen), generationsByID(g, b, maxGen)
	var out []CommonAncestor
	for id, as := range ga {
		bs, ok := gb[id]
		if !ok {
			continue
		}
		ca := CommonAncestor{ID: id, GenA: sortedKeys(as), GenB: sortedKeys(bs)}
		ca.Depth = ca.GenA[0] + ca.GenB[0]
		out = append(out, ca)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Depth != out[j].Depth {
			return out[i].Depth < out[j].Depth
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// DetectCycle returns the IDs forming a pedigree cycle (an animal that is its
// own ancestor), or nil when the graph is acyclic. A cycle is biologically
// impossible and indicates corrupted or wrongly imported data.
func DetectCycle(g Graph) []string {
	const (
		white = 0
		grey  = 1
		black = 2
	)
	color := map[string]int{}
	var stack []string
	var found []string

	var visit func(id string) bool
	visit = func(id string) bool {
		color[id] = grey
		stack = append(stack, id)
		n := g[id]
		if n != nil {
			for _, p := range []string{n.FatherID, n.MotherID} {
				if p == "" || g[p] == nil {
					continue
				}
				switch color[p] {
				case grey:
					for i, s := range stack {
						if s == p {
							found = append([]string{}, stack[i:]...)
							return true
						}
					}
				case white:
					if visit(p) {
						return true
					}
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
		return false
	}

	ids := make([]string, 0, len(g))
	for id := range g {
		ids = append(ids, id)
	}
	sort.Strings(ids) // deterministic result
	for _, id := range ids {
		if color[id] == white && visit(id) {
			return found
		}
	}
	return nil
}

// GenerationStat describes how complete one generation of a pedigree is.
type GenerationStat struct {
	Generation int     `json:"generation"`
	Expected   int     `json:"expected"` // 2^generation positions
	Known      int     `json:"known"`    // positions with a recorded ancestor
	Percent    float64 `json:"percent"`
}

// Completeness reports known vs expected ancestor positions per generation.
func Completeness(g Graph, id string, maxGen int) []GenerationStat {
	stats := make([]GenerationStat, maxGen)
	for i := range stats {
		stats[i] = GenerationStat{Generation: i + 1, Expected: 1 << (i + 1)}
	}
	for _, o := range Ancestors(g, id, maxGen) {
		stats[o.Generation-1].Known++
	}
	for i := range stats {
		stats[i].Percent = 100 * float64(stats[i].Known) / float64(stats[i].Expected)
	}
	return stats
}

// Depth is the number of generations (up to maxGen) that contain at least one
// known ancestor. 0 means no parent is recorded.
func Depth(g Graph, id string, maxGen int) int {
	d := 0
	for _, o := range Ancestors(g, id, maxGen) {
		if o.Generation > d {
			d = o.Generation
		}
	}
	return d
}

// Issue is a pedigree inconsistency.
type Issue struct {
	Severity string `json:"severity"` // "error" | "warning"
	Code     string `json:"code"`
	AnimalID string `json:"animalId"`
	Message  string `json:"message"`
}

// CheckConsistency finds basic inconsistencies among the loaded nodes.
// Parents absent from the graph are skipped (they may simply be beyond the
// load depth); referential integrity is verified separately by the database.
func CheckConsistency(g Graph) []Issue {
	var out []Issue
	add := func(sev, code, id, msg string) { out = append(out, Issue{sev, code, id, msg}) }

	ids := make([]string, 0, len(g))
	for id := range g {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		n := g[id]
		label := n.Name
		if label == "" {
			label = n.Registration
		}
		if n.FatherID != "" && n.FatherID == id || n.MotherID != "" && n.MotherID == id {
			add("error", "self_parent", id, label+": o animal consta como o próprio pai ou mãe.")
		}
		if n.FatherID != "" && n.FatherID == n.MotherID {
			add("error", "same_parents", id, label+": pai e mãe são o mesmo animal.")
		}
		if f := g[n.FatherID]; f != nil {
			if f.Sex != "male" {
				add("error", "father_not_male", id, label+": o pai registrado não é macho.")
			}
			if after(f.BirthDate, n.BirthDate) {
				add("error", "parent_younger_than_child", id, label+": o pai nasceu depois do animal.")
			}
		}
		if m := g[n.MotherID]; m != nil {
			if m.Sex != "female" {
				add("error", "mother_not_female", id, label+": a mãe registrada não é fêmea.")
			}
			if after(m.BirthDate, n.BirthDate) {
				add("error", "parent_younger_than_child", id, label+": a mãe nasceu depois do animal.")
			}
		}
	}
	if cyc := DetectCycle(g); cyc != nil {
		add("error", "cycle", cyc[0], "Ciclo no pedigree: "+strings.Join(cyc, " → ")+".")
	}
	return out
}

// after reports whether parent was born on or after the child (both known).
func after(parentBirth, childBirth string) bool {
	return parentBirth != "" && childBirth != "" && parentBirth >= childBirth
}
