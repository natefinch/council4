package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledActivationExactXRejectsWrongCountBeforeCosts(t *testing.T) {
	for _, tc := range []struct {
		name        string
		typeLine    string
		body        string
		creatures   bool
		destination zone.Type
	}{
		{"tap lands", "Artifact", "{X}, {T}: Tap X target lands.", false, zone.None},
		{"double X", "Artifact", "{X}{X}{1}, {T}: Tap X target lands.", false, zone.None},
		{"sacrifice destroy", "Creature", "{X}{X}{R}, {T}, Sacrifice this creature: Destroy X target lands.", false, zone.Graveyard},
		{"discard return", "Creature", "{X}{U}, {T}, Discard two cards: Return X target creatures to their owners' hands.", true, zone.Hand},
	} {
		t.Run(tc.name, func(t *testing.T) {
			card := &cardgen.ScryfallCard{Name: "Count Activation", Layout: "normal", TypeLine: tc.typeLine, OracleText: tc.body}
			if tc.typeLine == "Creature" {
				card.Power, card.Toughness = new("2"), new("2")
			}
			defs, diagnostics, err := cardgen.CompileCardDefs(card)
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile: defs=%d diagnostics=%+v err=%v", len(defs), diagnostics, err)
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
			source := addCombatPermanent(g, game.Player1, defs[0])
			source.SummoningSick = false
			chosen := addLandPermanent(g, game.Player2, "Chosen Land")
			if tc.creatures {
				chosen = addCreaturePermanent(g, game.Player2)
			}
			second := addLandPermanent(g, game.Player2, "Second Land")
			if tc.creatures {
				second = addCreaturePermanent(g, game.Player2)
			}
			for _, color := range []mana.Color{mana.C, mana.U, mana.R} {
				g.Players[game.Player1].ManaPool.Add(color, 8)
			}
			for range 3 {
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cost Card"}})
			}
			engine := NewEngine(nil)
			targets := []game.Target{game.PermanentTarget(chosen.ObjectID)}
			before := g.Clone()
			wrong := action.ActivateAbility(source.ObjectID, 0, targets, 0)
			if engine.applyAction(g, game.Player1, wrong) {
				t.Fatal("activation accepted one target with X=0")
			}
			if source.Tapped || g.Stack.Size() != 0 ||
				!reflect.DeepEqual(before.Players[game.Player1].ManaPool, g.Players[game.Player1].ManaPool) ||
				!reflect.DeepEqual(before.Players[game.Player1].Hand.All(), g.Players[game.Player1].Hand.All()) {
				t.Fatal("invalid target count paid costs or announced an ability")
			}
			if _, exists := permanentByObjectID(g, source.ObjectID); !exists {
				t.Fatal("invalid target count sacrificed its source")
			}
			for _, x := range []int{0, 1, 2, 3} {
				act := action.ActivateAbility(source.ObjectID, 0, targets, x)
				if got := containsAction(engine.legalActions(g, game.Player1), act); got != (x == 1) {
					t.Fatalf("legal one-target activation with X=%d: %v", x, got)
				}
				for _, targetCount := range []int{0, 2} {
					clone := g.Clone()
					var group []game.Target
					if targetCount == 2 {
						group = []game.Target{game.PermanentTarget(chosen.ObjectID), game.PermanentTarget(second.ObjectID)}
					}
					if !engine.applyAction(clone, game.Player1, action.ActivateAbility(source.ObjectID, 0, group, targetCount)) {
						t.Fatalf("exact X=%d activation rejected", targetCount)
					}
					obj, ok := clone.Stack.Peek()
					if !ok || obj.XValue != targetCount || len(obj.TargetCounts) != 1 || obj.TargetCounts[0] != targetCount {
						t.Fatalf("X/count %d not retained in announcement: %+v", targetCount, obj)
					}
				}
			}
			correct := action.ActivateAbility(source.ObjectID, 0, targets, 1)
			if !engine.applyAction(g, game.Player1, correct) {
				t.Fatal("exact one-target X=1 activation rejected")
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.XValue != 1 || len(obj.TargetCounts) != 1 || obj.TargetCounts[0] != 1 {
				t.Fatalf("announcement lost chosen X/count: %+v", obj)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if tc.destination == zone.None {
				if !chosen.Tapped {
					t.Fatal("selected land was not tapped")
				}
			} else {
				if _, exists := permanentByObjectID(g, chosen.ObjectID); exists {
					t.Fatal("chosen target remained on the battlefield")
				}
				dest := g.Players[game.Player2].Graveyard
				if tc.destination == zone.Hand {
					dest = g.Players[game.Player2].Hand
				}
				if !dest.Contains(chosen.CardInstanceID) {
					t.Fatal("selected target did not reach its owner's destination")
				}
			}
		})
	}
}
