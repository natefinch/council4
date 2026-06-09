package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LoxodonWarhammer is the card definition for Loxodon Warhammer.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature gets +3/+0 and has trample and lifelink.
//	Equip {3}
var LoxodonWarhammer = func() *game.CardDef {
	card := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Loxodon Warhammer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			OracleText: `
				Equipped creature gets +3/+0 and has trample and lifelink.
				Equip {3}
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities, game.StaticAbility{
		Text: `
				Equipped creature gets +3/+0 and has trample and lifelink.
			`,
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:      game.LayerPowerToughnessModify,
				Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
				PowerDelta: 3,
			},
			{
				Layer: game.LayerAbility,
				Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
				AddKeywords: []game.Keyword{
					game.Trample,
					game.Lifelink,
				},
			},
		},
	},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
	)
	return card
}()
