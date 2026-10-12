package genetics

import "sort"

// TreeNode is a node of the displayable pedigree tree. Unknown ancestors are
// placeholders (Known=false) that remember which animal and role they belong
// to, so the UI can offer to create a minimal record.
type TreeNode struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Registration string    `json:"registration"`
	Sex          string    `json:"sex"`
	BreedName    string    `json:"breedName"`
	BirthDate    string    `json:"birthDate"`
	Generation   int       `json:"generation"` // 0 = the animal itself
	Role         string    `json:"role"`       // "self" | "sire" | "dam"
	Known        bool      `json:"known"`
	ChildID      string    `json:"childId"`   // for placeholders: the animal missing this parent
	ChildName    string    `json:"childName"` // display name of that animal
	HasMore      bool      `json:"hasMore"`   // known ancestors exist beyond the displayed depth
	Sire         *TreeNode `json:"sire"`
	Dam          *TreeNode `json:"dam"`
}

// BuildTree builds the pedigree tree of id with `generations` ancestor
// generations (1 = parents, 2 = + grandparents, ...). Unknown ancestors are
// not expanded, which keeps sparse pedigrees compact.
func BuildTree(g Graph, id string, generations int) *TreeNode {
	root := g[id]
	if root == nil {
		return nil
	}
	return build(g, root, 0, "self", "", generations, map[string]bool{})
}

func build(g Graph, n *Node, gen int, role, childID string, max int, onPath map[string]bool) *TreeNode {
	t := &TreeNode{
		ID: n.ID, Name: n.Name, Registration: n.Registration, Sex: n.Sex, BreedName: n.BreedName,
		BirthDate: n.BirthDate, Generation: gen, Role: role, Known: true, ChildID: childID,
	}
	if onPath[n.ID] { // defensive: never recurse through a cycle
		return t
	}
	onPath[n.ID] = true
	defer delete(onPath, n.ID)

	if gen >= max {
		t.HasMore = n.FatherID != "" || n.MotherID != ""
		return t
	}
	t.Sire = parentNode(g, n, n.FatherID, gen+1, "sire", max, onPath)
	t.Dam = parentNode(g, n, n.MotherID, gen+1, "dam", max, onPath)
	return t
}

func parentNode(g Graph, child *Node, pid string, gen int, role string, max int, onPath map[string]bool) *TreeNode {
	if p := g[pid]; pid != "" && p != nil {
		return build(g, p, gen, role, child.ID, max, onPath)
	}
	name := child.Name
	if name == "" {
		name = child.Registration
	}
	return &TreeNode{Generation: gen, Role: role, Known: false, ChildID: child.ID, ChildName: name}
}

// Repeated is an ancestor appearing more than once in the displayed tree
// (pedigree collapse), which is the visible trace of inbreeding.
type Repeated struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Occurrences []string `json:"occurrences"` // paths such as "SD", "DSS"
}

// RepeatedAncestors lists ancestors that occur at multiple positions.
func RepeatedAncestors(g Graph, id string, maxGen int) []Repeated {
	paths := map[string][]string{}
	for _, o := range Ancestors(g, id, maxGen) {
		paths[o.ID] = append(paths[o.ID], o.Path)
	}
	var out []Repeated
	for aid, ps := range paths {
		if len(ps) < 2 {
			continue
		}
		sort.Strings(ps)
		r := Repeated{ID: aid, Occurrences: ps}
		if n := g[aid]; n != nil {
			r.Name = n.Name
			if r.Name == "" {
				r.Name = n.Registration
			}
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
