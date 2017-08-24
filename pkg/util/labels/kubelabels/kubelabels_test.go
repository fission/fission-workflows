package kubelabels

import (
	"testing"

	"k8s.io/apimachinery/pkg/labels"
)

func TestSelector(t *testing.T) {

	lbls := New(map[string]string{
		"foo": "bar",
	})

	selector := Selector{
		selector: labels.Everything(),
	}

	if !selector.Matches(lbls) {
		t.Error("Expected labels not matched")
	}
}

func TestEmptyLabel(t *testing.T) {
	lbls := New(nil)

	selector := Selector{
		selector: labels.Everything(),
	}

	if !selector.Matches(lbls) {
		t.Error("Expected labels not matched")
	}
}
