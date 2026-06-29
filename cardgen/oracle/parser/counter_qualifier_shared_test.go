package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// counterFields captures only the counter-qualifier fields the shared mappers
// touch, so the assertions stay focused and avoid comparing the unrelated
// (and non-comparable) parts of the selection and subject structs.
type counterFields struct {
	required   bool
	kind       counter.Kind
	any        bool
	absent     bool
	kindAbsent bool
}

// TestSelectionApplyCounterQualifier verifies the shared selection mapper
// records every counterQualifierKind variant onto the selection's counter
// fields. It is the single mapping all selection-qualifier call sites share, so
// the modeled fields must not drift between the named, kind-agnostic, negated,
// and kind-specific-negated forms.
func TestSelectionApplyCounterQualifier(t *testing.T) {
	tests := []struct {
		name  string
		match counterQualifierMatch
		want  counterFields
	}{
		{
			name:  "named",
			match: counterQualifierMatch{Kind: counter.PlusOnePlusOne},
			want:  counterFields{required: true, kind: counter.PlusOnePlusOne},
		},
		{
			name:  "any",
			match: counterQualifierMatch{Any: true},
			want:  counterFields{required: true, any: true},
		},
		{
			name:  "absent",
			match: counterQualifierMatch{Absent: true},
			want:  counterFields{absent: true},
		},
		{
			name:  "kind absent",
			match: counterQualifierMatch{Kind: counter.PlusOnePlusOne, KindAbsent: true},
			want:  counterFields{kindAbsent: true, kind: counter.PlusOnePlusOne},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sel SelectionSyntax
			sel.applyCounterQualifier(test.match)
			got := counterFields{
				required:   sel.CounterRequired,
				kind:       sel.CounterKind,
				any:        sel.CounterAny,
				absent:     sel.CounterAbsent,
				kindAbsent: sel.CounterKindAbsent,
			}
			if got != test.want {
				t.Fatalf("applyCounterQualifier(%+v) = %+v, want %+v", test.match, got, test.want)
			}
		})
	}
}

// TestStaticSubjectApplyCounterQualifier verifies the shared static-subject
// mapper accepts the positive named and kind-agnostic qualifiers, sets the
// subject's counter fields, and fails closed without mutation for the negated
// forms that have no modeled counter-filtered group subject.
func TestStaticSubjectApplyCounterQualifier(t *testing.T) {
	tests := []struct {
		name   string
		match  counterQualifierMatch
		wantOK bool
		want   counterFields
	}{
		{
			name:   "named",
			match:  counterQualifierMatch{Kind: counter.PlusOnePlusOne},
			wantOK: true,
			want:   counterFields{required: true, kind: counter.PlusOnePlusOne},
		},
		{
			name:   "any",
			match:  counterQualifierMatch{Any: true},
			wantOK: true,
			want:   counterFields{required: true, any: true},
		},
		{
			name:   "absent fails closed",
			match:  counterQualifierMatch{Absent: true},
			wantOK: false,
			want:   counterFields{},
		},
		{
			name:   "kind absent fails closed",
			match:  counterQualifierMatch{Kind: counter.PlusOnePlusOne, KindAbsent: true},
			wantOK: false,
			want:   counterFields{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sub EffectStaticSubjectSyntax
			ok := sub.applyCounterQualifier(test.match)
			if ok != test.wantOK {
				t.Fatalf("applyCounterQualifier(%+v) ok = %v, want %v", test.match, ok, test.wantOK)
			}
			got := counterFields{
				required: sub.CounterRequired,
				kind:     sub.CounterKind,
				any:      sub.CounterAny,
			}
			if got != test.want {
				t.Fatalf("applyCounterQualifier(%+v) subject = %+v, want %+v", test.match, got, test.want)
			}
		})
	}
}
