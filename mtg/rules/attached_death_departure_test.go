package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestAttachedDeathDepartureFreezesOnlyOriginalCompletedMove(t *testing.T) {
	def := attachedDeathWatcher(t)
	for _, scenario := range []string{
		"attachment SBA", "simultaneous aura first", "simultaneous host first", "clone before queue",
		"ordinary later departure", "replacement exile", "no legal host",
		"move before queue", "move before schedule", "move before delayed resolution",
		"reenter before queue", "reenter before schedule", "reenter before delayed resolution",
	} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			aura := addCombatPermanent(g, game.Player3, def)
			aura.Controller = game.Player2
			host := addCombatPermanent(g, game.Player4, vanillaCreature("Host", 2, 3))
			host.Controller = game.Player1
			survivor := addCombatPermanent(g, game.Player4, vanillaCreature("Next Host", 2, 3))
			survivor.Controller = game.Player1
			if !attachPermanent(g, aura, host) {
				t.Fatal("initial attachment failed")
			}
			if scenario == "replacement exile" {
				resolveInstruction(engine, g, &game.StackObject{
					Controller: game.Player2, Targets: []game.Target{game.PermanentTarget(aura.ObjectID)},
				}, game.CreateReplacement{
					Object: game.TargetPermanentReference(0), Duration: game.DurationThisTurn,
					Replacement: &game.ReplacementEffect{
						MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Battlefield,
						MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
					},
				}, nil)
			}
			if scenario == "simultaneous aura first" || scenario == "simultaneous host first" {
				subjects := []*game.Permanent{aura, host}
				if scenario == "simultaneous host first" {
					subjects = []*game.Permanent{host, aura}
				}
				if !movePermanentsToZoneSimultaneously(g, subjects, zone.Graveyard) {
					t.Fatal("simultaneous departure failed")
				}
			} else {
				if _, destroyed := destroyPermanent(g, host.ObjectID); !destroyed {
					t.Fatal("actual host death failed")
				}
				if scenario == "ordinary later departure" && !movePermanentToZone(g, aura, zone.Graveyard) {
					t.Fatal("ordinary non-SBA departure failed")
				}
				engine.applyStateBasedActions(g)
			}
			var originalEvent game.Event
			for _, event := range g.Events {
				if event.Kind == game.EventPermanentDied && event.PermanentID == host.ObjectID {
					originalEvent = event
				}
			}
			if len(originalEvent.TriggeredAbilities) != 1 {
				t.Fatal("host event did not retain one original body")
			}
			version := g.CardInstances[aura.CardInstanceID].ZoneVersion
			if scenario == "clone before queue" {
				clone := g.Clone()
				if !engine.putTriggeredAbilitiesOnStack(clone) {
					t.Fatal("cloned pending certificate was lost")
				}
				if _, _, known := g.AttachedDeathDeparture(host.ObjectID, aura.ObjectID); !known {
					t.Fatal("clone queue consumed original game's private certificate")
				}
				g = clone
			}
			divert := func(reenter bool) {
				t.Helper()
				if reenter {
					cardTarget := currentCardTarget(t, g, aura.CardInstanceID)
					resolveInstruction(engine, g, &game.StackObject{
						Controller: game.Player3, Targets: []game.Target{cardTarget},
					}, game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget})}, nil)
					current := permanentForCard(g, aura.CardInstanceID)
					if current == nil || current.ObjectID == aura.ObjectID {
						t.Fatal("legal Aura reentry did not create a new incarnation")
					}
					if !movePermanentToZone(g, current, zone.Graveyard) {
						t.Fatal("new incarnation departure failed")
					}
				} else if !moveCardBetweenZones(g, game.Player3, aura.CardInstanceID, zone.Graveyard, zone.Exile) {
					t.Fatal("captured card diversion failed")
				}
			}
			if scenario == "move before queue" || scenario == "reenter before queue" {
				divert(scenario == "reenter before queue")
			}
			if scenario != "clone before queue" && !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("valid captured body disappeared")
			}
			queued, _ := g.Stack.Peek()
			eligible := scenario != "ordinary later departure" && scenario != "replacement exile"
			if queued.SourceID != aura.ObjectID || queued.SourceCardID != aura.CardInstanceID ||
				queued.Controller != game.Player2 || !reflect.DeepEqual(queued.TriggerEvent, originalEvent) ||
				eligible && (queued.SourceZone != zone.Graveyard || queued.SourceZoneVersion != version) ||
				!eligible && (queued.SourceZone != zone.None || queued.SourceZoneVersion != 0) {
				t.Fatal("certificate binding changed event/body/controller or accepted an uncertified departure")
			}
			if _, _, known := g.AttachedDeathDeparture(host.ObjectID, aura.ObjectID); known {
				t.Fatal("queue retained an unconsumed causal certificate")
			}
			if scenario == "move before schedule" || scenario == "reenter before schedule" {
				divert(scenario == "reenter before schedule")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if len(g.DelayedTriggers) != 1 {
				t.Fatal("captured body did not schedule its delayed instruction")
			}
			delayed := g.DelayedTriggers[0]
			captureExpected := eligible && scenario != "move before queue" && scenario != "reenter before queue" &&
				scenario != "move before schedule" && scenario != "reenter before schedule"
			if captureExpected && (delayed.CapturedCardID != aura.CardInstanceID || delayed.CapturedCardZoneVersion != version) ||
				!captureExpected && (delayed.CapturedCardID != 0 || delayed.CapturedCardZoneVersion != 0) {
				t.Fatal("strict delayed capture accepted a newer or uncertified source incarnation")
			}
			if scenario == "move before delayed resolution" || scenario == "reenter before delayed resolution" {
				divert(scenario == "reenter before delayed resolution")
			}
			if scenario == "no legal host" && !movePermanentToZone(g, survivor, zone.Exile) {
				t.Fatal("last legal attachment subject removal failed")
			}
			g.Turn.Step = game.StepEnd
			emitBeginningOfStepEvent(g, game.StepEnd)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("printed delayed step did not queue")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			returned := permanentForCard(g, aura.CardInstanceID)
			wantReturn := scenario == "attachment SBA" || scenario == "simultaneous aura first" ||
				scenario == "simultaneous host first" || scenario == "clone before queue"
			if wantReturn && (returned == nil || returned.ObjectID == aura.ObjectID ||
				returned.Owner != game.Player3 || returned.Controller != game.Player3 ||
				!returned.AttachedTo.Exists || returned.AttachedTo.Val != survivor.ObjectID) ||
				!wantReturn && returned != nil {
				t.Fatal("delayed return lost default owner/legal attachment or accepted wrong incarnation")
			}
			if scenario == "no legal host" && !g.Players[game.Player3].Graveyard.Contains(aura.CardInstanceID) {
				t.Fatal("Aura without a legal attachment left its graveyard")
			}
		})
	}
}
