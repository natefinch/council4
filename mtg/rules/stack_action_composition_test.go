package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func compiledStackActionCard(t *testing.T, text string) *game.CardDef {
	t.Helper()
	return legacyCard(t, carddata.ScryfallCard{
		Name: "Stack Action Probe", Layout: "normal", TypeLine: "Instant", OracleText: text,
	})
}

func TestCompiledStackActionsPreserveSubjectAndNewControllerLegality(t *testing.T) {
	for _, prefix := range []string{
		"", "Scry 1. ", "Target player gains 1 life. ",
		"Exile target card from your graveyard. ",
		"Change the target of target spell with a single target. ",
	} {
		for _, accept := range []bool{false, true} {
			t.Run(prefix+map[bool]string{false: "decline", true: "accept"}[accept], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				old := addCreaturePermanent(g, game.Player2)
				next := addCreaturePermanent(g, game.Player1)
				neighborTarget := addCreaturePermanent(g, game.Player3)
				def := compiledStackActionCard(t, "Target creature you control gets +1/+1 until end of turn.")
				victim := addVictimSpellOnStack(g, game.Player4, game.Player2, def)
				victim.Targets = []game.Target{game.PermanentTarget(old.ObjectID)}
				victim.TargetCounts = []int{1}
				neighbor := addVictimSpellOnStack(g, game.Player3, game.Player3, def)
				neighbor.Targets = []game.Target{game.PermanentTarget(neighborTarget.ObjectID)}
				neighbor.TargetCounts = []int{1}
				originalID, originalSource := victim.ID, victim.SourceID
				targets := []game.Target{game.StackObjectTarget(victim.ID)}
				choices := [][]int{}
				switch prefix {
				case "", "Scry 1. ":
				case "Target player gains 1 life. ":
					targets = append([]game.Target{game.PlayerTarget(game.Player2)}, targets...)
				case "Exile target card from your graveyard. ":
					card := addCardToGraveyard(g, game.Player1, simpleSorceryDef("Earlier Card"))
					targets = append([]game.Target{currentCardTarget(t, g, card)}, targets...)
				case "Change the target of target spell with a single target. ":
					// The neighbor's controller has no other creature, so its
					// target can't change and no change prompt is offered (CR 115.7a).
					targets = append([]game.Target{game.StackObjectTarget(neighbor.ID)}, targets...)
				default:
					t.Fatalf("unexpected stack-action prefix %q", prefix)
				}
				optionalIndex := len(choices)
				decision := 0
				if accept {
					decision = 1
				}
				choices = append(choices, []int{decision})
				if accept {
					choices = append(choices, []int{0})
				}
				agent := &choiceOnlyAgent{choices: choices}
				action := addVictimSpellOnStack(g, game.Player3, game.Player1,
					compiledStackActionCard(t, prefix+"Gain control of target spell. You gain 2 life. You may choose new targets for it."))
				action.Targets = targets
				action.TargetCounts = make([]int, len(targets))
				for i := range action.TargetCounts {
					action.TargetCounts[i] = 1
				}
				log := &TurnLog{}
				NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, log)
				if victim.Controller != game.Player1 || victim.ID != originalID ||
					victim.SourceID != originalSource || g.CardInstances[victim.SourceID].Owner != game.Player4 {
					t.Fatal("stack mutation changed identity/ownership or used the resolving card's physical owner")
				}
				want := old.ObjectID
				wantCounts := []int{1}
				if accept {
					want = next.ObjectID
					request := log.Choices[optionalIndex+1].Request
					if request.Player != game.Player1 || len(request.Options) != 1 ||
						!slices.Equal(request.Options[0].Targets, []game.Target{game.PermanentTarget(next.ObjectID)}) {
						t.Fatalf("retarget legality did not use new controller: %+v", request)
					}
				}
				if victim.Targets[0].PermanentID != want || !slices.Equal(victim.TargetCounts, wantCounts) ||
					neighbor.Controller != game.Player3 || neighbor.Targets[0].PermanentID != neighborTarget.ObjectID {
					t.Fatalf("target=%v want=%v counts=%v neighbor=%v/%v want=%v",
						victim.Targets, want, victim.TargetCounts, neighbor.Controller, neighbor.Targets, neighborTarget.ObjectID)
				}
				if g.Players[game.Player1].Life != 42 || g.Players[game.Player3].Life != 40 ||
					g.Players[game.Player4].Life != 40 {
					t.Fatal("independent YOU rider lost its source controller")
				}
				if prefix == "Target player gains 1 life. " && g.Players[game.Player2].Life != 41 {
					t.Fatal("earlier player target was lost or used as the stack subject")
				}
				if prefix == "Exile target card from your graveyard. " && g.Players[game.Player1].Exile.Size() != 1 {
					t.Fatal("earlier card target was lost")
				}
			})
		}
	}
}

func TestCompiledStackActionsGoneTargetDoesNotRebindSameCard(t *testing.T) {
	for _, prefix := range []string{"", "Target player gains 1 life. "} {
		for _, replace := range []bool{false, true} {
			t.Run(prefix+map[bool]string{false: "gone", true: "same source new occurrence"}[replace], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				victim := addVictimSpellOnStack(g, game.Player4, game.Player2, simpleSorceryDef("Victim"))
				neighbor := addVictimSpellOnStack(g, game.Player3, game.Player3, simpleSorceryDef("Neighbor"))
				action := addVictimSpellOnStack(g, game.Player3, game.Player1,
					compiledStackActionCard(t, prefix+"Gain control of target spell. You gain 2 life. You may choose new targets for it."))
				action.Targets = []game.Target{game.StackObjectTarget(victim.ID)}
				action.TargetCounts = []int{1}
				if prefix != "" {
					action.Targets = append([]game.Target{game.PlayerTarget(game.Player2)}, action.Targets...)
					action.TargetCounts = []int{1, 1}
				}
				if _, ok := g.Stack.RemoveByID(victim.ID); !ok {
					t.Fatal("cannot remove victim")
				}
				if replace {
					replacement := *victim
					replacement.ID = g.IDGen.Next()
					g.CardInstances[victim.SourceID].ZoneVersion++
					// Keep the resolving action on top; the same CardID is
					// not the removed stack occurrence.
					g.Stack.Pop()
					g.Stack.Push(&replacement)
					g.Stack.Push(action)
					neighbor = &replacement
				}
				NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
				if victim.Controller != game.Player2 || neighbor.Controller == game.Player1 {
					t.Fatal("gone target rebound by card identity or neighboring stack position")
				}
				wantLife := 40
				if prefix != "" {
					wantLife = 42
				}
				if g.Players[game.Player1].Life != wantLife {
					t.Fatal("all-target fizzle or independently legal player target was not preserved")
				}
			})
		}
	}
}

func TestCompiledStackActionsNoLegalRetargetStillChangesControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	old := addCreaturePermanent(g, game.Player2)
	victim := addVictimSpellOnStack(g, game.Player4, game.Player2,
		compiledStackActionCard(t, "Target creature you control gets +1/+1 until end of turn."))
	victim.Targets = []game.Target{game.PermanentTarget(old.ObjectID)}
	victim.TargetCounts = []int{1}
	action := addVictimSpellOnStack(g, game.Player3, game.Player1,
		compiledStackActionCard(t, "Gain control of target spell. You gain 2 life. You may choose new targets for it."))
	action.Targets = []game.Target{game.StackObjectTarget(victim.ID)}
	action.TargetCounts = []int{1}
	NewEngine(nil).resolveTopOfStackWithChoices(g,
		[game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}, &TurnLog{})
	if victim.Controller != game.Player1 || victim.Targets[0].PermanentID != old.ObjectID || g.Players[game.Player1].Life != 42 {
		t.Fatal("unavailable new targets undid control or the independent rider")
	}
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if len(g.ContinuousEffects) != 0 || !g.Players[game.Player4].Graveyard.Contains(victim.SourceID) ||
		g.Players[game.Player1].Graveyard.Contains(victim.SourceID) {
		t.Fatal("now-illegal old target resolved or stolen spell lost its physical owner's destination")
	}
}

func TestCompiledStackActionsRejectIllegalSpellKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	victim := addVictimSpellOnStack(g, game.Player4, game.Player2,
		&game.CardDef{CardFace: game.CardFace{Name: "Creature", Types: []types.Card{types.Creature}}})
	action := addVictimSpellOnStack(g, game.Player3, game.Player1,
		compiledStackActionCard(t, "Gain control of target noncreature spell. You gain 2 life. You may choose new targets for it."))
	action.Targets = []game.Target{game.StackObjectTarget(victim.ID)}
	action.TargetCounts = []int{1}
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if victim.Controller != game.Player2 || g.Players[game.Player1].Life != 40 {
		t.Fatal("illegal creature-spell target bypassed the original filter/fizzle contract")
	}
	if !g.Players[game.Player3].Graveyard.Contains(action.SourceID) || g.CardInstances[victim.SourceID].Owner != game.Player4 {
		t.Fatal("fizzle mutated either source owner's destination")
	}
}

func TestCompiledStackActionCopyRetainsProtocol(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addVictimSpellOnStack(g, game.Player4, game.Player2, simpleSorceryDef("First"))
	second := addVictimSpellOnStack(g, game.Player3, game.Player2, simpleSorceryDef("Second"))
	action := addVictimSpellOnStack(g, game.Player4, game.Player1,
		compiledStackActionCard(t, "Gain control of target spell. You gain 2 life. You may choose new targets for it."))
	action.Targets = []game.Target{game.StackObjectTarget(first.ID)}
	action.TargetCounts = []int{1}
	copiedAction := game.NewStackObjectCopy(action, g.IDGen.Next())
	copiedAction.Controller = game.Player3
	g.Stack.Push(copiedAction)
	// The existing copy protocol chooses the copy's targets independently.
	copiedAction.Targets = []game.Target{game.StackObjectTarget(second.ID)}
	NewEngine(nil).resolveTopOfStackWithChoices(g,
		[game.NumPlayers]PlayerAgent{game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}}}, &TurnLog{})
	if second.Controller != game.Player3 || first.Controller != game.Player2 ||
		g.Players[game.Player3].Life != 42 || g.Players[game.Player1].Life != 40 {
		t.Fatal("copied body used original controller or original target")
	}
	NewEngine(nil).resolveTopOfStackWithChoices(g,
		[game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}, &TurnLog{})
	if first.Controller != game.Player1 || second.Controller != game.Player3 ||
		g.Players[game.Player1].Life != 42 || g.CardInstances[action.SourceID].Owner != game.Player4 ||
		!g.Players[game.Player4].Graveyard.Contains(action.SourceID) {
		t.Fatal("copy leaked target/controller/owner state into the original")
	}
}

func TestCompiledStackActionsPreserveChosenModesAndX(t *testing.T) {
	for _, accept := range []bool{false, true} {
		t.Run(map[bool]string{false: "decline", true: "accept"}[accept], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			old := addCombatPermanent(g, game.Player2, vanillaCreature("Old Target", 1, 1))
			next := addCombatPermanent(g, game.Player1, vanillaCreature("New Target", 1, 1))
			victimDef := compiledStackActionCard(t, "Choose one \u2014\n\u2022 Draw a card.\n\u2022 Target creature you control gets +2/+2 until end of turn.")
			victim := addVictimSpellOnStack(g, game.Player4, game.Player2, victimDef)
			victim.ChosenModes = []int{1}
			victim.XValue = 7
			victim.Targets = []game.Target{game.PermanentTarget(old.ObjectID)}
			victim.TargetCounts = []int{1}
			action := addVictimSpellOnStack(g, game.Player3, game.Player1,
				compiledStackActionCard(t, "Gain control of target spell. You gain 2 life. You may choose new targets for it."))
			action.Targets = []game.Target{game.StackObjectTarget(victim.ID)}
			action.TargetCounts = []int{1}
			choices := [][]int{{0}}
			if accept {
				choices = [][]int{{1}, {0}}
			}
			engine := NewEngine(nil)
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: choices},
			}, &TurnLog{})
			if !slices.Equal(victim.ChosenModes, []int{1}) || victim.XValue != 7 {
				t.Fatal("control/retarget changed the announced mode or X")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			want := 1
			if accept {
				want = 3
			}
			if effectivePower(g, next) != want || effectivePower(g, old) != 1 {
				t.Fatal("stolen modal spell resolved a different mode or an illegal retained target")
			}
		})
	}
}
