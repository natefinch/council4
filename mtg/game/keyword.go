package game

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
)

// KeywordAbility is a sealed data-only variant for one printed keyword ability.
type KeywordAbility interface {
	isKeywordAbility()
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

// KeywordAbilityKind returns the Keyword represented by a sealed keyword variant.
func KeywordAbilityKind(ability KeywordAbility) Keyword {
	switch keywordAbility := ability.(type) {
	case SimpleKeyword:
		return keywordAbility.Kind
	case WardKeyword:
		return Ward
	case EquipKeyword:
		return Equip
	case EnchantKeyword:
		return Enchant
	case CyclingKeyword:
		return Cycling
	case KickerKeyword:
		return Kicker
	case MadnessKeyword:
		return Madness
	case MorphKeyword:
		return Morph
	case DisguiseKeyword:
		return Disguise
	case SuspendKeyword:
		return Suspend
	case ProtectionKeyword:
		return Protection
	case nil:
		panic("game: nil KeywordAbility")
	default:
		panic(fmt.Sprintf("game: unsupported KeywordAbility %T", ability))
	}
}

// KeywordKinds returns all keywords represented by this ability, including the
// legacy Keywords field while cards migrate to KeywordAbilities.
func (ability *AbilityDef) KeywordKinds() []Keyword {
	keywords := append([]Keyword(nil), ability.Keywords...)
	for _, keywordAbility := range ability.KeywordAbilities {
		keyword := KeywordAbilityKind(keywordAbility)
		if !slices.Contains(keywords, keyword) {
			keywords = append(keywords, keyword)
		}
	}
	if body, ok := ability.Body.(StaticAbilityBody); ok {
		for _, keywordAbility := range body.KeywordAbilities {
			keyword := KeywordAbilityKind(keywordAbility)
			if !slices.Contains(keywords, keyword) {
				keywords = append(keywords, keyword)
			}
		}
	}
	return keywords
}

// HasKeyword reports whether this ability grants the given keyword.
func (ability *AbilityDef) HasKeyword(keyword Keyword) bool {
	if slices.Contains(ability.Keywords, keyword) {
		return true
	}
	for _, keywordAbility := range ability.KeywordAbilities {
		if KeywordAbilityKind(keywordAbility) == keyword {
			return true
		}
	}
	if body, ok := ability.Body.(StaticAbilityBody); ok {
		for _, keywordAbility := range body.KeywordAbilities {
			if KeywordAbilityKind(keywordAbility) == keyword {
				return true
			}
		}
	}
	return false
}

// AddKeywordKindsTo adds every keyword represented by this ability to keywords.
func (ability *AbilityDef) AddKeywordKindsTo(keywords map[Keyword]bool) {
	for _, keyword := range ability.Keywords {
		keywords[keyword] = true
	}
	for _, keywordAbility := range ability.KeywordAbilities {
		keywords[KeywordAbilityKind(keywordAbility)] = true
	}
	if body, ok := ability.Body.(StaticAbilityBody); ok {
		for _, keywordAbility := range body.KeywordAbilities {
			keywords[KeywordAbilityKind(keywordAbility)] = true
		}
	}
}
