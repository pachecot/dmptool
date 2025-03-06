package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tpacheco/dmptool/dmp"
)

type node struct {
	name     string
	object   *dmp.Object
	children []*node
}

func (n *node) Name() string {
	return n.name
}

func (n *node) String() string {
	return n.name
}

func (n *node) Path() string {
	if n.object == nil {
		return n.name
	}
	return n.object.DeviceId
}

func (n *node) Children() []*node {
	return n.children
}

func (n *node) Properties() map[string]string {
	if n.object == nil {
		return make(map[string]string)
	}
	return n.object.Properties
}

type treeHandler struct {
	// dmp.EmptyHandler
	objects     []*dmp.Object
	rootPath    string
	currentPath string
	tree        tree
}

func (h *treeHandler) Dictionary(dd *dmp.Dictionary) {
}

func (h *treeHandler) Object(do *dmp.Object) {
	if do.DeviceId == "" {
		do.DeviceId = h.currentPath
	}
	h.objects = append(h.objects, do)
}

func (h *treeHandler) Path(s string) {
	h.rootPath = s
	h.currentPath = strings.ReplaceAll(s, " ", "")
}

func (h *treeHandler) Begin(tag string, name string) {
	switch tag {
	case "Container", "Device":
		h.currentPath = filepath.Join(h.currentPath, name)
	}
}

func (h *treeHandler) End(tag string, name string) {
	switch tag {
	case "Container", "Device":
		h.currentPath = filepath.Dir(h.currentPath)
	}
}

type Command struct {
	FileName string
	OutFile  string
	Ascii    bool
	Depth    int
	Parents  bool
}

func prune(n *node) *node {
	x := len(n.children)
	if x == 0 {
		return nil
	}
	children := make([]*node, 0, x)
	for _, c := range n.children {
		a := prune(c)
		if a == nil {
			continue
		}
		children = append(children, a)
	}
	if len(children) == 0 && n.object == nil {
		return nil
	}
	n.children = children
	return n
}

func (cmd *Command) Execute() {

	h := &treeHandler{tree: newTree(cmd.Depth, cmd.Ascii)}

	dmp.ParseFile(cmd.FileName, h)

	if len(h.objects) == 0 {
		fmt.Println("no results")
		return
	}
	slices.SortFunc(h.objects, dmp.ObjectPathCompare)

	root := buildGraph(h)
	if cmd.Parents {
		root = prune(root)
	}
	view := strings.Split(h.tree.view(root, ""), "\n")

	w := os.Stdout
	if cmd.OutFile != "" {
		out, err := os.Create(cmd.OutFile)
		if err != nil {
			fmt.Println("could not create file")
			return
		}
		defer func() {
			out.Sync()
			out.Close()
		}()
		w = out
	}
	writeFile(w, view)
}

type tree struct {
	vt       string
	br       string
	lf       string
	ws       string
	indent   string
	depth    int
	maxDepth int
}

const (
	ansi_vt = `│   `
	ansi_br = `├───`
	ansi_lf = `└───`
	ansi_ws = `    `

	ascii_vt = `|   `
	ascii_br = `|---`
	ascii_lf = `\---`
	ascii_ws = `    `
)

func newTree(maxDepth int, ascii bool) tree {
	if ascii {
		return tree{
			maxDepth: maxDepth,
			vt:       ascii_vt,
			br:       ascii_br,
			lf:       ascii_lf,
			ws:       ascii_ws,
		}
	}
	return tree{
		maxDepth: maxDepth,
		vt:       ansi_vt,
		br:       ansi_br,
		lf:       ansi_lf,
		ws:       ansi_ws,
	}
}

func buildGraph(h *treeHandler) *node {
	rootPath := h.rootPath

	root := &node{name: h.rootPath}

	items := make([]*node, 0, len(h.objects))
	nm := make(map[string]*node, len(h.objects))

	nm[rootPath] = root

	for _, o := range h.objects {

		n := &node{
			name:   o.Name,
			object: o,
		}
		pt := o.Path
		if n2, ok := nm[pt]; ok {
			n = n2
			n.object = o
			n.name = o.Name
		}
		nm[pt] = n

		parent := filepath.Dir(o.Path)
		devNode := nm[parent]
		if devNode == nil {
			devNode = &node{
				name: o.Type,
			}
			nm[parent] = devNode
		}
		id := parent + "/" + o.Type
		typeNode := nm[id]
		if typeNode == nil {
			typeNode = &node{
				name: o.Type,
			}
			nm[id] = typeNode
			devNode.children = append(devNode.children, typeNode)
		}
		typeNode.children = append(typeNode.children, n)

		items = append(items, n)
	}
	if root.children == nil {
		root.children = items
	}
	if r, ok := nm[h.rootPath]; ok {
		r.name = h.rootPath
		root = r
	}
	return root
}

func (t tree) view(item *node, prefix string) string {
	t.depth++
	cs := item.children
	ss := make([]string, 0, len(cs)+1)
	if item.object != nil {
		io := item.object
		if len(io.Alias) > 0 && io.Alias != item.name {
			ss = append(ss, t.indent+prefix+io.Alias+" ["+item.name+"]")
		} else {
			ss = append(ss, t.indent+prefix+item.name)
		}
	} else {
		ss = append(ss, t.indent+prefix+item.name)
	}
	switch prefix {
	case t.lf:
		t.indent += t.ws
	case t.br:
		t.indent += t.vt
	}
	pfx := t.br
	if t.maxDepth > 0 && t.depth > t.maxDepth {
		return strings.Join(ss, "\n")
	}

	for i, c := range cs {
		if i == len(cs)-1 {
			pfx = t.lf
		}
		ss = append(ss, t.view(c, pfx))
	}
	return strings.Join(ss, "\n")
}

func writeFile(w *os.File, table []string) {
	for _, row := range table {
		fmt.Fprintln(w, row)
	}
}
