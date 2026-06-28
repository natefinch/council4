package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
)

// objectCostChoice is the shared descriptor for the permanents an additional
// cost may select. Candidate enumeration (agent-facing choice presentation) and
// payment planning both derive their eligible set from one objectCostChoice, so
// the objects offered to the agent are exactly the objects the planner will
// accept (CR 601.2b/602.2b). Only object eligibility is unified here; action
// execution (sacrifice vs tap vs return vs counter removal) stays distinct
// because those have different authoritative rules behavior.
type objectCostChoice struct {
	// selection is the canonical characteristic predicate the cost imposes,
	// converted once via SelectionForAdditionalCost (the #1730 matcher).
	selection game.Selection
	// excludeSource drops the cost's own source permanent, backing "another"
	// object costs.
	excludeSource bool
	// requireUntapped restricts eligibility to untapped permanents, for tap
	// costs that tap the chosen permanents.
	requireUntapped bool
}

// objectCostChoiceForCost converts an additional cost's permanent constraint
// into the shared descriptor. It returns false when the constraint is not
// representable as a Selection so callers fail closed, mirroring
// SelectionForAdditionalCost.
func objectCostChoiceForCost(additional cost.Additional) (objectCostChoice, bool) {
	sel, ok := SelectionForAdditionalCost(additional)
	if !ok {
		return objectCostChoice{}, false
	}
	return objectCostChoice{
		selection:       sel,
		excludeSource:   additional.ExcludeSource,
		requireUntapped: additional.Kind == cost.AdditionalTapPermanents || additional.RequireUntapped,
	}, true
}

// permanentSatisfiesObjectCost reports whether one permanent is an eligible
// object for the cost: the player controls it (continuous-effect aware), it is
// not phased out, it honors any untapped/"another" restriction, and it matches
// the canonical selection. It is the single per-permanent gate shared by
// candidate enumeration, preference validation, and deterministic fallback.
func permanentSatisfiesObjectCost(s State, playerID game.PlayerID, p *game.Permanent, choice objectCostChoice, source *game.Permanent) bool {
	if p.PhasedOut || s.EffectiveController(p) != playerID {
		return false
	}
	if choice.requireUntapped && p.Tapped {
		return false
	}
	if choice.excludeSource && source != nil && p.ObjectID == source.ObjectID {
		return false
	}
	return s.PermanentMatchesSelection(p, choice.selection)
}

// candidatePermanentsForObjectCost enumerates, in battlefield order, every
// permanent eligible for the cost that is not already reserved by an earlier
// selection in the same payment. It is the one reservation-aware enumerator used
// by both the choice layer and the planner.
func candidatePermanentsForObjectCost(s State, playerID game.PlayerID, choice objectCostChoice, source *game.Permanent, reserved map[id.ID]bool) []*game.Permanent {
	var candidates []*game.Permanent
	for _, p := range s.Battlefield() {
		if reserved[p.ObjectID] {
			continue
		}
		if permanentSatisfiesObjectCost(s, playerID, p, choice, source) {
			candidates = append(candidates, p)
		}
	}
	return candidates
}

// reservedPermanentIDs collects the object IDs of permanents already selected by
// earlier costs so the shared enumerator never offers the same object twice.
func reservedPermanentIDs(permanents []*game.Permanent) map[id.ID]bool {
	reserved := make(map[id.ID]bool, len(permanents))
	for _, p := range permanents {
		reserved[p.ObjectID] = true
	}
	return reserved
}

// truncatePermanents returns at most amount permanents from a battlefield-order
// candidate list, used by the fixed-count permanent costs (sacrifice, tap,
// return) that take the first eligible objects. A shorter result signals the
// player cannot supply enough objects, which the caller treats as unpayable.
func truncatePermanents(permanents []*game.Permanent, amount int) []*game.Permanent {
	if len(permanents) > amount {
		return permanents[:amount]
	}
	return permanents
}

// CandidatePermanentsForCost returns the permanents the player may choose for an
// additional permanent cost — the same eligible set the planner enumerates. The
// choice layer calls it through the payment State so candidate presentation and
// payment planning share one selection pipeline. excludedIDs are permanents
// already reserved by earlier costs in the same payment.
func CandidatePermanentsForCost(s State, playerID game.PlayerID, additional cost.Additional, source *game.Permanent, excludedIDs ...id.ID) []*game.Permanent {
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	reserved := make(map[id.ID]bool, len(excludedIDs))
	for _, excludedID := range excludedIDs {
		reserved[excludedID] = true
	}
	return candidatePermanentsForObjectCost(s, playerID, choice, source, reserved)
}
