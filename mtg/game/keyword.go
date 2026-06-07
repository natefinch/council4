package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
)

// KeywordAbility is a sealed data-only variant for one printed keyword ability.
type KeywordAbility interface {
	isKeywordAbility()
	keyword() Keyword
}

// SimpleKeyword is a non-parameterized keyword ability such as flying or haste.
type SimpleKeyword struct {
	Kind Keyword
}

// WardKeyword parameterizes Ward for mana-valued ward costs.
type WardKeyword struct {
	Cost cost.Mana
}

// EquipKeyword parameterizes Equip activation costs.
type EquipKeyword struct {
	Cost cost.Mana
}

// EnchantKeyword parameterizes Enchant attachment legality.
type EnchantKeyword struct {
	Target TargetSpec
}

// CyclingKeyword parameterizes Cycling activation costs.
type CyclingKeyword struct {
	Cost cost.Mana
}

// KickerKeyword parameterizes Kicker additional costs and bonus instructions.
type KickerKeyword struct {
	Cost         cost.Mana
	BonusContent AbilityContent
}

// MadnessKeyword parameterizes Madness alternative costs.
type MadnessKeyword struct {
	Cost cost.Mana
}

// MorphKeyword parameterizes Morph turn-face-up costs.
type MorphKeyword struct {
	Cost cost.Mana
}

// DisguiseKeyword parameterizes Disguise turn-face-up costs.
type DisguiseKeyword struct {
	Cost cost.Mana
}

// SuspendKeyword parameterizes Suspend costs and time counters.
type SuspendKeyword struct {
	Cost         cost.Mana
	TimeCounters int
}

// ProtectionKeyword parameterizes Protection effects.
type ProtectionKeyword struct {
	FromColors []color.Color
}

func (SimpleKeyword) isKeywordAbility()     {}
func (WardKeyword) isKeywordAbility()       {}
func (EquipKeyword) isKeywordAbility()      {}
func (EnchantKeyword) isKeywordAbility()    {}
func (CyclingKeyword) isKeywordAbility()    {}
func (KickerKeyword) isKeywordAbility()     {}
func (MadnessKeyword) isKeywordAbility()    {}
func (MorphKeyword) isKeywordAbility()      {}
func (DisguiseKeyword) isKeywordAbility()   {}
func (SuspendKeyword) isKeywordAbility()    {}
func (ProtectionKeyword) isKeywordAbility() {}

func (ability SimpleKeyword) keyword() Keyword { return ability.Kind }
func (WardKeyword) keyword() Keyword           { return Ward }
func (EquipKeyword) keyword() Keyword          { return Equip }
func (EnchantKeyword) keyword() Keyword        { return Enchant }
func (CyclingKeyword) keyword() Keyword        { return Cycling }
func (KickerKeyword) keyword() Keyword         { return Kicker }
func (MadnessKeyword) keyword() Keyword        { return Madness }
func (MorphKeyword) keyword() Keyword          { return Morph }
func (DisguiseKeyword) keyword() Keyword       { return Disguise }
func (SuspendKeyword) keyword() Keyword        { return Suspend }
func (ProtectionKeyword) keyword() Keyword     { return Protection }

// SimpleKeywords returns sealed keyword variants for non-parameterized keywords.
func SimpleKeywords(keywords ...Keyword) []KeywordAbility {
	abilities := make([]KeywordAbility, 0, len(keywords))
	for _, keyword := range keywords {
		abilities = append(abilities, SimpleKeyword{Kind: keyword})
	}
	return abilities
}

// KeywordAbilityKind returns the Keyword represented by a sealed keyword variant.
func KeywordAbilityKind(ability KeywordAbility) Keyword {
	if ability == nil {
		panic("game: nil KeywordAbility")
	}
	return ability.keyword()
}

// BodyKeywordAbilities returns the keyword abilities carried by a sealed body.
func BodyKeywordAbilities(body Ability) []KeywordAbility {
	switch b := body.(type) {
	case StaticAbility:
		return b.KeywordAbilities
	case *StaticAbility:
		if b == nil {
			return nil
		}
		return b.KeywordAbilities
	case ActivatedAbility:
		return b.KeywordAbilities
	case *ActivatedAbility:
		if b == nil {
			return nil
		}
		return b.KeywordAbilities
	case TriggeredAbility:
		return b.KeywordAbilities
	case *TriggeredAbility:
		if b == nil {
			return nil
		}
		return b.KeywordAbilities
	default:
		return nil
	}
}

// BodyHasKeyword reports whether a sealed body carries the given keyword.
func BodyHasKeyword(body Ability, kw Keyword) bool {
	_, ok := BodyKeywordAbility(body, kw)
	return ok
}

// BodyKeywordAbility returns the first sealed keyword variant matching kw on body.
func BodyKeywordAbility(body Ability, kw Keyword) (KeywordAbility, bool) {
	for _, ka := range BodyKeywordAbilities(body) {
		if KeywordAbilityKind(ka) == kw {
			return ka, true
		}
	}
	return nil, false
}

// BodyAddKeywordKindsTo adds every keyword represented by body to m.
func BodyAddKeywordKindsTo(body Ability, m map[Keyword]bool) {
	for _, ka := range BodyKeywordAbilities(body) {
		m[KeywordAbilityKind(ka)] = true
	}
}

// BodyWardCost returns the Ward cost from a TriggeredAbilityBody's keywords.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards and triggers.
func BodyWardCost(body TriggeredAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Ward)
	if !ok {
		return nil, false
	}
	ward, ok := ka.(WardKeyword)
	if !ok {
		return nil, false
	}
	return ward.Cost, true
}

// BodyMadnessCost returns the Madness cost from a TriggeredAbilityBody's keywords.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards and triggers.
func BodyMadnessCost(body TriggeredAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Madness)
	if !ok {
		return nil, false
	}
	madness, ok := ka.(MadnessKeyword)
	if !ok {
		return nil, false
	}
	return madness.Cost, true
}

// StaticBodyEnchantTarget returns the Enchant target spec for a static ability body.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards and permanents.
func StaticBodyEnchantTarget(body StaticAbility) (TargetSpec, bool) {
	ka, ok := BodyKeywordAbility(body, Enchant)
	if !ok {
		return TargetSpec{}, false
	}
	enchant, ok := ka.(EnchantKeyword)
	if !ok {
		return TargetSpec{}, false
	}
	return enchant.Target, true
}

// ActivatedBodySuspendInfo returns suspend info from an ActivatedAbilityBody.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards.
func ActivatedBodySuspendInfo(body ActivatedAbility) (cost.Mana, int, bool) {
	ka, ok := BodyKeywordAbility(body, Suspend)
	if !ok {
		return nil, 0, false
	}
	suspend, ok := ka.(SuspendKeyword)
	if !ok || suspend.TimeCounters <= 0 {
		return nil, 0, false
	}
	return suspend.Cost, suspend.TimeCounters, true
}

// StaticBodySuspendInfo returns suspend info from a StaticAbilityBody.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards.
func StaticBodySuspendInfo(body StaticAbility) (cost.Mana, int, bool) {
	ka, ok := BodyKeywordAbility(body, Suspend)
	if !ok {
		return nil, 0, false
	}
	suspend, ok := ka.(SuspendKeyword)
	if !ok || suspend.TimeCounters <= 0 {
		return nil, 0, false
	}
	return suspend.Cost, suspend.TimeCounters, true
}

// ActivatedBodyMorphCost returns the morph cost from an ActivatedAbilityBody.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards.
func ActivatedBodyMorphCost(body ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Morph)
	if !ok {
		return nil, false
	}
	morph, ok := ka.(MorphKeyword)
	if !ok {
		return nil, false
	}
	return morph.Cost, true
}

// StaticBodyMorphCost returns the morph cost from a StaticAbilityBody.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards.
func StaticBodyMorphCost(body StaticAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Morph)
	if !ok {
		return nil, false
	}
	morph, ok := ka.(MorphKeyword)
	if !ok {
		return nil, false
	}
	return morph.Cost, true
}

// ActivatedBodyDisguiseCost returns the disguise cost from an ActivatedAbilityBody.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards.
func ActivatedBodyDisguiseCost(body ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Disguise)
	if !ok {
		return nil, false
	}
	disguise, ok := ka.(DisguiseKeyword)
	if !ok {
		return nil, false
	}
	return disguise.Cost, true
}

// StaticBodyDisguiseCost returns the disguise cost from a StaticAbilityBody.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards.
func StaticBodyDisguiseCost(body StaticAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Disguise)
	if !ok {
		return nil, false
	}
	disguise, ok := ka.(DisguiseKeyword)
	if !ok {
		return nil, false
	}
	return disguise.Cost, true
}

// ActivatedBodyEternalize reports whether the body is an eternalize ability.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards.
func ActivatedBodyEternalize(body ActivatedAbility) bool {
	return BodyHasKeyword(body, Eternalize)
}

// ActivatedBodyKicker returns the KickerKeyword from an ActivatedAbilityBody's keywords.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards.
func ActivatedBodyKicker(body ActivatedAbility) (KickerKeyword, bool) {
	ka, ok := BodyKeywordAbility(body, Kicker)
	if !ok {
		return KickerKeyword{}, false
	}
	kicker, ok := ka.(KickerKeyword)
	return kicker, ok
}

// StaticBodyProtectionColors returns the protection colors from a StaticAbilityBody.
//
//nolint:gocritic // Keep value-based API for simple use with body values stored by cards and permanents.
func StaticBodyProtectionColors(body StaticAbility) []color.Color {
	ka, ok := BodyKeywordAbility(body, Protection)
	if !ok {
		return nil
	}
	protection, ok := ka.(ProtectionKeyword)
	if !ok {
		return nil
	}
	return protection.FromColors
}
