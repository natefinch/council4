package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BasiliskCollar is the card definition for Basilisk Collar.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature has deathtouch and lifelink.
//	Equip {2}
var BasiliskCollar = func() *game.CardDef {
	card := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Basilisk Collar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			OracleText: `
				Equipped creature has deathtouch and lifelink. (Any amount of damage it deals to a creature is enough to destroy it. Damage dealt by this creature also causes you to gain that much life.)
				Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities, game.StaticAbility{
		Text: `
				Equipped creature has deathtouch and lifelink.
			`,
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer: game.LayerAbility,
				Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
				AddKeywords: []game.Keyword{
					game.Deathtouch,
					game.Lifelink,
				},
			},
		},
	},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
	)
	return card
}
