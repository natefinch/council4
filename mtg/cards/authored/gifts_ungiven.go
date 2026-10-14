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

const giftsUngivenOracleID = "58aec411-167d-4709-8560-793eaaed62c5"

var giftsUngivenSearchProgram = game.ProgramReference{ID: giftsUngivenOracleID + "/spell/0", Version: "1"}

// GiftsUngiven keeps the search and opponent's partition as separate Go decisions.
var GiftsUngiven = cardset.Entry{
	OracleID: giftsUngivenOracleID, Name: "Gifts Ungiven", New: newGiftsUngiven,
	Programs: []cardprogram.Program{{Reference: giftsUngivenSearchProgram, Resolve: resolveGiftsUngivenSearch}},
}

func newGiftsUngiven() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Gifts Ungiven", ManaCost: opt.Val(cost.Mana{cost.O(3), cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			OracleText: "Search your library for up to four cards with different names and reveal them. Target opponent chooses two of those cards. Put the chosen cards into your graveyard and the rest into your hand. Then shuffle.",
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer, Constraint: "opponent"}},
				Program: giftsUngivenSearchProgram,
			}.Ability()),
		},
	}
}

func resolveGiftsUngivenSearch(ctx cardprogram.Context) error {
	opponent, available, err := ctx.View().Player(game.TargetPlayerReference(0))
	if err != nil || !available {
		return err
	}
	found, err := ctx.Search(game.Search{
		Player: game.ControllerReference(), Amount: game.Fixed(4),
		Spec: game.SearchSpec{
			SourceZone: zone.Library, Destination: zone.None, Reveal: true, RevealOnly: true,
			DifferentNames: true, FailToFindPolicy: game.SearchMayFailToFind,
		},
	})
	if err != nil {
		return err
	}
	chosen := make(map[int]bool)
	if len(found.Cards) > 0 {
		options := make([]game.ChoiceOption, len(found.Cards))
		for index, card := range found.Cards {
			options[index] = game.ChoiceOption{Index: index, Label: card.Info.Name, Card: opt.Val(card.Info)}
		}
		count := min(2, len(found.Cards))
		selected, err := ctx.Ask(game.ChoiceRequest{
			Kind: game.ChoiceZoneSelection, Player: opponent,
			Prompt:  "Choose the revealed cards to put into the searching player's graveyard.",
			Options: options, MinChoices: count, MaxChoices: count,
		})
		if err != nil {
			return err
		}
		for _, index := range selected {
			chosen[index] = true
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
