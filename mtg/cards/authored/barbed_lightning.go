package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const barbedLightningOracleID = "f6e1eb52-9f13-425c-8309-1b08ed6557db"

var BarbedLightning = cardset.Entry{
	OracleID: barbedLightningOracleID,
	Name:     "Barbed Lightning",
	New:      newBarbedLightning,
	Verification: "mtg/rules/barbed_lightning_test.go: " +
		"TestBarbedLightningEntwinedCreatureAndPlayerDamage",
}

func newBarbedLightning() *game.CardDef {
	const text = "Choose one —\n• Barbed Lightning deals 3 damage to target creature.\n• Barbed Lightning deals 3 damage to target player or planeswalker.\nEntwine {2} (Choose both if you pay the entwine cost.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Barbed Lightning",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.R}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Red},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{
						Text: "Barbed Lightning deals 3 damage to target creature.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature",
							Allow:      game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.Damage{
								Amount:    game.Fixed(3),
								Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(0)),
							},
						}},
					},
					{
						Text: "Barbed Lightning deals 3 damage to target player or planeswalker.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target player or planeswalker",
							Allow:      game.TargetAllowPlayer | game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Planeswalker},
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.Damage{
								Amount:    game.Fixed(3),
								Recipient: game.AnyTargetDamageRecipient(0),
							},
						}},
					},
				},
				MinModes:     1,
				MaxModes:     2,
				EscalateCost: opt.Val(cost.Mana{cost.O(2)}),
			}),
		},
	}
}
