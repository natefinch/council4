package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var GhastlyDiscovery = cardset.Entry{
	OracleID: "1ca875f0-50a7-4332-8643-c2b4bbda18b1",
	Name:     "Ghastly Discovery",
	New:      newGhastlyDiscovery,
	Verification: "mtg/rules/ghastly_discovery_test.go: " +
		"TestGhastlyDiscoveryDrawsThenDiscardsWithConspire; CR 2026-11-06 " +
		"105.2-3, 121.1-2, 307.1, 601.2a-i, 608.2c-e, 701.9a-b, 702.78a, 707.10",
}

func newGhastlyDiscovery() *game.CardDef {
	const text = "Draw two cards, then discard a card."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Ghastly Discovery",
			OracleText: text + "\nConspire (As you cast this spell, you may tap two untapped creatures you control that share a color with it. When you do, copy it.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.U}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Blue},
			StaticAbilities: []game.StaticAbility{
				game.ConspireStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Sequence: []game.Instruction{
					{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
					{Primitive: game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
				},
			}.Ability()),
		},
	}
}
