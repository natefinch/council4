package authored

import (
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const necromanticSelectionOracleID = "b48a149c-aead-43b2-a99e-c964d6d28aa0"

var necromanticSelectionSpell = game.ProgramReference{
	ID: necromanticSelectionOracleID + "/spell/0", Version: "1",
}

var NecromanticSelection = cardset.Entry{
	OracleID: necromanticSelectionOracleID, Name: "Necromantic Selection", New: newNecromanticSelection,
	Verification: "mtg/rules/necromantic_selection_test.go: TestNecromanticSelectionCheckedAdmissionAndChoices; TestNecromanticSelectionNoEligibleArrival; TestNecromanticSelectionExcludesRegeneratedCreatureAndKeepsEntryEffects; TDC 189 (2025-04-11), printing 2002d779-9ea2-432b-8c15-d61995b8adda; Scryfall rulings 2014-11-07 (2); CR 2026-09-25 105.3, 400.6-7, 608.2c-f, 611.2a/e, 613.1d/e, 701.8",
	Programs: []cardprogram.Program{{
		Reference: necromanticSelectionSpell, Resolve: resolveNecromanticSelection,
	}},
}

func newNecromanticSelection() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Necromantic Selection", ManaCost: opt.Val(cost.Mana{cost.O(4), cost.B, cost.B, cost.B}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Black},
			OracleText:   "Destroy all creatures, then return a creature card put into a graveyard this way to the battlefield under your control. It's a black Zombie in addition to its other colors and types. Exile Necromantic Selection.",
			SpellAbility: opt.Val(game.Mode{Program: necromanticSelectionSpell}.Ability()),
		},
	}
}

func resolveNecromanticSelection(ctx cardprogram.Context) error {
	destruction, err := ctx.Destroy(game.Destroy{Group: game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
	})})
	if err != nil {
		return err
	}

	var eligible []cardprogram.FoundCard
	for _, member := range destruction.Members {
		if !member.Moved || member.Destination != zone.Graveyard {
			continue
		}
		for _, card := range member.Cards {
			if slices.Contains(card.Info.Types, types.Creature) {
				eligible = append(eligible, card)
			}
		}
	}

	if len(eligible) > 0 {
		options := make([]game.ChoiceOption, len(eligible))
		for i, card := range eligible {
			options[i] = game.ChoiceOption{Index: i, Label: card.Info.Name, Card: opt.Val(card.Info)}
		}
		selected, err := ctx.Ask(game.ChoiceRequest{
			Kind: game.ChoiceResolution, Player: ctx.View().Source().Controller,
			Prompt:  "Choose a creature card put into a graveyard this way to return to the battlefield.",
			Options: options, MinChoices: 1, MaxChoices: 1,
		})
		if err != nil {
			return err
		}
		if _, err := ctx.MoveCards(cardprogram.CardPlacement{
			Cards: []cardprogram.FoundCard{eligible[selected[0]]}, Destination: zone.Battlefield,
			Recipient: opt.Val(game.ControllerReference()),
			ContinuousEffects: []game.ContinuousEffect{
				{Layer: game.LayerColor, AddColors: []color.Color{color.Black}},
				{Layer: game.LayerType, AddSubtypes: []types.Sub{types.Zombie}},
			},
		}); err != nil {
			return err
		}
	}

	_, err = ctx.Execute(game.MoveResolvingSpell{Destination: zone.Exile})
	return err
}
