package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	tokenh "github.com/natefinch/council4/mtg/cards/tokens/h"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const konstrariImproviserOracleID = "bad2e3b2-ad47-4e17-b35e-0fcc13e82937"
const konstrariImproviserOracleText = "This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)"
const soulTetherOracleText = "Create a Heartwood token. (It's a red and green artifact with \"{T}: Add {R} or {G}.\")"

var KonstrariImproviserSoulTether = cardset.Entry{
	OracleID: konstrariImproviserOracleID,
	Name:     "Konstrari Improviser // Soul Tether",
	New:      newKonstrariImproviserSoulTether,
	Verification: "mtg/rules/authored_existing_paragraph_cards_test.go: " +
		"TestKonstrariImproviserSoulTetherScenario",
}

func newKonstrariImproviserSoulTether() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:           "Konstrari Improviser",
			ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.HybridMana(mana.R, mana.G)}),
			Colors:         []color.Color{color.Red, color.Green},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Human, types.Artificer},
			Power:          opt.Val(game.PT{Value: 2}),
			Toughness:      opt.Val(game.PT{Value: 2}),
			OracleText:     konstrariImproviserOracleText,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name:     "Soul Tether",
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.HybridMana(mana.R, mana.G)}),
			Colors:   []color.Color{color.Red, color.Green},
			Types:    []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.CreateToken{
						Amount:    game.Fixed(1),
						Source:    game.TokenDef(tokenh.HeartwoodTokened1d8118b182413b99ba52844d35a35b),
						Recipient: opt.Val(game.ControllerReference()),
					},
				}},
			}.Ability()),
			OracleText: soulTetherOracleText,
		}),
	}
}
