package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledNonBattlefieldActivationCountsBeforeSourceCost(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		zone zone.Type
	}{
		{"hand", "Channel — {X}, Discard this card: Tap X target lands.", zone.Hand},
		{"graveyard", "{X}, Exile this card from your graveyard: Tap X target lands.", zone.Graveyard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defs := legacyCardDefs(t, carddata.ScryfallCard{
				Name: "Count Source", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("2"),
				OracleText: tc.body,
			})

			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
			source := addCardToHand(g, game.Player1, defs[0])
			if tc.zone == zone.Graveyard {
				g.Players[game.Player1].Hand.Remove(source)
				g.Players[game.Player1].Graveyard.Add(source)
			}
			land := addLandPermanent(g, game.Player2, "Chosen Land")
			targets := []game.Target{game.PermanentTarget(land.ObjectID)}
			g.Players[game.Player1].ManaPool.Add(mana.C, 4)
			engine := NewEngine(nil)
			before := g.Clone()
			if engine.applyAction(g, game.Player1, action.ActivateAbility(source, 0, targets, 0)) {
				t.Fatal("wrong X count accepted")
			}
			if g.Stack.Size() != 0 || !reflect.DeepEqual(before.Players[game.Player1], g.Players[game.Player1]) {
				t.Fatal("invalid non-battlefield activation moved its source or paid mana")
			}
			if !engine.applyAction(g, game.Player1, action.ActivateAbility(source, 0, targets, 1)) {
				t.Fatal("exact count non-battlefield activation rejected")
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.XValue != 1 || len(obj.TargetCounts) != 1 || obj.TargetCounts[0] != 1 {
				t.Fatalf("chosen count/X not captured: %+v", obj)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !land.Tapped {
				t.Fatal("non-battlefield activation failed to resolve")
			}
		})
	}
}

func TestCompiledModalActivationCountUsesSelectedModes(t *testing.T) {
	defs := legacyCardDefs(t, carddata.ScryfallCard{
		Name: "Count Modes", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{X}: Choose one —\n• Tap X target lands.\n• Draw a card.",
	})

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	source := addCombatPermanent(g, game.Player1, defs[0])
	land := addLandPermanent(g, game.Player2, "Chosen Land")
	targets := []game.Target{game.PermanentTarget(land.ObjectID)}
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	engine := NewEngine(nil)
	for _, x := range []int{0, 1, 2} {
		act := action.ActivateAbilityWithModes(source.ObjectID, 0, targets, x, []int{0})
		if got := containsAction(engine.legalActions(g, game.Player1), act); got != (x == 1) {
			t.Fatalf("targeted mode X=%d legal=%v", x, got)
		}
	}
	if !containsAction(engine.legalActions(g, game.Player1),
		action.ActivateAbilityWithModes(source.ObjectID, 0, nil, 2, []int{1})) {
		t.Fatal("unselected X-target mode constrained the untargeted selected mode")
	}
	before := g.Clone()
	wrongCounts := action.ActivateAbilityWithModesAndTargetCounts(source.ObjectID, 0, targets, []int{2}, 1, []int{0})
	if engine.applyAction(g, game.Player1, wrongCounts) ||
		!reflect.DeepEqual(before.Players[game.Player1].ManaPool, g.Players[game.Player1].ManaPool) {
		t.Fatal("invalid recorded mode counts were accepted or paid costs")
	}
	correct := action.ActivateAbilityWithModesAndTargetCounts(source.ObjectID, 0, targets, []int{1}, 1, []int{0})
	if !engine.applyAction(g, game.Player1, correct) {
		t.Fatal("correct recorded mode count rejected")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !land.Tapped {
		t.Fatal("selected mode did not affect its chosen target")
	}
}
