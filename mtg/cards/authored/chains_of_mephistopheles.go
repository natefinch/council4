package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const chainsOracleID = "eae87919-6322-4bd2-ae9c-b1ce25d686da"
const chainsText = "If a player would draw a card except the first one they draw in each of their draw steps, that player discards a card instead. If the player discards a card this way, they draw a card. If the player doesn't discard a card this way, they mill a card."

var chainsProgram = game.ProgramReference{ID: chainsOracleID + "/replacement/0", Version: "1"}

// ChainsOfMephistopheles composes discard, draw and mill in the shared draw lineage.
var ChainsOfMephistopheles = cardset.Entry{
	OracleID: chainsOracleID, Name: "Chains of Mephistopheles", New: newChainsOfMephistopheles,
	Verification: "mtg/rules/chains_of_mephistopheles_test.go; CR 121, 614.5, 614.11, 616 (2026-09-25)",
	Programs:     []cardprogram.Program{{Reference: chainsProgram, Resolve: replaceChainsDraw}},
}

func newChainsOfMephistopheles() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Chains of Mephistopheles", ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
			Types: []types.Card{types.Enchantment}, Colors: []color.Color{color.Black},
			OracleText: chainsText,
			ReplacementAbilities: []game.ReplacementAbility{{
				Text: chainsText,
				Replacement: game.ReplacementEffect{
					Description: chainsText, MatchEvent: game.EventCardDrawn,
					Duration: game.DurationPermanent, DrawCardExceptFirstInDrawStep: true,
					Program: chainsProgram,
				},
			}},
		},
	}
}

func replaceChainsDraw(ctx cardprogram.Context) error {
	discard, err := ctx.Execute(game.Discard{Player: game.EventPlayerReference(), Amount: game.Fixed(1)})
	if err != nil {
		return err
	}
	if discard.Succeeded {
		_, err = ctx.Execute(game.Draw{Player: game.EventPlayerReference(), Amount: game.Fixed(1)})
	} else {
		_, err = ctx.Execute(game.MoveTopOfLibrary{
			Player: game.EventPlayerReference(), Amount: game.Fixed(1), Destination: zone.Graveyard,
		})
	}
	return err
}
