package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NemesisMask is the card definition for Nemesis Mask.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	All creatures able to block equipped creature do so.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var NemesisMask = newNemesisMask

func newNemesisMask() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Nemesis Mask",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All creatures able to block equipped creature do so.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectMustBeBlockedByAllAble,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			All creatures able to block equipped creature do so.
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
