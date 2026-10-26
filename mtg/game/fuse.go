package game

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CanFuse reports whether this card offers the fused cast choice FaceFused: a
// split card with fuse (CR 702.102a) whose halves are both ordinary
// instant/sorcery spells this Engine can combine exactly. Each half must have a
// printed mana cost without X, one locked non-modal spell body without shared
// targets or a legacy card implementation, no alternative, optional or
// mandatory additional costs, no overload, and only Fuse or split second
// static abilities. A fuse card outside that subset is not offered the fused
// choice rather than receiving an approximate combined spell; its halves remain
// castable one at a time.
func (c *CardDef) CanFuse() bool {
	if c.Layout != LayoutSplit || !c.Alternate.Exists {
		return false
	}
	left, right := &c.CardFace, &c.Alternate.Val
	if !left.HasKeyword(Fuse) && !right.HasKeyword(Fuse) {
		return false
	}
	return fusableHalf(left) && fusableHalf(right)
}

func fusableHalf(f *CardFace) bool {
	if (!f.HasType(types.Instant) && !f.HasType(types.Sorcery)) || f.IsPermanent() {
		return false
	}
	if !f.ManaCost.Exists || slices.ContainsFunc(f.ManaCost.Val, func(s cost.Symbol) bool { return s.Kind == cost.VariableSymbol }) {
		return false
	}
	if f.ImplementationID != "" || f.Overload.Exists || len(f.AlternativeCosts) > 0 ||
		len(f.AdditionalCostChoices) > 0 || len(f.AdditionalCosts) > 0 {
		return false
	}
	if len(f.ActivatedAbilities) > 0 || len(f.ManaAbilities) > 0 || len(f.LoyaltyAbilities) > 0 ||
		len(f.TriggeredAbilities) > 0 || len(f.ChapterAbilities) > 0 || len(f.ReplacementAbilities) > 0 {
		return false
	}
	if !f.SpellAbility.Exists {
		return false
	}
	content := &f.SpellAbility.Val
	if content.IsModal() || len(content.SharedTargets) > 0 || len(content.LinkedProducers) > 0 ||
		content.Modes[0].Cost.Exists {
		return false
	}
	for i := range f.StaticAbilities {
		if !fusableStaticAbility(&f.StaticAbilities[i]) {
			return false
		}
	}
	return true
}

func fusableStaticAbility(ability *StaticAbility) bool {
	if len(ability.KeywordAbilities) == 0 || ability.Condition.Exists ||
		len(ability.ContinuousEffects) > 0 || len(ability.RuleEffects) > 0 {
		return false
	}
	for _, keyword := range ability.KeywordAbilities {
		simple, ok := keyword.(SimpleKeyword)
		if !ok || (simple.Kind != Fuse && simple.Kind != SplitSecond) {
			return false
		}
	}
	return true
}

// fusedFace combines both halves into the characteristics of a fused split
// spell (CR 702.102b, 709.4d). The name joins both printed names for display;
// the mana cost keeps each half's separate symbols (CR 709.4b), so the fused
// mana value and colors come from the combined cost, and the total cost
// includes the mana cost of each half (CR 702.102c). The spell body locks both
// halves as two chosen modes, left then right (CR 702.102d), so each half
// keeps its own target segment and later halves still resolve when one
// half's targets become illegal (CR 608.2b). There is no mode choice
// (CR 700.2): FaceFused always casts with modes 0 and 1.
func (c *CardDef) fusedFace() (CardFace, bool) {
	if !c.CanFuse() {
		return CardFace{}, false
	}
	left, right := &c.CardFace, &c.Alternate.Val
	manaCost := make(cost.Mana, 0, len(left.ManaCost.Val)+len(right.ManaCost.Val))
	manaCost = append(manaCost, left.ManaCost.Val...)
	manaCost = append(manaCost, right.ManaCost.Val...)
	return CardFace{
		Name:       left.Name + " // " + right.Name,
		ManaCost:   opt.Val(manaCost),
		Colors:     fusedColors(left.Colors, right.Colors),
		Supertypes: unionInOrder(left.Supertypes, right.Supertypes),
		Types:      unionInOrder(left.Types, right.Types),
		Subtypes:   unionInOrder(left.Subtypes, right.Subtypes),
		SpellAbility: opt.Val(AbilityContent{
			Modes:    []Mode{left.SpellAbility.Val.Modes[0], right.SpellAbility.Val.Modes[0]},
			MinModes: 2,
			MaxModes: 2,
		}),
		StaticAbilities: append(slices.Clone(left.StaticAbilities), right.StaticAbilities...),
		OracleText:      left.OracleText + "\n//\n" + right.OracleText,
	}, true
}

// FusedModes returns the locked chosen modes of a FaceFused cast: the left half
// then the right half (CR 702.102d).
func FusedModes() []int {
	return []int{0, 1}
}

// FaceLabel returns a display label for a cast face choice. A fused split
// spell is labeled with both halves' names and "(fused)"; every printed face
// uses its own name.
func (c *CardDef) FaceLabel(index FaceIndex) string {
	face, ok := c.FaceView(index)
	if !ok {
		return c.Name
	}
	if index == FaceFused {
		return face.Name + " (fused)"
	}
	return face.Name
}

func fusedColors(left, right []color.Color) []color.Color {
	var colors []color.Color
	for _, candidate := range color.AllColors() {
		if slices.Contains(left, candidate) || slices.Contains(right, candidate) {
			colors = append(colors, candidate)
		}
	}
	return colors
}

func unionInOrder[T comparable](left, right []T) []T {
	result := slices.Clone(left)
	for _, value := range right {
		if !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}
