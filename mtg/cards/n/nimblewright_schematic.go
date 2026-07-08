package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NimblewrightSchematic is the card definition for Nimblewright Schematic.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	When this artifact enters or is put into a graveyard from the battlefield, create a 1/1 colorless Construct artifact creature token.
var NimblewrightSchematic = newNimblewrightSchematic

func newNimblewrightSchematic() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Nimblewright Schematic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventPermanentDied,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(nimblewrightSchematicToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this artifact enters or is put into a graveyard from the battlefield, create a 1/1 colorless Construct artifact creature token.
		`,
		},
	}
}

var nimblewrightSchematicToken = newNimblewrightSchematicToken()

func newNimblewrightSchematicToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Construct",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
