package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestAvailablePlayerManaCountsTappableSources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1

	addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}}, mana.G, 1)
	addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Sol Ring", Types: []types.Card{types.Artifact}}}, mana.C, 2)
	addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Llanowar Elves",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})}}, mana.G, 1)

	total, colors := availablePlayerMana(g, game.Player1)
	if total != 2 {
		t.Fatalf("availablePlayerMana total = %d, want 2 (summoning-sick dork excluded)", total)
	}
	if !slices.Equal(colors, []string{"G"}) {
		t.Fatalf("availablePlayerMana colors = %v, want [G]", colors)
	}
}

func TestAvailablePlayerManaCountsTappedAndNonSickDork(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1

	tappedLand := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Island", Types: []types.Card{types.Land}}}, mana.U, 1)
	tappedLand.Tapped = true

	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Birds of Paradise",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1})}}, mana.G, 1)
	dork.SummoningSick = false

	total, colors := availablePlayerMana(g, game.Player1)
	if total != 2 {
		t.Fatalf("availablePlayerMana total = %d, want 2 (tapped land + ready dork)", total)
	}
	if !slices.Equal(colors, []string{"U", "G"}) {
		t.Fatalf("availablePlayerMana colors = %v, want [U G]", colors)
	}
}

func TestAvailablePlayerManaExcludesPhasedAndOtherPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1

	phased := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Plains", Types: []types.Card{types.Land}}}, mana.W, 1)
	phased.PhasedOut = true

	addManaAbilityPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Swamp", Types: []types.Card{types.Land}}}, mana.B, 1)

	total, colors := availablePlayerMana(g, game.Player1)
	if total != 0 {
		t.Fatalf("availablePlayerMana total = %d, want 0 (phased out + opponent excluded)", total)
	}
	if colors != nil {
		t.Fatalf("availablePlayerMana colors = %v, want nil", colors)
	}
}

func TestCountLandsPlayed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	actions := []ActionLog{
		{Player: game.Player1, Action: action.PlayLand(g.IDGen.Next())},
		{Player: game.Player1, Action: action.Pass()},
		{Player: game.Player2, Action: action.PlayLand(g.IDGen.Next())},
		{Player: game.Player1, Action: action.PlayLand(g.IDGen.Next())},
	}
	if got := countLandsPlayed(actions, game.Player1); got != 2 {
		t.Fatalf("countLandsPlayed(Player1) = %d, want 2", got)
	}
	if got := countLandsPlayed(actions, game.Player2); got != 1 {
		t.Fatalf("countLandsPlayed(Player2) = %d, want 1", got)
	}
}
