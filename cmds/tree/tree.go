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
}

func (cmd *Command) Execute() {

	h := &treeHandler{}

	dmp.ParseFile(cmd.FileName, h)

	if len(h.objects) == 0 {
		fmt.Println("no results")
		return
	}
	slices.SortFunc(h.objects, dmp.ObjectPathCompare)

	tree := buildTree(h)

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
	writeFile(w, tree)
}

const (
	indentI = "│  "
	prefixT = "├──"
	prefixL = "└──"
	indentS = "   "
)

func buildGraph(h *treeHandler) *node {
	// rootPath := strings.ReplaceAll(h.rootPath, " ", "")
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

		// id := o.DeviceId
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

		// switch o.Type {
		// case "Container", "Folder", "Device":
		// 	dev := filepath.Join(o.DeviceId, o.Name)
		// 	if _, ok := nm[dev]; !ok {
		// 		// nm[dev] = typeNode
		// 	}
		// }
	}
	if root.children == nil {
		root.children = items
	}
	if r, ok := nm[h.rootPath]; ok {
		r.name = h.rootPath
		root = r
	}
	// fmt.Printf("%#v", root.children[1].children[0])
	return root
}

func buildGraphS(h *treeHandler) *node {
	rootPath := strings.ReplaceAll(h.rootPath, " ", "")

	root := &node{name: h.rootPath}

	items := make([]*node, 0, len(h.objects))
	nm := make(map[string]*node, len(h.objects))

	nm[rootPath] = root

	for _, o := range h.objects {
		if !strings.HasPrefix(o.DeviceId, rootPath) {
			b := filepath.Base(rootPath)
			if strings.HasPrefix(o.DeviceId, b) {
				o.DeviceId = filepath.Join(filepath.Dir(rootPath), o.DeviceId)
			}
		}
		if !strings.HasPrefix(o.DeviceId, o.Path) {
			o.Path = filepath.Join(o.DeviceId, filepath.Base(o.Path))
		}
	}

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

		devNode := nm[o.DeviceId]
		if devNode == nil {
			devNode = &node{
				name: o.Type,
			}
			nm[o.DeviceId] = devNode
		}

		// id := o.DeviceId
		id := o.DeviceId + "/" + o.Type
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

		// switch o.Type {
		// case "Container", "Folder", "Device":
		// 	dev := filepath.Join(o.DeviceId, o.Name)
		// 	if _, ok := nm[dev]; !ok {
		// 		// nm[dev] = typeNode
		// 	}
		// }
	}
	if root.children == nil {
		root.children = items
	}
	if r, ok := nm[h.rootPath]; ok {
		r.name = h.rootPath
		root = r
	}
	// fmt.Printf("%#v", root.children[1].children[0])
	return root
}

func nodeView(item *node, prefix string, indent string) string {
	cs := item.children
	ss := make([]string, 0, len(cs)+1)
	if item.object != nil {
		io := item.object
		if len(io.Alias) > 0 && io.Alias != item.name {
			ss = append(ss, indent+prefix+io.Alias+" ["+item.name+"]")
		} else {
			ss = append(ss, indent+prefix+item.name)
		}
	} else {
		ss = append(ss, indent+prefix+item.name)
	}
	switch prefix {
	case prefixL:
		indent += indentS
	case prefixT:
		indent += indentI
	}
	for i, c := range cs {
		pfx := prefixT
		if i == len(cs)-1 {
			pfx = prefixL
		}
		ss = append(ss, nodeView(c, pfx, indent))
	}
	return strings.Join(ss, "\n")
}

func buildTree(h *treeHandler) []string {
	root := buildGraph(h)
	ss := strings.Split(nodeView(root, "", ""), "\n")
	return ss
}

func writeFile(w *os.File, table []string) {
	for _, row := range table {
		fmt.Fprintln(w, row)
	}
}
