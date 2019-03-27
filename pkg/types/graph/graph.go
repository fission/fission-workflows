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

type TaskInvocationNode struct {
	*types.TaskInvocation
}

func (n *TaskInvocationNode) ID() int64 {
	return createID(n.Task().ID())
}

func (n *TaskInvocationNode) Links() []int64 {
	var links []int64

	for k := range n.Task().Spec.Requires {
		links = append(links, createID(k))
	}
	return links
}

type TaskSpecNode struct {
	id string
	*types.TaskSpec
}

func (n *TaskSpecNode) Links() []int64 {
	var links []int64
	for k := range n.Requires {
		links = append(links, createID(k))
	}
	return links
}

func (n *TaskSpecNode) ID() int64 {
	return createID(n.id)
}

type Iterator interface {
	Get(ptr int) LinkedNode
}

type TaskInstanceIterator struct {
	contents map[string]*types.TaskInvocation
	keys     []string
}

func NewTaskInstanceIterator(contents map[string]*types.TaskInvocation) *TaskInstanceIterator {
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
		return &TaskInvocationNode{
			TaskInvocation: ti.contents[ti.keys[ptr]],
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
			id:       ts.keys[ptr],
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
	for ptr := 0; it.Get(ptr) != nil; ptr++ {
		depGraph.AddNode(it.Get(ptr))
	}

	// Add dependencies
	for ptr := 0; it.Get(ptr) != nil; ptr++ {
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
		case *TaskInvocationNode:
			deps[n.ID()] = n.Task().Spec.Requires
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

func createID(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func injectDynamicTask(depGraph *simple.DirectedGraph, dynamic graph.Node,
	nodeRequires map[int64]map[string]*types.TaskDependencyParameters) {

	// find parent id
	var parentTaskID string
	var found bool
	for dep, params := range nodeRequires[dynamic.ID()] {
		if params != nil && params.Type == types.TaskDependencyParameters_DYNAMIC_OUTPUT {
			parentTaskID = dep
			found = true
			break
		}
	}
	if !found {
		return
	}

	// Add edges from the dynamic task to all nodes depending on the parent.
	for nodeID, deps := range nodeRequires {
		if _, ok := deps[parentTaskID]; ok && nodeID != dynamic.ID() {
			depNode := depGraph.Node(nodeID)
			depGraph.SetEdge(depGraph.NewEdge(dynamic, depNode))
		}
	}
}

func Get(g graph.Graph, id string) graph.Node {
	idHash := createID(id)
	for _, v := range g.Nodes() {
		if v.ID() == idHash {
			return v
		}
	}
	return nil
}
