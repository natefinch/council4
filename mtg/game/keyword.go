package game

import (
	"slices"

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

// KickerKeyword parameterizes Kicker additional costs and bonus effects.
type KickerKeyword struct {
	Cost  cost.Mana
	Bonus []Effect
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

// KeywordKinds returns all keywords represented by this ability.
func (ability *AbilityDef) KeywordKinds() []Keyword {
	var keywords []Keyword
	for _, keywordAbility := range ability.KeywordAbilities {
		keyword := KeywordAbilityKind(keywordAbility)
		if !slices.Contains(keywords, keyword) {
			keywords = append(keywords, keyword)
		}
	}
	addFromKeywords := func(kas []KeywordAbility) {
		for _, keywordAbility := range kas {
			keyword := KeywordAbilityKind(keywordAbility)
			if !slices.Contains(keywords, keyword) {
				keywords = append(keywords, keyword)
			}
		}
	}
	switch body := ability.Body.(type) {
	case StaticAbilityBody:
		addFromKeywords(body.KeywordAbilities)
	case ActivatedAbilityBody:
		addFromKeywords(body.KeywordAbilities)
	default:
	}
	return keywords
}

// HasKeyword reports whether this ability grants the given keyword.
func (ability *AbilityDef) HasKeyword(keyword Keyword) bool {
	if _, ok := ability.KeywordAbility(keyword); ok {
		return true
	}
	return false
}

// KeywordAbility returns the first sealed keyword variant matching keyword.
func (ability *AbilityDef) KeywordAbility(keyword Keyword) (KeywordAbility, bool) {
	for _, keywordAbility := range ability.KeywordAbilities {
		if KeywordAbilityKind(keywordAbility) == keyword {
			return keywordAbility, true
		}
	}
	var bodyKeywords []KeywordAbility
	switch body := ability.Body.(type) {
	case StaticAbilityBody:
		bodyKeywords = body.KeywordAbilities
	case ActivatedAbilityBody:
		bodyKeywords = body.KeywordAbilities
	default:
	}
	for _, keywordAbility := range bodyKeywords {
		if KeywordAbilityKind(keywordAbility) == keyword {
			return keywordAbility, true
		}
	}
	return nil, false
}

// EnchantTarget returns the target restriction for an Enchant keyword ability.
func (ability *AbilityDef) EnchantTarget() (TargetSpec, bool) {
	keyword, ok := ability.KeywordAbility(Enchant)
	if !ok {
		return TargetSpec{}, false
	}
	enchant, ok := keyword.(EnchantKeyword)
	if !ok {
		return TargetSpec{}, false
	}
	return enchant.Target, true
}

// WardCost returns the mana-valued Ward cost for this ability.
func (ability *AbilityDef) WardCost() (cost.Mana, bool) {
	keyword, ok := ability.KeywordAbility(Ward)
	if !ok {
		return nil, false
	}
	ward, ok := keyword.(WardKeyword)
	if !ok {
		return nil, false
	}
	return ward.Cost, true
}

// MadnessCost returns the mana-valued Madness cost for this ability.
func (ability *AbilityDef) MadnessCost() (cost.Mana, bool) {
	keyword, ok := ability.KeywordAbility(Madness)
	if !ok {
		return nil, false
	}
	madness, ok := keyword.(MadnessKeyword)
	if !ok {
		return nil, false
	}
	return madness.Cost, true
}

// SuspendInfo returns the Suspend cost and time counters for this ability.
func (ability *AbilityDef) SuspendInfo() (cost.Mana, int, bool) {
	keyword, ok := ability.KeywordAbility(Suspend)
	if !ok {
		return nil, 0, false
	}
	suspend, ok := keyword.(SuspendKeyword)
	if !ok || suspend.TimeCounters <= 0 {
		return nil, 0, false
	}
	return suspend.Cost, suspend.TimeCounters, true
}

// MorphCost returns the turn-face-up cost for a Morph keyword ability.
func (ability *AbilityDef) MorphCost() (cost.Mana, bool) {
	keyword, ok := ability.KeywordAbility(Morph)
	if !ok {
		return nil, false
	}
	morph, ok := keyword.(MorphKeyword)
	if !ok {
		return nil, false
	}
	return morph.Cost, true
}

// DisguiseCost returns the turn-face-up cost for a Disguise keyword ability.
func (ability *AbilityDef) DisguiseCost() (cost.Mana, bool) {
	keyword, ok := ability.KeywordAbility(Disguise)
	if !ok {
		return nil, false
	}
	disguise, ok := keyword.(DisguiseKeyword)
	if !ok {
		return nil, false
	}
	return disguise.Cost, true
}

// ProtectionColors returns the colors this ability grants protection from.
func (ability *AbilityDef) ProtectionColors() []color.Color {
	keyword, ok := ability.KeywordAbility(Protection)
	if !ok {
		return nil
	}
	protection, ok := keyword.(ProtectionKeyword)
	if !ok {
		return nil
	}
	return protection.FromColors
}

// Kicker returns the Kicker keyword variant for this ability.
func (ability *AbilityDef) Kicker() (KickerKeyword, bool) {
	keyword, ok := ability.KeywordAbility(Kicker)
	if !ok {
		return KickerKeyword{}, false
	}
	kicker, ok := keyword.(KickerKeyword)
	if !ok {
		return KickerKeyword{}, false
	}
	return kicker, true
}

// KickerCost returns the optional additional mana cost for Kicker.
func (ability *AbilityDef) KickerCost() (cost.Mana, bool) {
	kicker, ok := ability.Kicker()
	if !ok {
		return nil, false
	}
	return kicker.Cost, true
}

// KickerBonusEffects returns effects applied if this spell was kicked.
func (ability *AbilityDef) KickerBonusEffects() []Effect {
	kicker, ok := ability.Kicker()
	if !ok {
		return nil
	}
	return kicker.Bonus
}

// AddKeywordKindsTo adds every keyword represented by this ability to keywords.
func (ability *AbilityDef) AddKeywordKindsTo(keywords map[Keyword]bool) {
	for _, keywordAbility := range ability.KeywordAbilities {
		keywords[KeywordAbilityKind(keywordAbility)] = true
	}
	var bodyKeywords []KeywordAbility
	switch body := ability.Body.(type) {
	case StaticAbilityBody:
		bodyKeywords = body.KeywordAbilities
	case ActivatedAbilityBody:
		bodyKeywords = body.KeywordAbilities
	default:
	}
	for _, keywordAbility := range bodyKeywords {
		keywords[KeywordAbilityKind(keywordAbility)] = true
	}
}
