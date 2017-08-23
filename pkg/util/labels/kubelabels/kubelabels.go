package kubelabels

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/util/labels"
	kubelabels "k8s.io/apimachinery/pkg/labels"
)

type LabelSet kubelabels.Set

type Labels struct {
	labels kubelabels.Labels
}

func New(labelSet LabelSet) labels.Labels {
	return &Labels{
		labels: kubelabels.Set(labelSet),
	}
}

func (kl *Labels) String() string {
	return fmt.Sprintf("%v", kl.labels)
}

type Selector struct {
	selector kubelabels.Selector
}

func (kl *Selector) Matches(labels labels.Labels) bool {
	klabel, ok := labels.(*Labels)
	if !ok {
		panic("Invalid label type")
	}
	if kl.selector.Empty() {
		return true
	}

	return kl.selector.Matches(klabel.labels)
}
