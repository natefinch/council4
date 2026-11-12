package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const manifestDreadOracleID = "df5a66d1-61f6-44d9-b464-3ee739e78dce"
const manifestDreadOracleText = "Manifest dread. (Look at the top two cards of your library. Put one onto the battlefield face down as a 2/2 creature and the other into your graveyard. Turn it face up any time for its mana cost if it's a creature card.)"

var ManifestDread = cardset.Entry{
	OracleID: manifestDreadOracleID,
	Name:     "Manifest Dread",
	New:      newManifestDread,
	Verification: "mtg/rules/authored_existing_paragraph_cards_test.go: " +
		"TestManifestDreadScenario",
}

func newManifestDread() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:     "Manifest Dread",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Manifest{Dread: true},
				}},
			}.Ability()),
			OracleText: manifestDreadOracleText,
		},
	}
}
