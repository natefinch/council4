package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	skycloudEggOracleID = "1c4b6543-777e-4c3b-a9fb-5b7210d458d5"
	skycloudEggText     = "{2}, {T}, Sacrifice this artifact: Add {W}{U}. Draw a card. (Activate only as an instant.)"
)

var SkycloudEgg = cardset.Entry{
	OracleID: skycloudEggOracleID,
	Name:     "Skycloud Egg",
	New:      newSkycloudEgg,
	Verification: "mtg/rules/skycloud_egg_test.go: TestSkycloudEggPaysCostsThenAddsManaAndDraws; " +
		"Odyssey printing e588cc41-4b86-4e93-94e8-dd765b27ab63; Scryfall ruling 2008-08-01; " +
		"CR 2026-11-06 113.3b, 118.1, 121.1, 602.1b, 602.2a-b, 602.5e, 605.1a, 605.3b, 608.2c",
}

func newSkycloudEgg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.White),
		CardFace: game.CardFace{
			Name:       "Skycloud Egg",
			ManaCost:   opt.Val(cost.Mana{cost.O(1)}),
			Types:      []types.Card{types.Artifact},
			OracleText: skycloudEggText,
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:     skycloudEggText,
				ManaCost: opt.Val(cost.Mana{cost.O(2)}),
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalTap},
					{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this artifact", Amount: 1},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.W}},
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.U}},
					{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
				}}.Ability(),
			}},
		},
	}
}
