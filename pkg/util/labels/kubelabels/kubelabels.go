package kubelabels

import (
	"github.com/fission/fission-workflow/pkg/util/labels"
	kubelabels "k8s.io/apimachinery/pkg/labels"
)

type LabelSet kubelabels.Set

type Labels struct {
	kubelabels.Labels
}

func New(labelSet LabelSet) labels.Labels {
	return &Labels{kubelabels.Set(labelSet)}
}

type Selector struct {
	selector kubelabels.Selector
}

func NewSelector(selector kubelabels.Selector) *Selector {
	return &Selector{selector}
}

func (kl *Selector) Matches(labels labels.Labels) bool {
	klabel, ok := labels.(*Labels)
	if !ok {
		panic("Invalid label type")
	}
	if kl.selector.Empty() {
		return true
	}

	return kl.selector.Matches(klabel)
}
