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

const abundanceOracleID = "eccc9a54-2b56-4a02-9926-258d5b2e25fb"
const abundanceText = "If you would draw a card, you may instead choose land or nonland and reveal cards from the top of your library until you reveal a card of the chosen kind. Put that card into your hand and put all other cards revealed this way on the bottom of your library in any order."

var abundanceProgram = game.ProgramReference{ID: abundanceOracleID + "/replacement/0", Version: "1"}

// Abundance replaces individual draws through the Engine's CR 616 loop.
var Abundance = cardset.Entry{
	OracleID: abundanceOracleID, Name: "Abundance", New: newAbundance,
	Verification: "mtg/rules/abundance_test.go: TestAbundanceIndividualDrawsDifferentKindsDeclineAndBottomOrder; TestAbundanceExhaustionEmptyAllMatchingAndDeclinedEmpty",
	Programs:     []cardprogram.Program{{Reference: abundanceProgram, Resolve: replaceAbundanceDraw}},
}

func newAbundance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Abundance", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.G, cost.G}),
			Types: []types.Card{types.Enchantment}, Colors: []color.Color{color.Green},
			OracleText: abundanceText,
			ReplacementAbilities: []game.ReplacementAbility{{
				Text: abundanceText,
				Replacement: game.ReplacementEffect{
					Description: abundanceText, MatchEvent: game.EventCardDrawn,
					ControllerFilter: game.TriggerControllerYou, Duration: game.DurationPermanent,
					Program: abundanceProgram, Optional: true,
				},
			}},
		},
	}
}

func replaceAbundanceDraw(ctx cardprogram.Context) error {
	player, available, err := ctx.View().Player(game.EventPlayerReference())
	if err != nil || !available {
		return err
	}
	choice, err := ctx.Ask(game.ChoiceRequest{
		Kind: game.ChoiceResolution, Player: player, Prompt: "Choose land or nonland.",
		Options:    []game.ChoiceOption{{Index: 0, Label: "land"}, {Index: 1, Label: "nonland"}},
		MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
	})
	if err != nil {
		return err
	}
	selection := game.Selection{RequiredTypes: []types.Card{types.Land}}
	if choice[0] == 1 {
		selection = game.Selection{ExcludedTypes: []types.Card{types.Land}}
	}
	_, err = ctx.Execute(game.RevealUntil{
		Player: game.EventPlayerReference(), Until: selection, Destination: zone.Hand,
		MatchToDestinationRestChosenBottom: true,
	})
	return err
}
