package graph

import (
	"hash/fnv"

	"github.com/fission/fission-workflows/pkg/types"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type LinkedNode interface {
	graph.Node
	Links() []int64
}

type TaskInstanceNode struct {
	*types.TaskInstance
}

func (n *TaskInstanceNode) ID() int64 {
	return createId(n.Task.Id())
}

func (n *TaskInstanceNode) Links() []int64 {
	var links []int64

	for k := range n.Task.Spec.Requires {
		links = append(links, createId(k))
	}
	return links
}

type TaskSpecNode struct {
	Id string
	*types.TaskSpec
}

func (n *TaskSpecNode) Links() []int64 {
	var links []int64
	for k := range n.Requires {
		links = append(links, createId(k))
	}
	return links
}

func (n *TaskSpecNode) ID() int64 {
	return createId(n.Id)
}

type Iterator interface {
	Get(ptr int) LinkedNode
}

type TaskInstanceIterator struct {
	contents map[string]*types.TaskInstance
	keys     []string
}

func NewTaskInstanceIterator(contents map[string]*types.TaskInstance) *TaskInstanceIterator {
	var keys []string
	for k := range contents {
		keys = append(keys, k)
	}
	return &TaskInstanceIterator{
		contents: contents,
		keys:     keys,
	}
}

func (ti *TaskInstanceIterator) Get(ptr int) LinkedNode {
	if len(ti.keys) > ptr {
		return &TaskInstanceNode{
			TaskInstance: ti.contents[ti.keys[ptr]],
		}
	}
	return nil
}

type TaskSpecIterator struct {
	contents map[string]*types.TaskSpec
	keys     []string
}

func (ts *TaskSpecIterator) Get(ptr int) LinkedNode {
	if len(ts.keys) > ptr {
		return &TaskSpecNode{
			TaskSpec: ts.contents[ts.keys[ptr]],
			Id:       ts.keys[ptr],
		}
	}
	return nil
}

func NewTaskSpecIterator(contents map[string]*types.TaskSpec) *TaskSpecIterator {
	var keys []string
	for k := range contents {
		keys = append(keys, k)
	}
	return &TaskSpecIterator{
		contents: contents,
		keys:     keys,
	}
}

func Parse(it Iterator) graph.Directed {
	depGraph := simple.NewDirectedGraph()
	// Add Nodes
	for ptr := 0; it.Get(ptr) != nil; ptr += 1 {
		depGraph.AddNode(it.Get(ptr))
	}

	// Add dependencies
	for ptr := 0; it.Get(ptr) != nil; ptr += 1 {
		n := it.Get(ptr)
		if n.Links() != nil {
			for _, dep := range n.Links() {
				depNode := depGraph.Node(dep)
				if depNode != nil {
					e := depGraph.NewEdge(depNode, n)
					depGraph.SetEdge(e)
				}
			}
		}
	}

	// Temporary: Alter the dependencies of dynamic task dependents
	// This will eventually be moved into the eventstore implementation in the form of a
	// middleware, to avoid having to recalculate the graph.
	deps := map[int64]map[string]*types.TaskDependencyParameters{}
	for _, v := range depGraph.Nodes() {
		switch n := v.(type) {
		case *TaskSpecNode:
			deps[n.ID()] = n.Requires
		case *TaskInstanceNode:
			deps[n.ID()] = n.Task.Spec.Requires
		}
	}
	for _, v := range depGraph.Nodes() {
		injectDynamicTask(depGraph, v, deps)
	}

	return depGraph
}

func Roots(g graph.Directed) []graph.Node {
	var roots []graph.Node
	for _, n := range g.Nodes() {
		in := g.To(n)
		if len(in) == 0 {
			roots = append(roots, n)
		}
	}
	return roots
}

func createId(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func injectDynamicTask(depGraph *simple.DirectedGraph, dynamic graph.Node,
	nodeRequires map[int64]map[string]*types.TaskDependencyParameters) {

	// find parent id
	var parentTaskId string
	var found bool
	for dep, params := range nodeRequires[dynamic.ID()] {
		if params != nil && params.Type == types.TaskDependencyParameters_DYNAMIC_OUTPUT {
			parentTaskId = dep
			found = true
			break
		}
	}
	if !found {
		return
	}

	// Add edges from the dynamic task to all nodes depending on the parent.
	for nodeId, deps := range nodeRequires {
		if _, ok := deps[parentTaskId]; ok && nodeId != dynamic.ID() {
			depNode := depGraph.Node(nodeId)
			depGraph.SetEdge(depGraph.NewEdge(dynamic, depNode))
		}
	}
}

func Get(g graph.Graph, id string) graph.Node {
	idHash := createId(id)
	for _, v := range g.Nodes() {
		if v.ID() == idHash {
			return v
		}
	}
	return nil
}
