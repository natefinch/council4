package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// bandsWithOtherQualities returns the [quality] of every "bands with other"
// ability the permanent currently has (CR 702.22a). Abilities lost to an
// effect removing banding or "bands with other" are excluded (CR 702.22b), as
// are level-band abilities whose level range isn't met. Abilities with the same
// quality are redundant (CR 702.22m), so duplicates may be returned; callers
// treat the list as a set.
func bandsWithOtherQualities(g *game.Game, permanent *game.Permanent) []game.Selection {
	values := effectivePermanentValues(g, permanent)
	if !values.keywords.has(game.BandsWithOther) {
		return nil
	}
	var qualities []game.Selection
	for _, ability := range values.abilities[min(values.bandsWithOtherFrom, len(values.abilities)):] {
		body, ok := ability.(*game.StaticAbility)
		if !ok || !levelBandKeywordConditionSatisfied(permanent, body.Condition) {
			continue
		}
		for _, keyword := range game.StaticBodyBandsWithOtherKeywords(body) {
			qualities = append(qualities, keyword.Quality)
		}
	}
	return qualities
}

// bandQuality is one "bands with other [quality]" ability among a group of
// creatures: the creature holding it and which creatures in the group have
// that quality.
type bandQuality struct {
	holder  int
	matches []bool
}

// bandEligibility records, for a group of creatures, which have banding and
// every "bands with other" ability whose holder itself has the ability's
// quality (CR 702.22c, 702.22j-k). It is computed once per decision so band
// legality checks are pure.
type bandEligibility struct {
	banding   []bool
	qualities []bandQuality
}

// newBandEligibility evaluates banding and "bands with other" for creatures.
// Nil entries have neither. Each quality is matched against every creature's
// current characteristics, with "you" being the holder's controller.
func newBandEligibility(g *game.Game, creatures []*game.Permanent) bandEligibility {
	eligibility := bandEligibility{banding: make([]bool, len(creatures))}
	for holder, creature := range creatures {
		if creature == nil {
			continue
		}
		eligibility.banding[holder] = hasKeyword(g, creature, game.Banding)
		for _, quality := range bandsWithOtherQualities(g, creature) {
			matches := make([]bool, len(creatures))
			for index, other := range creatures {
				matches[index] = other != nil && permanentHasBandQuality(g, creature, other, quality)
			}
			if matches[holder] {
				eligibility.qualities = append(eligibility.qualities, bandQuality{holder: holder, matches: matches})
			}
		}
	}
	return eligibility
}

// permanentHasBandQuality reports whether creature has the [quality] of
// holder's "bands with other" ability, read from creature's current
// characteristics.
func permanentHasBandQuality(g *game.Game, holder, creature *game.Permanent, quality game.Selection) bool {
	controller := effectiveController(g, holder)
	subject := selectionSubject{
		kind:       subjectPermanent,
		g:          g,
		permanent:  creature,
		values:     effectivePermanentValues(g, creature),
		controller: effectiveController(g, creature),
		viewer:     controller,
	}
	return matchSelection(&subject, &quality)
}

// canStartBand reports whether creature index may anchor an attacking band:
// it has banding, or a "bands with other" ability whose quality it has.
func (b bandEligibility) canStartBand(index int) bool {
	if b.banding[index] {
		return true
	}
	for _, quality := range b.qualities {
		if quality.holder == index {
			return true
		}
	}
	return false
}

// legalBand reports whether the creatures at members may form one attacking
// band (CR 702.22c): at least two creatures, and either one or more with
// banding and at most one without (a creature with only "bands with other"
// counts as one without), or every member has some quality [quality] and at
// least one member has "bands with other [quality]". An ability
// held by a creature outside members never makes the band legal.
func (b bandEligibility) legalBand(members []int) bool {
	if len(members) < 2 {
		return false
	}
	withBanding, withoutBanding := 0, 0
	for _, member := range members {
		if b.banding[member] {
			withBanding++
		} else {
			withoutBanding++
		}
	}
	if withBanding > 0 && withoutBanding <= 1 {
		return true
	}
	for _, quality := range b.qualities {
		if slices.Contains(members, quality.holder) && allMatch(quality.matches, members) {
			return true
		}
	}
	return false
}

// sharesBandsWithOther reports whether a [quality] creature with "bands with
// other [quality]" and another [quality] creature are both in the group, the
// "bands with other" half of CR 702.22j-k.
func (b bandEligibility) sharesBandsWithOther() bool {
	for _, quality := range b.qualities {
		for index, match := range quality.matches {
			if match && index != quality.holder {
				return true
			}
		}
	}
	return false
}

func allMatch(matches []bool, members []int) bool {
	for _, member := range members {
		if !matches[member] {
			return false
		}
	}
	return true
}
