package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CataclysmicGearhulk is the card definition for Cataclysmic Gearhulk.
//
// Type: Artifact Creature — Construct
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Vigilance
//	When this creature enters, each player chooses an artifact, a creature, an enchantment, and a planeswalker from among the nonland permanents they control, then sacrifices the rest.
var CataclysmicGearhulk = newCataclysmicGearhulk

func newCataclysmicGearhulk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Cataclysmic Gearhulk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.KeepOnePerType{
									Players:           game.AllPlayersReference(),
									Types:             []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Planeswalker},
									AffectedSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance
			When this creature enters, each player chooses an artifact, a creature, an enchantment, and a planeswalker from among the nonland permanents they control, then sacrifices the rest.
		`,
		},
	}
}
