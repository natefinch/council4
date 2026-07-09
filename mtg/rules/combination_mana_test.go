package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func combinationManaSource(colors []mana.Color, amount int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Combination Source",
		Types: []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{{
			AdditionalCosts: cost.Tap,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{Amount: game.Fixed(amount), CombinationColors: colors},
			}}}.Ability(),
		}},
	}}
}

// TestCombinationManaAbilityHonorsChosenSplit proves the recipient of an "add N
// mana in any combination of <colors>" ability distributes the produced mana
// across the offered colors exactly as chosen, including sending every unit to a
// single color (a zero share for the other colors is legal).
func TestCombinationManaAbilityHonorsChosenSplit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		split []int
		wantR int
		wantG int
	}{
		{"all red", []int{0, 0, 0}, 3, 0},
		{"one red two green", []int{0, 1, 1}, 1, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1,
				combinationManaSource([]mana.Color{mana.R, mana.G}, 3))
			activate := action.ActivateAbility(source.ObjectID, 0, nil, 0)
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{tc.split}},
			}
			if !engine.applyActionWithChoices(g, game.Player1, activate, agents, &TurnLog{}) {
				t.Fatal("combination mana activation failed")
			}
			pool := g.Players[game.Player1].ManaPool
			if got := pool.Amount(mana.R); got != tc.wantR {
				t.Fatalf("red mana = %d, want %d", got, tc.wantR)
			}
			if got := pool.Amount(mana.G); got != tc.wantG {
				t.Fatalf("green mana = %d, want %d", got, tc.wantG)
			}
			if got := pool.Total(); got != 3 {
				t.Fatalf("total mana = %d, want 3", got)
			}
		})
	}
}

// TestCombinationManaAbilityDefaultSplit proves that with no answering agent the
// engine falls back to the deterministic round-robin default, producing the full
// requested amount split across the offered colors.
func TestCombinationManaAbilityDefaultSplit(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1,
		combinationManaSource([]mana.Color{mana.R, mana.G}, 3))
	activate := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, activate) {
		t.Fatal("combination mana activation failed")
	}
	pool := g.Players[game.Player1].ManaPool
	if got := pool.Total(); got != 3 {
		t.Fatalf("total mana = %d, want 3", got)
	}
	// Round-robin over [R, G] for three units is R, G, R.
	if got := pool.Amount(mana.R); got != 2 {
		t.Fatalf("red mana = %d, want 2 (round-robin default)", got)
	}
	if got := pool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana = %d, want 1 (round-robin default)", got)
	}
}

// TestDefaultManaCombination pins the round-robin default and its zero-total and
// zero-color guards so the fallback never panics or over-produces.
func TestDefaultManaCombination(t *testing.T) {
	t.Parallel()
	if got := defaultManaCombination(0, 3); got != nil {
		t.Fatalf("defaultManaCombination(0,3) = %v, want nil", got)
	}
	if got := defaultManaCombination(3, 0); got != nil {
		t.Fatalf("defaultManaCombination(3,0) = %v, want nil", got)
	}
	got := defaultManaCombination(5, 2)
	want := []int{0, 1, 0, 1, 0}
	if len(got) != len(want) {
		t.Fatalf("defaultManaCombination(5,2) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("defaultManaCombination(5,2) = %v, want %v", got, want)
		}
	}
}
