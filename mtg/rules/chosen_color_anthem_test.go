package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func addColoredCreaturePermanent(g *game.Game, controller game.PlayerID, power int, colors ...color.Color) *game.Permanent {
	pt := game.PT{Value: power}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Colored Creature",
		Types:     []types.Card{types.Creature},
		Colors:    colors,
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
}

// TestChosenColorAnthemBuffsOnlyChosenColorControlledCreatures verifies the
// "Creatures you control of the chosen color get +1/+0" anthem (Heraldic
// Banner): only the source controller's creatures sharing the source's
// entry-time color choice receive the bonus.
func TestChosenColorAnthemBuffsOnlyChosenColorControlledCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	anchor := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Heraldic Banner",
		Types: []types.Card{types.Artifact},
	}})
	anchor.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: mana.R},
	}

	chosenColorAlly := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	wrongColorAlly := addColoredCreaturePermanent(g, game.Player1, 2, color.White)
	chosenColorOpponent := addColoredCreaturePermanent(g, game.Player2, 2, color.Red)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: anchor.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			ColorChoice:   game.ColorChoiceSourceEntry,
		}),
		PowerDelta: 1,
	})

	if got := effectivePower(g, chosenColorAlly); got != 3 {
		t.Fatalf("controlled chosen-color creature power = %d, want 3", got)
	}
	if got := effectivePower(g, wrongColorAlly); got != 2 {
		t.Fatalf("controlled wrong-color creature power = %d, want 2 (unbuffed)", got)
	}
	if got := effectivePower(g, chosenColorOpponent); got != 2 {
		t.Fatalf("opponent chosen-color creature power = %d, want 2 (unbuffed)", got)
	}
}
