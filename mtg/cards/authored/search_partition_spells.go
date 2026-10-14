package authored

import (
	"fmt"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	intuitionOracleID       = "3c9faba7-f2d3-4978-be94-020dc8003dc0"
	realmsUnchartedOracleID = "e21c8fc6-d4ef-42b6-b11e-d9c931da1387"
)

var (
	intuitionProgram       = game.ProgramReference{ID: intuitionOracleID + "/spell/0", Version: "1"}
	realmsUnchartedProgram = game.ProgramReference{ID: realmsUnchartedOracleID + "/spell/0", Version: "1"}
)

// Intuition requires three unrestricted finds and a targeted opponent's hand choice.
var Intuition = cardset.Entry{
	OracleID: intuitionOracleID, Name: "Intuition", New: newIntuition,
	Programs:     []cardprogram.Program{{Reference: intuitionProgram, Resolve: resolveIntuition}},
	Verification: "mtg/rules/search_partition_spells_test.go: TestAuthoredIntuitionForcedSearchQuantity, TestAuthoredIntuitionCastAndTargetLegality, TestAuthoredSearchPartitionZoneReplacementsAndDeferredShuffle; CR 2026-09-25 701.23d, 608.2b; WotC ruling 2006-07-15",
}

// RealmsUncharted chooses its opponent during resolution, without targeting.
var RealmsUncharted = cardset.Entry{
	OracleID: realmsUnchartedOracleID, Name: "Realms Uncharted", New: newRealmsUncharted,
	Programs:     []cardprogram.Program{{Reference: realmsUnchartedProgram, Resolve: resolveRealmsUncharted}},
	Verification: "mtg/rules/search_partition_spells_test.go: TestAuthoredRealmsUnchartedDifferentNamesAndFewerFinds, TestAuthoredRealmsUnchartedLibraryExhaustionAndLivingOpponent, TestAuthoredSearchPartitionOppositionAgent; CR 2026-09-25 201.2b, 701.23b, 608.2d; WotC ruling 2010-06-15",
}

func newIntuition() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Intuition", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			OracleText: "Search your library for three cards and reveal them. Target opponent chooses one. Put that card into your hand and the rest into your graveyard. Then shuffle.",
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer, Constraint: "opponent"}},
				Program: intuitionProgram,
			}.Ability()),
		},
	}
}

func resolveIntuition(ctx cardprogram.Context) error {
	opponent, available, err := ctx.View().Player(game.TargetPlayerReference(0))
	if err != nil || !available {
		return err
	}
	found, err := ctx.Search(game.Search{
		Player: game.ControllerReference(), Amount: game.Fixed(3),
		Spec: game.SearchSpec{
			SourceZone: zone.Library, Reveal: true, RevealOnly: true,
			FailToFindPolicy: game.SearchMustFindIfAvailable,
		},
	})
	if err != nil {
		return err
	}
	var hand, graveyard []cardprogram.FoundCard
	if len(found.Cards) > 0 {
		selected, err := ctx.Ask(game.ChoiceRequest{
			Kind: game.ChoiceZoneSelection, Player: opponent,
			Prompt:  "Choose the revealed card to put into the searching player's hand.",
			Options: foundCardOptions(found.Cards), MinChoices: 1, MaxChoices: 1,
		})
		if err != nil {
			return err
		}
		for index, card := range found.Cards {
			if index == selected[0] {
				hand = append(hand, card)
			} else {
				graveyard = append(graveyard, card)
			}
		}
	}
	if _, err := ctx.MoveCards(
		cardprogram.CardPlacement{Cards: hand, Destination: zone.Hand},
		cardprogram.CardPlacement{Cards: graveyard, Destination: zone.Graveyard},
	); err != nil {
		return err
	}
	_, err = ctx.Execute(game.ShuffleLibrary{Player: game.ControllerReference()})
	return err
}

func newRealmsUncharted() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Realms Uncharted", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.G}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Green},
			OracleText:   "Search your library for up to four land cards with different names and reveal them. An opponent chooses two of those cards. Put the chosen cards into your graveyard and the rest into your hand. Then shuffle.",
			SpellAbility: opt.Val(game.Mode{Program: realmsUnchartedProgram}.Ability()),
		},
	}
}

func resolveRealmsUncharted(ctx cardprogram.Context) error {
	found, err := ctx.Search(game.Search{
		Player: game.ControllerReference(), Amount: game.Fixed(4),
		Spec: game.SearchSpec{
			SourceZone: zone.Library, Reveal: true, RevealOnly: true,
			Filter:         game.Selection{RequiredTypes: []types.Card{types.Land}},
			DifferentNames: true, FailToFindPolicy: game.SearchMayFailToFind,
		},
	})
	if err != nil {
		return err
	}
	chosen := make(map[int]bool)
	if len(found.Cards) > 0 {
		opponents, err := ctx.View().Players(game.OpponentsReference())
		if err != nil {
			return err
		}
		if len(opponents) > 0 {
			options := make([]game.ChoiceOption, len(opponents))
			for index, opponent := range opponents {
				options[index] = game.ChoiceOption{Index: index, Label: fmt.Sprintf("Player %d", opponent+1)}
			}
			selected, err := ctx.Ask(game.ChoiceRequest{
				Kind: game.ChoicePlayer, Player: ctx.View().Source().Controller,
				Prompt:  "Choose an opponent to choose the revealed lands for your graveyard.",
				Options: options, MinChoices: 1, MaxChoices: 1,
			})
			if err != nil {
				return err
			}
			count := min(2, len(found.Cards))
			selected, err = ctx.Ask(game.ChoiceRequest{
				Kind: game.ChoiceZoneSelection, Player: opponents[selected[0]],
				Prompt:  "Choose the revealed lands to put into the searching player's graveyard.",
				Options: foundCardOptions(found.Cards), MinChoices: count, MaxChoices: count,
			})
			if err != nil {
				return err
			}
			for _, index := range selected {
				chosen[index] = true
			}
		}
	}
	var graveyard, hand []cardprogram.FoundCard
	for index, card := range found.Cards {
		if chosen[index] {
			graveyard = append(graveyard, card)
		} else {
			hand = append(hand, card)
		}
	}
	if _, err := ctx.MoveCards(
		cardprogram.CardPlacement{Cards: graveyard, Destination: zone.Graveyard},
		cardprogram.CardPlacement{Cards: hand, Destination: zone.Hand},
	); err != nil {
		return err
	}
	_, err = ctx.Execute(game.ShuffleLibrary{Player: game.ControllerReference()})
	return err
}

func foundCardOptions(cards []cardprogram.FoundCard) []game.ChoiceOption {
	options := make([]game.ChoiceOption, len(cards))
	for index, card := range cards {
		options[index] = game.ChoiceOption{Index: index, Label: card.Info.Name, Card: opt.Val(card.Info)}
	}
	return options
}
