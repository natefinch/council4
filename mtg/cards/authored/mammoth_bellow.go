package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var MammothBellow = cardset.Entry{
	OracleID: "6f2e4607-bece-42ba-a2a3-5dc50f246793",
	Name:     "Mammoth Bellow",
	New:      newMammothBellow,
	Verification: "mtg/rules/mammoth_bellow_test.go: TestMammothBellowHarmonizeCreatesElephant; " +
		"printing 468b17b4-79ce-4dfa-8873-a9cfc347e38f; Scryfall rulings retrieved " +
		"2026-11-21: 7; CR 2026-11-06 111.1-4, 118.7a, 307.1, 601.2a-i, 601.3, " +
		"608.2c, 702.180a-b",
}

func newMammothBellow() *game.CardDef {
	const spellText = "Create a 5/5 green Elephant creature token."
	token := &game.CardDef{CardFace: game.CardFace{
		Name: "Elephant Token", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elephant},
		Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 5}), Toughness: opt.Val(game.PT{Value: 5}),
	}}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:       "Mammoth Bellow",
			OracleText: spellText + "\nHarmonize {5}{G}{U}{R} (You may cast this card from your graveyard for its harmonize cost. You may tap a creature you control to reduce that cost by {X}, where X is its power. Then exile this spell.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.G, cost.U, cost.R}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Green, color.Blue, color.Red},
			AlternativeCosts: []cost.Alternative{
				cost.HarmonizeAlternative(cost.Mana{cost.O(5), cost.G, cost.U, cost.R}),
			},
			SpellAbility: opt.Val(game.Mode{
				Text:     spellText,
				Sequence: []game.Instruction{{Primitive: game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(token)}}},
			}.Ability()),
		},
	}
}
