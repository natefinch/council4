package rules

import (
	"math/rand/v2"
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

func TestManaDevelopmentReflectsGoldfishBoard(t *testing.T) {
	commander := &game.CardDef{CardFace: game.CardFace{
		Name:       "Goldfish Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
	}}
	forest := &game.CardDef{CardFace: game.CardFace{
		Name:          "Forest",
		Supertypes:    []types.Super{types.Basic},
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.G)},
	}}
	config := game.PlayerConfig{Name: "Goldfish", Commander: commander, Deck: repeatedCard(forest, 99)}

	engine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	g := engine.NewGoldfishGame(config)
	result := engine.RunGoldfish(g, goldfishTestAgent{}, 10)

	cumulativeLands := 0
	for _, turn := range result.Turns {
		cumulativeLands += turn.LandsPlayed
		if turn.ManaAvailable != cumulativeLands {
			t.Fatalf("turn %d: ManaAvailable = %d, want %d (every Forest is an untapped mana source)",
				turn.TurnNumber, turn.ManaAvailable, cumulativeLands)
		}
		if cumulativeLands > 0 && !slices.Equal(turn.ManaColors, []string{"G"}) {
			t.Fatalf("turn %d: ManaColors = %v, want [G]", turn.TurnNumber, turn.ManaColors)
		}
	}
	if cumulativeLands == 0 {
		t.Fatal("goldfish never played a land; cannot validate mana development")
	}
}
