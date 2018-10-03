package typedvalues

import (
	"sort"
	"strconv"

	"github.com/sirupsen/logrus"
)

const (
	MetadataPriority = "priority"
)

// NamedInput provides the TypedValue along with an associated key.
type NamedInput struct {
	Key string
	Val *TypedValue
}

type namedInputSlice []NamedInput

func (n namedInputSlice) Len() int      { return len(n) }
func (n namedInputSlice) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n namedInputSlice) Less(i, j int) bool {
	return priority(n[i].Val) < priority(n[j].Val)
}

func sortNamedInputSlices(inputs []NamedInput) {
	sort.Sort(sort.Reverse(namedInputSlice(inputs)))
}

func priority(t *TypedValue) int {
	var p int
	if ps, ok := t.GetMetadata()[MetadataPriority]; ok {
		i, err := strconv.Atoi(ps)
		if err != nil {
			logrus.Warnf("Ignoring invalid priority: %v", ps)
		} else {
			p = i
		}
	}
	return p
}

func toNamedInputs(inputs map[string]*TypedValue) []NamedInput {
	out := make([]NamedInput, len(inputs))
	var i int
	for k, v := range inputs {
		out[i] = NamedInput{
			Val: v,
			Key: k,
		}
		i++
	}
	return out
}

// Prioritize sorts the inputs based on the priority label (descending order)
func Prioritize(inputs map[string]*TypedValue) []NamedInput {
	namedInputs := toNamedInputs(inputs)
	sortNamedInputSlices(namedInputs)
	return namedInputs
}
