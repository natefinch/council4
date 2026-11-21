package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var UnendingWhisper = cardset.Entry{
	OracleID: "c5da5c31-0cf1-4393-8701-eda7880ecebc",
	Name:     "Unending Whisper",
	New:      newUnendingWhisper,
	Verification: "mtg/rules/unending_whisper_test.go: " +
		"TestUnendingWhisperHarmonizeDrawsAndExiles; printing " +
		"fc48180a-ccac-469f-938d-c050821d0160; Scryfall rulings retrieved " +
		"2026-11-21: 7; CR 2026-11-06 118.7a, 121.1, 307.1, 601.2b/f-h, " +
		"601.3, 702.180a-b",
}

func newUnendingWhisper() *game.CardDef {
	const spellText = "Draw a card."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Unending Whisper",
			ManaCost:   opt.Val(cost.Mana{cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Sorcery},
			OracleText: spellText + "\nHarmonize {5}{U} (You may cast this card from your graveyard for its harmonize cost. You may tap a creature you control to reduce that cost by {X}, where X is its power. Then exile this spell.)",
			AlternativeCosts: []cost.Alternative{
				cost.HarmonizeAlternative(cost.Mana{cost.O(5), cost.U}),
			},
			SpellAbility: opt.Val(game.Mode{
				Text: spellText,
				Sequence: []game.Instruction{{Primitive: game.Draw{
					Player: game.ControllerReference(),
					Amount: game.Fixed(1),
				}}},
			}.Ability()),
		},
	}
}
