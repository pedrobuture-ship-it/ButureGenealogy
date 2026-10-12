package genetics

import (
	"fmt"
	"reflect"
	"testing"
)

func n(id, sex, father, mother string) *Node {
	return &Node{ID: id, Name: id, Sex: sex, FatherID: father, MotherID: mother}
}

func graph(nodes ...*Node) Graph {
	g := Graph{}
	for _, x := range nodes {
		g[x.ID] = x
	}
	return g
}

// Pedigree used in several tests:
//
//	calf  = sire S x dam D
//	S     = GS (grandsire)  x GD1
//	D     = GS (same bull!) x GD2     <- half-sib mating: GS is on both sides
func collapsed() Graph {
	return graph(
		n("calf", "male", "S", "D"),
		n("S", "male", "GS", "GD1"),
		n("D", "female", "GS", "GD2"),
		n("GS", "male", "", ""),
		n("GD1", "female", "", ""),
		n("GD2", "female", "", ""),
	)
}

func TestParentsAreGenerationOne(t *testing.T) {
	g := graph(n("a", "male", "s", "d"), n("s", "male", "", ""), n("d", "female", "", ""))
	got := Ancestors(g, "a", 3)
	want := []Occurrence{{"s", 1, "S"}, {"d", 1, "D"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if Depth(g, "a", 5) != 1 {
		t.Fatalf("depth: %d", Depth(g, "a", 5))
	}
}

func TestAncestorPathsAndRepetition(t *testing.T) {
	occ := Ancestors(collapsed(), "calf", 3)
	paths := map[string][]string{}
	for _, o := range occ {
		paths[o.ID] = append(paths[o.ID], o.Path)
	}
	if !reflect.DeepEqual(paths["GS"], []string{"SS", "DS"}) {
		t.Fatalf("GS paths: %v", paths["GS"])
	}
	rep := RepeatedAncestors(collapsed(), "calf", 3)
	if len(rep) != 1 || rep[0].ID != "GS" || len(rep[0].Occurrences) != 2 {
		t.Fatalf("repeated: %+v", rep)
	}
}

func TestCommonAncestors(t *testing.T) {
	g := collapsed()
	// the parents S and D share GS as a common ancestor (they are half-siblings)
	ca := CommonAncestors(g, "S", "D", 5)
	if len(ca) != 1 || ca[0].ID != "GS" || ca[0].Depth != 2 {
		t.Fatalf("got %+v", ca)
	}
	// parent and offspring: the parent itself is a common ancestor (generation 0 for A)
	ca = CommonAncestors(g, "S", "calf", 5)
	if len(ca) < 1 || ca[0].ID != "S" || ca[0].GenA[0] != 0 || ca[0].GenB[0] != 1 {
		t.Fatalf("parent/offspring: %+v", ca)
	}
	// unrelated animals share nothing
	g["x"] = n("x", "male", "", "")
	if got := CommonAncestors(g, "calf", "x", 5); len(got) != 0 {
		t.Fatalf("unrelated: %+v", got)
	}
}

func TestIncompletePedigree(t *testing.T) {
	g := graph(n("a", "male", "s", ""), n("s", "male", "", "")) // dam unknown
	stats := Completeness(g, "a", 3)
	if stats[0].Known != 1 || stats[0].Expected != 2 || stats[0].Percent != 50 {
		t.Fatalf("gen1: %+v", stats[0])
	}
	if stats[1].Known != 0 || stats[2].Known != 0 {
		t.Fatalf("deeper generations must be empty: %+v", stats)
	}
	if Depth(g, "a", 5) != 1 {
		t.Fatal("depth must be 1")
	}
	if Depth(graph(n("lonely", "male", "", "")), "lonely", 5) != 0 {
		t.Fatal("no parents => depth 0")
	}
}

func TestDeepPedigree(t *testing.T) {
	// a straight sire line 30 generations long
	g := Graph{}
	for i := 0; i < 30; i++ {
		father := ""
		if i < 29 {
			father = fmt.Sprintf("b%d", i+1)
		}
		g[fmt.Sprintf("b%d", i)] = n(fmt.Sprintf("b%d", i), "male", father, "")
	}
	if d := Depth(g, "b0", 10); d != 10 {
		t.Fatalf("depth must be capped by maxGen, got %d", d)
	}
	if d := Depth(g, "b0", 50); d != 29 {
		t.Fatalf("full depth should be 29, got %d", d)
	}
	st := Completeness(g, "b0", 5)
	for i, s := range st {
		if s.Known != 1 {
			t.Fatalf("gen %d known=%d", i+1, s.Known)
		}
	}
	tree := BuildTree(g, "b0", 5)
	cur, depth := tree, 0
	for cur.Sire != nil && cur.Sire.Known {
		cur, depth = cur.Sire, depth+1
	}
	if depth != 5 || !cur.HasMore {
		t.Fatalf("tree depth=%d hasMore=%v", depth, cur.HasMore)
	}
}

func TestCycleDetection(t *testing.T) {
	if DetectCycle(collapsed()) != nil {
		t.Fatal("collapsed pedigree is acyclic (shared ancestor is not a cycle)")
	}
	// a -> b -> c -> a
	g := graph(n("a", "male", "b", ""), n("b", "male", "c", ""), n("c", "male", "a", ""))
	cyc := DetectCycle(g)
	if len(cyc) != 3 {
		t.Fatalf("cycle: %v", cyc)
	}
	// traversals must terminate on cyclic data
	if got := Ancestors(g, "a", 6); len(got) != 6 {
		t.Fatalf("bounded traversal: %d", len(got))
	}
	if tree := BuildTree(g, "a", 5); tree == nil {
		t.Fatal("tree must still build")
	}
	issues := CheckConsistency(g)
	found := false
	for _, i := range issues {
		found = found || i.Code == "cycle"
	}
	if !found {
		t.Fatalf("cycle issue missing: %+v", issues)
	}
}

func TestSelfParentIsACycle(t *testing.T) {
	g := graph(n("a", "male", "a", ""))
	if DetectCycle(g) == nil {
		t.Fatal("self loop must be detected")
	}
}

func TestConsistency(t *testing.T) {
	g := graph(
		&Node{ID: "kid", Name: "kid", Sex: "male", BirthDate: "2020-01-01", FatherID: "mum", MotherID: "dad"},
		&Node{ID: "mum", Name: "mum", Sex: "female", BirthDate: "2021-01-01"}, // used as father & born later
		&Node{ID: "dad", Name: "dad", Sex: "male", BirthDate: "2015-01-01"},   // used as mother
		&Node{ID: "twin", Name: "twin", Sex: "male", FatherID: "dad", MotherID: "dad"},
	)
	codes := map[string]int{}
	for _, i := range CheckConsistency(g) {
		codes[i.Code]++
	}
	for _, want := range []string{"father_not_male", "mother_not_female", "parent_younger_than_child", "same_parents"} {
		if codes[want] == 0 {
			t.Fatalf("missing issue %s in %v", want, codes)
		}
	}
	if len(CheckConsistency(collapsed())) != 0 {
		t.Fatal("valid pedigree must have no issues")
	}
}

func TestBuildTreeUnknownIsCompact(t *testing.T) {
	g := graph(n("a", "male", "", ""))
	tree := BuildTree(g, "a", 5)
	if tree.Sire == nil || tree.Sire.Known || tree.Sire.Sire != nil {
		t.Fatalf("unknown parent must be a single placeholder: %+v", tree.Sire)
	}
	if tree.Sire.ChildID != "a" || tree.Sire.ChildName != "a" || tree.Sire.Role != "sire" || tree.Dam.Role != "dam" {
		t.Fatalf("placeholder must remember child and role: %+v", tree.Sire)
	}
	if BuildTree(g, "missing", 3) != nil {
		t.Fatal("unknown root => nil")
	}
}

func TestInbreedingIsNotInvented(t *testing.T) {
	var c RelationshipCalculator = UnavailableCalculator{}
	r := c.Inbreeding(collapsed(), "calf")
	if r.Status != StatusNotCalculated || r.Value != nil || r.Message != "não calculado nesta versão" {
		t.Fatalf("%+v", r)
	}
	if r := c.Relationship(collapsed(), "S", "D"); r.Value != nil {
		t.Fatal("relationship must not invent a value")
	}
}
