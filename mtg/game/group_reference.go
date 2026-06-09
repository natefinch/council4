package game

import (
	"fmt"

	"github.com/natefinch/council4/opt"
)

// GroupReferenceDomain identifies the candidate set a GroupReference draws
// permanents from before its Selection narrows them.
type GroupReferenceDomain int

// Group reference domain values identify supported candidate domains. The zero
// value groupDomainNone is invalid so a zero-value GroupReference never names a
// real group.
const (
	groupDomainNone GroupReferenceDomain = iota

	// GroupDomainBattlefield draws from every permanent on the battlefield.
	GroupDomainBattlefield

	// GroupDomainAttachedObject draws the single permanent that the anchor
	// object is attached to, such as the creature an Equipment equips.
	GroupDomainAttachedObject

	// GroupDomainObjectControlled draws from battlefield permanents controlled
	// by the controller of the anchor object, such as the creatures the
	// defending player controls.
	GroupDomainObjectControlled
)

// GroupReference is pure rules data describing WHERE a mass effect finds a group
// of permanents: a candidate domain, a Selection that narrows it, an optional
// anchor object that the domain is defined relative to, and an optional excluded
// object. Selection still describes WHAT matches; GroupReference describes the
// binding. The zero value is invalid.
type GroupReference struct {
	domain    GroupReferenceDomain
	selection Selection
	anchor    opt.V[ObjectReference]
	exclude   opt.V[ObjectReference]
}

// BattlefieldGroup matches every battlefield permanent satisfying selection.
func BattlefieldGroup(selection Selection) GroupReference {
	return GroupReference{domain: GroupDomainBattlefield, selection: selection}
}

// BattlefieldGroupExcluding matches every battlefield permanent satisfying
// selection except the permanent identified by exclude.
func BattlefieldGroupExcluding(selection Selection, exclude ObjectReference) GroupReference {
	return GroupReference{domain: GroupDomainBattlefield, selection: selection, exclude: opt.Val(exclude)}
}

// AttachedObjectGroup matches the single permanent that anchor is attached to.
func AttachedObjectGroup(anchor ObjectReference) GroupReference {
	return GroupReference{domain: GroupDomainAttachedObject, anchor: opt.Val(anchor)}
}

// ObjectControlledGroup matches every battlefield permanent controlled by the
// controller of anchor and satisfying selection.
func ObjectControlledGroup(anchor ObjectReference, selection Selection) GroupReference {
	return GroupReference{domain: GroupDomainObjectControlled, selection: selection, anchor: opt.Val(anchor)}
}

// ObjectControlledGroupExcluding matches every battlefield permanent controlled
// by the controller of anchor and satisfying selection, except the permanent
// identified by exclude.
func ObjectControlledGroupExcluding(anchor ObjectReference, selection Selection, exclude ObjectReference) GroupReference {
	return GroupReference{domain: GroupDomainObjectControlled, selection: selection, anchor: opt.Val(anchor), exclude: opt.Val(exclude)}
}

// Domain reports the candidate domain the group draws from.
func (r GroupReference) Domain() GroupReferenceDomain { return r.domain }

// Selection returns the characteristic predicate that narrows the domain.
func (r GroupReference) Selection() Selection { return r.selection }

// Anchor returns the object the domain is defined relative to, if any.
func (r GroupReference) Anchor() (ObjectReference, bool) {
	return r.anchor.Val, r.anchor.Exists
}

// Exclusion returns the object dropped from the group, if any.
func (r GroupReference) Exclusion() (ObjectReference, bool) {
	return r.exclude.Val, r.exclude.Exists
}

// Valid reports whether the GroupReference names a structurally sound group.
func (r GroupReference) Valid() bool {
	return len(r.Validate()) == 0
}

// Validate reports structural problems with a GroupReference that represent
// card-definition bugs rather than board-state outcomes.
func (r GroupReference) Validate() []string {
	var problems []string
	switch r.domain {
	case GroupDomainBattlefield:
		if r.anchor.Exists {
			problems = append(problems, "battlefield group must not set an anchor object")
		}
	case GroupDomainAttachedObject:
		if !r.anchor.Exists {
			problems = append(problems, "attached-object group requires an anchor object")
		}
		if !r.selection.Empty() {
			problems = append(problems, "attached-object group must not set a Selection")
		}
		if r.exclude.Exists {
			problems = append(problems, "attached-object group must not set an exclusion")
		}
	case GroupDomainObjectControlled:
		if !r.anchor.Exists {
			problems = append(problems, "object-controlled group requires an anchor object")
		}
	case groupDomainNone:
		problems = append(problems, "group reference has no domain")
	default:
		problems = append(problems, fmt.Sprintf("unknown group domain %d", r.domain))
	}
	if r.anchor.Exists {
		problems = appendPrefixed(problems, "anchor", r.anchor.Val.Validate())
	}
	if r.exclude.Exists {
		problems = appendPrefixed(problems, "exclude", r.exclude.Val.Validate())
	}
	problems = append(problems, r.selection.Validate()...)
	return problems
}

// appendPrefixed appends each problem from src to dst with a "prefix: " label so
// nested reference problems retain their location.
func appendPrefixed(dst []string, prefix string, src []string) []string {
	for _, problem := range src {
		dst = append(dst, prefix+": "+problem)
	}
	return dst
}
