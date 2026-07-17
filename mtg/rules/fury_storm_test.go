package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func furyStormTestStack(
	g *game.Game,
	triggerController game.PlayerID,
) (spell *game.StackObject, victimA, victimB *game.Permanent) {
	victimA = addCreaturePermanent(g, game.Player3)
	victimB = addCreaturePermanent(g, game.Player3)
	spellID := g.IDGen.Next()
	g.CardInstances[spellID] = &game.CardInstance{
		ID: spellID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Fury Storm",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
			}.Ability()),
		}},
		Owner: game.Player1,
	}
	spell = &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     spellID,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PermanentTarget(victimA.ObjectID)},
		TargetCounts: []int{1},
	}
	g.Stack.Push(spell)
	trigger := game.TriggeredAbility{Content: game.Mode{Sequence: []game.Instruction{{
		Primitive: game.CopyStackObject{
			Object: game.EventStackObjectReference(),
			DynamicCount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountCommanderCastCount,
				Multiplier: 1,
			}),
			MayChooseNewTargets: true,
		},
	}}}.Ability()}
	g.Stack.Push(&game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		Controller:      triggerController,
		InlineTrigger:   &trigger,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:          game.EventSpellCast,
			Controller:    game.Player1,
			StackObjectID: spell.ID,
		},
	})
	return spell, victimA, victimB
}

func TestFuryStormDynamicBatchUsesCurrentTriggerController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell, victimA, victimB := furyStormTestStack(g, game.Player2)
	g.Players[game.Player1].CommanderCastCount = 1
	g.Players[game.Player2].CommanderCastCount = 2
	g.Events = append(g.Events, game.Event{Kind: game.EventSpellCast, StackObjectID: spell.ID})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}, {1}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Stack.Size(); got != 3 {
		t.Fatalf("stack size = %d, want original plus two copies", got)
	}
	copies := 0
	for _, obj := range g.Stack.Objects() {
		if !obj.Copy {
			continue
		}
		copies++
		if obj.Controller != game.Player2 {
			t.Fatalf("copy controller = %v, want changed trigger controller Player2", obj.Controller)
		}
		if len(obj.Targets) != 1 || obj.Targets[0].PermanentID != victimB.ObjectID {
			t.Fatalf("copy targets = %+v, want victim B %v", obj.Targets, victimB.ObjectID)
		}
	}
	if copies != 2 {
		t.Fatalf("copies = %d, want 2", copies)
	}
	if spell.Targets[0].PermanentID != victimA.ObjectID {
		t.Fatal("retargeting copies changed the original spell")
	}
	if got := countEventsOfKind(g.Events, game.EventSpellCast); got != 1 {
		t.Fatalf("spell-cast events = %d, want copies not to be cast", got)
	}
	if got := countEventsOfKind(g.Events, game.EventSpellCopied); got != 2 {
		t.Fatalf("spell-copied events = %d, want 2", got)
	}
}

func TestFuryStormCopyMayKeepIllegalOriginalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	_, victimA, _ := furyStormTestStack(g, game.Player1)
	g.Players[game.Player1].CommanderCastCount = 1
	for i, permanent := range g.Battlefield {
		if permanent.ObjectID == victimA.ObjectID {
			g.Battlefield = append(g.Battlefield[:i], g.Battlefield[i+1:]...)
			break
		}
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	copiedSpell, ok := g.Stack.Peek()
	if !ok || !copiedSpell.Copy {
		t.Fatalf("copy not created: %+v", copiedSpell)
	}
	if len(copiedSpell.Targets) != 1 || copiedSpell.Targets[0].PermanentID != victimA.ObjectID {
		t.Fatalf("declined copy targets = %+v, want original illegal target %v", copiedSpell.Targets, victimA.ObjectID)
	}
}

func TestFuryStormDynamicBatchReplacementAppliesOnce(t *testing.T) {
	for _, test := range []struct {
		name  string
		count int
		want  int
	}{
		{name: "zero", count: 0, want: 0},
		{name: "two", count: 2, want: 5},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			furyStormTestStack(g, game.Player1)
			g.Players[game.Player1].CommanderCastCount = test.count
			for i, addend := range []int{1, 2} {
				source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Copy Replacement " + string(rune('A'+i)),
					ReplacementAbilities: []game.ReplacementAbility{
						game.AdditionalSpellCopyReplacement("additional copies", addend, false),
					},
				}})
				registerPermanentReplacementEffects(g, source)
			}

			engine.resolveTopOfStack(g, &TurnLog{})

			copies := 0
			for _, obj := range g.Stack.Objects() {
				if obj.Copy {
					copies++
				}
			}
			if copies != test.want {
				t.Fatalf("copies = %d, want %d", copies, test.want)
			}
		})
	}
}

func TestFuryStormDynamicBatchSurvivesTriggerCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	furyStormTestStack(g, game.Player1)
	g.Players[game.Player1].CommanderCastCount = 1
	g.Players[game.Player2].CommanderCastCount = 3
	trigger, _ := g.Stack.Peek()
	addEffectSpellToStack(g, game.Player2,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(trigger.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})

	copies := 0
	for _, obj := range g.Stack.Objects() {
		if obj.Copy && obj.Kind == game.StackSpell {
			copies++
			if obj.Controller != game.Player2 {
				t.Fatalf("copied trigger made a spell copy controlled by %v, want Player2", obj.Controller)
			}
		}
	}
	if copies != 3 {
		t.Fatalf("spell copies = %d, want copied trigger controller's count 3", copies)
	}
}

func countEventsOfKind(events []game.Event, kind game.EventKind) int {
	count := 0
	for _, event := range events {
		if event.Kind == kind {
			count++
		}
	}
	return count
}
