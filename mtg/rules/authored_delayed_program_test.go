package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

var delayedExileProgram = game.ProgramReference{ID: "fixture/delayed-exile", Version: "1"}
var delayedReturnProgram = game.ProgramReference{ID: "fixture/delayed-return", Version: "1"}

const delayedCaptureKey game.LinkedKey = "fixture/delayed-card"

func resolveDelayedExileFixture(ctx cardprogram.Context) error {
	exiled, err := ctx.Execute(game.MovePermanent{
		Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinked: delayedCaptureKey,
	})
	if err != nil || !exiled.Succeeded {
		return err
	}
	_, err = ctx.Execute(game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing:              game.DelayedAtBeginningOfNextEndStep,
		CapturedCard:        opt.Val(game.LinkedObjectReference(string(delayedCaptureKey))),
		CapturedObjectGroup: opt.Val(game.LinkedObjectReference(string(delayedCaptureKey))),
		Content:             game.Mode{Program: delayedReturnProgram}.Ability(),
	}})
	return err
}

func resolveDelayedReturnFixture(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.PutOnBattlefield{
		Source:            game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceCaptured}),
		LinkedReturnZones: []zone.Type{zone.Exile},
	})
	return err
}

func delayedProgramEntry() cardset.Entry {
	return cardset.Entry{
		OracleID: "fixture/delayed", Name: "Delayed linked return fixture",
		Verification: "mtg/rules/authored_delayed_program_test.go: delayed program fixture",
		New: func() *game.CardDef {
			return &game.CardDef{CardFace: game.CardFace{
				Name: "Delayed linked return fixture", Types: []types.Card{types.Artifact},
				ActivatedAbilities: []game.ActivatedAbility{{
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{{Allow: game.TargetAllowPermanent, MinTargets: 1, MaxTargets: 1}},
						Program: delayedExileProgram,
					}.Ability(),
				}},
			}}
		},
		Programs: []cardprogram.Program{
			{Reference: delayedExileProgram, Resolve: resolveDelayedExileFixture, Dependencies: []game.ProgramReference{delayedReturnProgram}},
			{Reference: delayedReturnProgram, Resolve: resolveDelayedReturnFixture},
		},
	}
}

func delayedFixtureGame(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	registry, err := cards.NewProgramRegistry([]cardset.Entry{delayedProgramEntry()})
	if err != nil {
		t.Fatal(err)
	}
	def, err := registry.LookupOracleID("fixture/delayed")
	if err != nil {
		t.Fatal(err)
	}
	g, _ := setupBestowMain(t)
	return g, NewEngineWithPrograms(nil, registry.Programs()), addCombatPermanent(g, game.Player1, def)
}

func activateDelayedFixture(t *testing.T, g *game.Game, e *Engine, source, target *game.Permanent) {
	t.Helper()
	if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)) {
		t.Fatal("registered delayed fixture did not activate")
	}
	e.resolveTopOfStack(g, &TurnLog{})
}

func TestAuthoredDelayedRegisteredCaptureCloneLiveAndSimulator(t *testing.T) {
	g, e, source := delayedFixtureGame(t)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	before := g.Clone()
	activateDelayedFixture(t, g, e, source, first)
	activateDelayedFixture(t, g, e, source, second)
	if len(before.DelayedTriggers) != 0 || len(before.LinkedObjects) != 0 || len(g.DelayedTriggers) != 2 {
		t.Fatal("pre-capture clone aliased scheduling/publication")
	}
	for i, target := range []*game.Permanent{first, second} {
		trigger := g.DelayedTriggers[i]
		card, _ := g.GetCardInstance(target.CardInstanceID)
		if trigger.CapturedCardID != card.ID || trigger.CapturedCardZoneVersion != card.ZoneVersion ||
			trigger.Ability.Content.Modes[0].Program != delayedReturnProgram {
			t.Fatal("scheduled state lost explicit captured incarnation or stable handler binding")
		}
	}
	movePermanentToZone(g, source, zone.Graveyard)
	clone := g.Clone()
	clone.DelayedTriggers[0].CapturedCardID = 0
	if len(clone.DelayedTriggers[0].CapturedObjectIDs) != 1 {
		t.Fatal("actual activation did not capture linked object history")
	}
	clone.DelayedTriggers[0].CapturedObjectIDs[0] = 0
	if g.DelayedTriggers[0].CapturedObjectIDs[0] != first.ObjectID {
		t.Fatal("Clone aliased mutable captured object history")
	}
	e.runEndingPhase(clone, [game.NumPlayers]PlayerAgent{})
	if permanentByCardID(clone, first.CardInstanceID) != nil || permanentByCardID(clone, second.CardInstanceID) == nil ||
		len(g.DelayedTriggers) != 2 || g.DelayedTriggers[0].CapturedCardID != first.CardInstanceID {
		t.Fatal("delayed clone shared captures, lost bindings or required a live source")
	}
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !e.putTriggeredAbilitiesOnStack(g) || len(g.DelayedTriggers) != 0 || g.Stack.Size() != 2 {
		t.Fatal("actual boundary failed to queue both delayed programs")
	}
	branch := e.Simulator().ResolvePriority(g, [game.NumPlayers]PlayerAgent{})
	if g.Stack.Size() != 2 || permanentByCardID(g, first.CardInstanceID) != nil {
		t.Fatal("pending stack programs mutated the Simulator input")
	}
	e.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, target := range []*game.Permanent{first, second} {
		for _, state := range []*game.Game{g, branch} {
			returned := permanentByCardID(state, target.CardInstanceID)
			if returned == nil || returned.ObjectID == target.ObjectID || returned.Controller != target.Owner {
				t.Fatal("live/Simulator delayed programs did not return exact captures under owner control")
			}
		}
	}
}

func TestAuthoredDelayedDoesNotReturnLaterExileIncarnation(t *testing.T) {
	g, e, source := delayedFixtureGame(t)
	captive := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	activateDelayedFixture(t, g, e, source, captive)
	moveCardBetweenZones(g, captive.Owner, captive.CardInstanceID, zone.Exile, zone.Hand)
	moveCardBetweenZones(g, captive.Owner, captive.CardInstanceID, zone.Hand, zone.Exile)
	e.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if permanentByCardID(g, captive.CardInstanceID) != nil || !g.Players[captive.Owner].Exile.Contains(captive.CardInstanceID) {
		t.Fatal("delayed captured CardID reclaimed a different zone incarnation")
	}
}

func TestAuthoredDelayedExecuteRejectsUninstalledNestedHandler(t *testing.T) {
	g, _, source := delayedFixtureGame(t)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Kind: game.StackActivatedAbility, Controller: game.Player1,
	}
	ctx := &programContext{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}, active: true}
	_, err := ctx.Execute(game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep, Content: game.Mode{Program: delayedReturnProgram}.Ability(),
	}})
	if err == nil || !strings.Contains(err.Error(), delayedReturnProgram.ID) || len(g.DelayedTriggers) != 0 {
		t.Fatal("Execute scheduled an uninstalled nested program")
	}
}

func TestAuthoredDelayedMissingCatalogRejectedBeforeSchedulingAndLoading(t *testing.T) {
	entry := delayedProgramEntry()
	entry.Programs = entry.Programs[:1]
	if _, err := cards.NewProgramRegistry([]cardset.Entry{entry}); err == nil || !strings.Contains(err.Error(), delayedReturnProgram.ID) {
		t.Fatal("registry silently admitted an omitted delayed handler")
	}
	for _, location := range []string{"scheduled", "stack"} {
		t.Run(location, func(t *testing.T) {
			g, e, source := delayedFixtureGame(t)
			captive := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			activateDelayedFixture(t, g, e, source, captive)
			if location == "stack" {
				emitBeginningOfStepEvent(g, game.StepEnd)
				e.putTriggeredAbilitiesOnStack(g)
			}
			parent := delayedProgramEntry().Programs[0]
			parent.Dependencies = nil
			incomplete, err := cardprogram.NewCatalog([]cardprogram.Program{parent})
			if err != nil {
				t.Fatal(err)
			}
			for _, run := range []func(){
				func() { NewEngineWithPrograms(nil, incomplete).validateLoadedPrograms(g) },
				func() {
					NewEngineWithPrograms(nil, incomplete).Simulator().ResolvePriority(g, [game.NumPlayers]PlayerAgent{})
				},
			} {
				func() {
					defer func() {
						failure, ok := recover().(UnsupportedError)
						if !ok || !strings.Contains(failure.Reason, delayedReturnProgram.ID) {
							t.Fatal("missing pending handler did not fail explicitly")
						}
					}()
					run()
					t.Fatal("missing pending handler silently continued")
				}()
			}
		})
	}
}
