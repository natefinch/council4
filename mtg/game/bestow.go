package game

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BestowSpellTypes returns the effective card types of a spell that was cast for
// its bestow cost (CR 702.103b). While on the stack such a spell is an Aura
// enchantment spell and is not a creature spell, so the Creature type is removed;
// all other printed types are kept in their printed order. The input slice is not
// modified.
func BestowSpellTypes(printed []types.Card) []types.Card {
	out := make([]types.Card, 0, len(printed))
	for _, cardType := range printed {
		if cardType == types.Creature {
			continue
		}
		out = append(out, cardType)
	}
	return out
}

// BestowSpellSubtypes returns the effective subtypes of a spell cast for its
// bestow cost (CR 702.103b): it becomes an Aura spell, so the Aura subtype is
// added. Printed subtypes are preserved and Aura is appended only when it is not
// already present. The input slice is not modified.
func BestowSpellSubtypes(printed []types.Sub) []types.Sub {
	out := append([]types.Sub(nil), printed...)
	if slices.Contains(out, types.Aura) {
		return out
	}
	return append(out, types.Aura)
}

// BestowStaticAbility builds the static ability that carries the Bestow keyword
// (CR 702.103) for an enchantment creature card. The ability serves two roles:
//
//   - It carries the BestowKeyword (fixed bestow mana cost and the enchant
//     creature target spec), so HasKeyword(Bestow) reports true and the rules
//     layer can offer the bestow alternative cost, require the enchant target on
//     a bestowed cast, and check attachment legality. The card keeps neither the
//     Aura subtype nor the Enchant keyword, so it stays an ordinary enchantment
//     creature that is cast without a target and is not an Aura card.
//
//   - It carries a self type-change continuous effect (remove the creature type,
//     add the Aura subtype) at the type layer, gated on the source being
//     bestowed. While the permanent is a bestowed Aura the effect makes it an
//     Aura and not a creature; when it ceases to be bestowed — because it became
//     unattached or was attached to an illegal object (CR 702.103e–g) — the gate
//     closes and it is a creature again. The gate reads the raw Permanent.Bestowed
//     flag, so it never depends on the type it defines.
func BestowStaticAbility(bestowCost cost.Mana, target *TargetSpec) StaticAbility {
	targetCopy := cloneTargetSpec(target)
	return StaticAbility{
		Text:             "Bestow " + bestowCost.String(),
		KeywordAbilities: []KeywordAbility{BestowKeyword{Cost: append(cost.Mana(nil), bestowCost...), Target: targetCopy}},
		Condition: opt.Val(Condition{
			Text:           "if this permanent is bestowed",
			SourceBestowed: true,
		}),
		ContinuousEffects: []ContinuousEffect{{
			Layer:          LayerType,
			AffectedSource: true,
			RemoveTypes:    []types.Card{types.Creature},
			AddSubtypes:    []types.Sub{types.Aura},
		}},
	}
}
