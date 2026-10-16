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

var warpWorldProgram = game.ProgramReference{ID: "50a228a2-c8b6-4416-b0ab-417926a9b9b6/spell/0", Version: "1"}

// WarpWorld uses two cross-player entry stages over exact revealed card handles.
var WarpWorld = cardset.Entry{
	OracleID: "50a228a2-c8b6-4416-b0ab-417926a9b9b6", Name: "Warp World", New: newWarpWorld,
	Programs:     []cardprogram.Program{{Reference: warpWorldProgram, Resolve: resolveWarpWorld}},
	Verification: "mtg/rules/warp_world_test.go; WotC 2006-10-15, 2009-10-01, 2019-05-03; CR 2026-09-25 101.4, 303.4, 400.7, 603.3, 614.12, 701.24; #3288",
}

func newWarpWorld() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Warp World",
			OracleText: "Each player shuffles all permanents they own into their library, then reveals that many cards from the top of their library. Each player puts all artifact, creature, and land cards revealed this way onto the battlefield, then does the same for enchantment cards, then puts all cards revealed this way that weren't put onto the battlefield on the bottom of their library.",
			ManaCost:   opt.Val(cost.Mana{cost.O(5), cost.R, cost.R, cost.R}),
			Types:      []types.Card{types.Sorcery}, Colors: []color.Color{color.Red},
			SpellAbility: opt.Val(game.Mode{Text: "Resolve Warp World", Program: warpWorldProgram}.Ability()),
		},
	}
}

func resolveWarpWorld(ctx cardprogram.Context) error {
	group := game.BattlefieldGroup(game.Selection{})
	before, err := ctx.View().Permanents(group)
	if err != nil {
		return err
	}
	players, err := ctx.View().Players(game.AllPlayersReference())
	if err != nil {
		return err
	}
	var counts [game.NumPlayers]int
	for _, permanent := range before {
		counts[permanent.Owner]++
	}
	if _, err := ctx.Execute(game.MovePermanent{Group: group, Destination: zone.Library}); err != nil {
		return err
	}
	for _, player := range players {
		if _, err := ctx.Execute(game.ShuffleLibrary{Player: game.CapturedPlayerReference(player)}); err != nil {
			return err
		}
	}
	var revealed, first, second []cardprogram.FoundCard
	for _, player := range players {
		if counts[player] == 0 {
			continue
		}
		result, err := ctx.Execute(game.Reveal{Player: game.CapturedPlayerReference(player), Amount: game.Fixed(counts[player])})
		if err != nil {
			return err
		}
		revealed = append(revealed, result.Cards...)
		for _, card := range result.Cards {
			if slices.Contains(card.Info.Types, types.Artifact) || slices.Contains(card.Info.Types, types.Creature) ||
				slices.Contains(card.Info.Types, types.Land) {
				first = append(first, card)
			} else if slices.Contains(card.Info.Types, types.Enchantment) {
				second = append(second, card)
			}
		}
	}
	if _, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: first, Destination: zone.Battlefield}); err != nil {
		return err
	}
	if _, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: second, Destination: zone.Battlefield}); err != nil {
		return err
	}
	_, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed, Destination: zone.Library, LibraryBottom: true})
	return err
}
