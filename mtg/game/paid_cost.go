package game

import (
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PaidCostKind distinguishes a consumed permanent from a discarded card.
type PaidCostKind uint8

// Paid cost subject kinds.
const (
	PaidCostUnknown PaidCostKind = iota
	PaidCostSacrifice
	PaidCostDiscard
)

// PaidCostSubject records one actual member of a paid cost component. Snapshot
// is frozen immediately before payment moves the object, not read from its
// later incarnation. An unavailable definition is not an empty characteristic set.
type PaidCostSubject struct {
	Key                  string
	Kind                 PaidCostKind
	Snapshot             ObjectSnapshot
	CardZoneVersion      opt.V[uint64]
	CharacteristicsKnown bool
}

// ClonePaidCostSubjects preserves the immutable payment facts in independent
// stack copies and cloned games.
func ClonePaidCostSubjects(subjects []PaidCostSubject) []PaidCostSubject {
	return cloneSliceFunc(subjects, func(subject PaidCostSubject) PaidCostSubject {
		subject.Snapshot = cloneObjectSnapshot(subject.Snapshot)
		return subject
	})
}

// PaidCostReference names a singular actual member of one exact cost component.
func PaidCostReference(key string, kind PaidCostKind, noun types.Card) ObjectReference {
	return ObjectReference{kind: ObjectReferencePaidCost, costKey: key, costKind: kind, costNoun: noun}
}

// CostKey reports the paid cost component's identity.
func (r ObjectReference) CostKey() string { return r.costKey }

// CostKind reports the paid subject's domain.
func (r ObjectReference) CostKind() PaidCostKind { return r.costKind }

// CostNoun is an optional actual-subject type restriction, separate from the
// predicate's negation and from the cost's payment eligibility.
func (r ObjectReference) CostNoun() types.Card { return r.costNoun }
