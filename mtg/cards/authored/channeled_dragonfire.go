package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var ChanneledDragonfire = cardset.Entry{
	OracleID: "91179937-6f55-4d1b-8f06-2efe1c8fea7a",
	Name:     "Channeled Dragonfire",
	New:      newChanneledDragonfire,
	Verification: "mtg/rules/channeled_dragonfire_test.go: " +
		"TestChanneledDragonfireHarmonizeDealsTwoDamage; printing " +
		"24204881-690c-4043-8771-20cb93385072; Scryfall rulings retrieved " +
		"2026-11-21: 7; CR 2026-11-06 115.1a, 115.4, 118.7a, 120.1-3, " +
		"307.1, 601.2a-i, 601.3, 608.2b, 702.180a-b",
}

func newChanneledDragonfire() *game.CardDef {
	const spellText = "Channeled Dragonfire deals 2 damage to any target."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Channeled Dragonfire",
			ManaCost:   opt.Val(cost.Mana{cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Sorcery},
			OracleText: spellText + "\nHarmonize {5}{R}{R} (You may cast this card from your graveyard for its harmonize cost. You may tap a creature you control to reduce that cost by {X}, where X is its power. Then exile this spell.)",
			AlternativeCosts: []cost.Alternative{
				cost.HarmonizeAlternative(cost.Mana{cost.O(5), cost.R, cost.R}),
			},
			SpellAbility: opt.Val(game.Mode{
				Text: spellText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "any target",
					Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle},
					}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Damage{
					Amount:    game.Fixed(2),
					Recipient: game.AnyTargetDamageRecipient(0),
				}}},
			}.Ability()),
		},
	}
}
