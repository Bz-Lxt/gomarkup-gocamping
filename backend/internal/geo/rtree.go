package geo

// RTree is a handwritten 2D R-tree with quadratic split.
// Used to coarse-filter danger polygons / circles before exact tests.

const (
	rtMin = 2
	rtMax = 8
)

type RTItem struct {
	ID   int64
	Box  BBox
	Data any
}

type rtNode struct {
	leaf     bool
	box      BBox
	children []*rtNode
	items    []RTItem
}

type RTree struct {
	root *rtNode
	n    int
}

func NewRTree() *RTree {
	return &RTree{root: &rtNode{leaf: true}}
}

func (t *RTree) Len() int { return t.n }

func (t *RTree) Insert(it RTItem) {
	t.n++
	split := t.insert(t.root, it)
	if split != nil {
		old := t.root
		t.root = &rtNode{leaf: false, children: []*rtNode{old, split}}
		t.root.box = old.box.Union(split.box)
	}
}

func (t *RTree) insert(n *rtNode, it RTItem) *rtNode {
	if n.leaf {
		n.items = append(n.items, it)
		n.box = n.box.Union(it.Box)
		if len(n.items) > rtMax {
			return splitLeaf(n)
		}
		return nil
	}
	best := 0
	bestEnl := n.children[0].box.Enlargement(it.Box)
	bestArea := n.children[0].box.Area()
	for i := 1; i < len(n.children); i++ {
		enl := n.children[i].box.Enlargement(it.Box)
		area := n.children[i].box.Area()
		if enl < bestEnl || (enl == bestEnl && area < bestArea) {
			best, bestEnl, bestArea = i, enl, area
		}
	}
	sp := t.insert(n.children[best], it)
	n.box = n.box.Union(it.Box)
	if sp != nil {
		n.children = append(n.children, sp)
		if len(n.children) > rtMax {
			return splitInternal(n)
		}
	}
	return nil
}

func splitLeaf(n *rtNode) *rtNode {
	i, j := pickSeedsItems(n.items)
	a := &rtNode{leaf: true, items: []RTItem{n.items[i]}}
	b := &rtNode{leaf: true, items: []RTItem{n.items[j]}}
	a.box, b.box = n.items[i].Box, n.items[j].Box
	used := map[int]bool{i: true, j: true}
	for len(used) < len(n.items) {
		if len(a.items) >= rtMin && len(b.items)+(len(n.items)-len(used)) <= rtMin {
			assignRemainingItems(n.items, used, b)
			break
		}
		if len(b.items) >= rtMin && len(a.items)+(len(n.items)-len(used)) <= rtMin {
			assignRemainingItems(n.items, used, a)
			break
		}
		k := nextItem(n.items, used, a.box, b.box)
		used[k] = true
		da := a.box.Enlargement(n.items[k].Box)
		db := b.box.Enlargement(n.items[k].Box)
		if da < db || (da == db && a.box.Area() <= b.box.Area()) {
			a.items = append(a.items, n.items[k])
			a.box = a.box.Union(n.items[k].Box)
		} else {
			b.items = append(b.items, n.items[k])
			b.box = b.box.Union(n.items[k].Box)
		}
	}
	*n = *a
	return b
}

func splitInternal(n *rtNode) *rtNode {
	i, j := pickSeedsNodes(n.children)
	a := &rtNode{leaf: false, children: []*rtNode{n.children[i]}}
	b := &rtNode{leaf: false, children: []*rtNode{n.children[j]}}
	a.box, b.box = n.children[i].box, n.children[j].box
	used := map[int]bool{i: true, j: true}
	for len(used) < len(n.children) {
		if len(a.children) >= rtMin && len(b.children)+(len(n.children)-len(used)) <= rtMin {
			for k, c := range n.children {
				if !used[k] {
					b.children = append(b.children, c)
					b.box = b.box.Union(c.box)
				}
			}
			break
		}
		if len(b.children) >= rtMin && len(a.children)+(len(n.children)-len(used)) <= rtMin {
			for k, c := range n.children {
				if !used[k] {
					a.children = append(a.children, c)
					a.box = a.box.Union(c.box)
				}
			}
			break
		}
		k := nextNode(n.children, used, a.box, b.box)
		used[k] = true
		c := n.children[k]
		da := a.box.Enlargement(c.box)
		db := b.box.Enlargement(c.box)
		if da < db || (da == db && a.box.Area() <= b.box.Area()) {
			a.children = append(a.children, c)
			a.box = a.box.Union(c.box)
		} else {
			b.children = append(b.children, c)
			b.box = b.box.Union(c.box)
		}
	}
	*n = *a
	return b
}

func pickSeedsItems(items []RTItem) (int, int) {
	bestI, bestJ := 0, 1
	best := -1.0
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			d := items[i].Box.Union(items[j].Box).Area() - items[i].Box.Area() - items[j].Box.Area()
			if d > best {
				best, bestI, bestJ = d, i, j
			}
		}
	}
	return bestI, bestJ
}

func pickSeedsNodes(ns []*rtNode) (int, int) {
	bestI, bestJ := 0, 1
	best := -1.0
	for i := 0; i < len(ns); i++ {
		for j := i + 1; j < len(ns); j++ {
			d := ns[i].box.Union(ns[j].box).Area() - ns[i].box.Area() - ns[j].box.Area()
			if d > best {
				best, bestI, bestJ = d, i, j
			}
		}
	}
	return bestI, bestJ
}

func nextItem(items []RTItem, used map[int]bool, a, b BBox) int {
	best, bestDiff := -1, -1.0
	for i := range items {
		if used[i] {
			continue
		}
		d := abs(a.Enlargement(items[i].Box) - b.Enlargement(items[i].Box))
		if d > bestDiff {
			best, bestDiff = i, d
		}
	}
	return best
}

func nextNode(ns []*rtNode, used map[int]bool, a, b BBox) int {
	best, bestDiff := -1, -1.0
	for i := range ns {
		if used[i] {
			continue
		}
		d := abs(a.Enlargement(ns[i].box) - b.Enlargement(ns[i].box))
		if d > bestDiff {
			best, bestDiff = i, d
		}
	}
	return best
}

func assignRemainingItems(items []RTItem, used map[int]bool, n *rtNode) {
	for i, it := range items {
		if !used[i] {
			n.items = append(n.items, it)
			n.box = n.box.Union(it.Box)
		}
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func (t *RTree) Search(q BBox) []RTItem {
	if t == nil || t.root == nil {
		return nil
	}
	var out []RTItem
	searchNode(t.root, q, &out)
	return out
}

func searchNode(n *rtNode, q BBox, out *[]RTItem) {
	if n == nil {
		return
	}
	if n.leaf {
		for _, it := range n.items {
			if it.Box.Intersects(q) {
				*out = append(*out, it)
			}
		}
		return
	}
	for _, c := range n.children {
		if c.box.Empty() || c.box.Intersects(q) {
			searchNode(c, q, out)
		}
	}
}

func (t *RTree) All() []RTItem {
	if t == nil || t.root == nil {
		return nil
	}
	var out []RTItem
	collect(t.root, &out)
	return out
}

func collect(n *rtNode, out *[]RTItem) {
	if n.leaf {
		*out = append(*out, n.items...)
		return
	}
	for _, c := range n.children {
		collect(c, out)
	}
}

func (t *RTree) SearchPoint(lat, lon float64) []RTItem {
	return t.Search(BBox{MinLat: lat, MinLon: lon, MaxLat: lat, MaxLon: lon})
}
