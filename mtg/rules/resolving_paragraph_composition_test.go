package rules

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestCompiledResolvingThirdParagraphUsesItsOwnEvaluation(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Independent Resolving Check", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You lose 4 life.\nFateful hour — If you have 5 or less life, you lose 8 life instead.\nFormidable — If you have 5 or less life, draw a card.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].Life = 6
	addCardToLibrary(g, game.Player2, vanillaCreature("Drawn", 1, 1))
	NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player2},
		def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player2].Life != 2 || g.Players[game.Player2].Hand.Size() != 1 {
		t.Fatal("third paragraph aliased earlier false predicate instead of checking after base life loss")
	}
}

func TestCompiledResolvingParagraphConditionCapturedBeforeMutation(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Resolving Life Branch", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You lose 4 life.\nFateful hour — If you have 5 or less life, you lose 8 life instead.",
	})
	for _, initial := range []int{5, 6} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Players[game.Player2].Life = initial
		NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player2},
			def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		amount := 4
		if initial == 5 {
			amount = 8
		}
		if got := g.Players[game.Player2].Life; got != initial-amount {
			t.Fatalf("initial=%d got=%d want=%d; replacement predicate rechecked after base mutation",
				initial, got, initial-amount)
		}
	}
}

func TestCompiledResolvingParagraphKeepsIndependentTargetAndRider(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Resolving Counter Branch", Layout: "normal", TypeLine: "Instant",
		OracleText: "Tap target creature. Put a +1/+1 counter on target creature.\nMorbid — Put three +1/+1 counters on that creature instead if a creature died this turn.",
	})
	for _, died := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		first := addCombatPermanent(g, game.Player1, vanillaCreature("First", 2, 4))
		second := addCombatPermanent(g, game.Player3, vanillaCreature("Second", 2, 4))
		if died {
			departed := addCombatPermanent(g, game.Player4, vanillaCreature("Departed", 1, 1))
			if !sacrificePermanent(g, departed) {
				t.Fatal("ordinary death-history setup failed")
			}
		}

		obj := &game.StackObject{Controller: game.Player2,
			Targets: []game.Target{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)}}
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
			[game.NumPlayers]PlayerAgent{}, &TurnLog{})
		want := 1
		if died {
			want = 3
		}
		if !first.Tapped || first.Counters.Get(counter.PlusOnePlusOne) != 0 ||
			second.Tapped || second.Counters.Get(counter.PlusOnePlusOne) != want {
			t.Fatalf("independent target/rider or exclusive branch lost: first=%+v second=%+v", first, second)
		}
	}
}
