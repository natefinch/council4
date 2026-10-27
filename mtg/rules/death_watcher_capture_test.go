package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func attachedDeathWatcher(t *testing.T) *game.CardDef {
	t.Helper()
	return legacyCard(t, carddata.ScryfallCard{
		Name: "Attached Death Watcher", Layout: "normal", TypeLine: "Enchantment — Aura",
		OracleText: "Enchant creature\nWhen enchanted creature dies, return this card to the battlefield at the beginning of the next end step.",
	})
}

func TestAttachedDeathWatcherCapturesBeforeAuraDeparture(t *testing.T) {
	for _, order := range []string{"later attachment SBA", "simultaneous aura first", "simultaneous host first"} {
		t.Run(order, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := attachedDeathWatcher(t)
			aura := addCombatPermanent(g, game.Player3, def)
			aura.Controller = game.Player2
			host := addCombatPermanent(g, game.Player4, vanillaCreature("Original Host", 2, 3))
			host.Controller = game.Player1
			foreignHost := addCombatPermanent(g, game.Player1, vanillaCreature("Other Host", 2, 3))
			foreignAura := addCombatPermanent(g, game.Player1, def)
			foreignAura.Controller = game.Player4
			if !attachPermanent(g, aura, host) || !attachPermanent(g, foreignAura, foreignHost) {
				t.Fatal("legal independent Aura attachment failed")
			}
			if order == "later attachment SBA" {
				if _, moved := destroyPermanent(g, host.ObjectID); !moved {
					t.Fatal("natural host death failed")
				}
				engine.applyStateBasedActions(g)
			} else {
				subjects := []*game.Permanent{aura, host}
				if order == "simultaneous host first" {
					subjects = []*game.Permanent{host, aura}
				}
				if !movePermanentsToZoneSimultaneously(g, subjects, zone.Graveyard) {
					t.Fatal("natural simultaneous completed moves failed")
				}
			}
			if !g.Players[game.Player3].Graveyard.Contains(aura.CardInstanceID) ||
				permanentForCard(g, foreignAura.CardInstanceID) == nil {
				t.Fatal("original Aura did not depart or foreign attachment was consumed")
			}
			var deaths []game.Event
			for _, event := range g.Events {
				if event.Kind == game.EventPermanentDied && event.PermanentID == host.ObjectID {
					deaths = append(deaths, event)
				}
			}
			if len(deaths) != 1 || deaths[0].CardID != host.CardInstanceID ||
				deaths[0].CardZoneVersion == 0 ||
				deaths[0].CardZoneVersion != g.CardInstances[host.CardInstanceID].ZoneVersion ||
				!deaths[0].TriggeredAbilitiesCaptured || len(deaths[0].TriggeredAbilities) != 1 {
				t.Fatal("completed host death did not capture exactly one eligible original watcher")
			}
			snapshot := deaths[0].TriggeredAbilities[0]
			if snapshot.SourceID != aura.ObjectID || snapshot.SourceCardID != aura.CardInstanceID ||
				snapshot.Controller != game.Player2 || !reflect.DeepEqual(snapshot.Ability, &def.TriggeredAbilities[0]) {
				t.Fatal("captured host relation/body/source/ability controller rebound to a foreign Aura or owner")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("Aura departure dropped or duplicated the captured ordinary watcher")
			}
			queued, _ := g.Stack.Peek()
			if queued.SourceID != aura.ObjectID || queued.SourceCardID != aura.CardInstanceID ||
				queued.Controller != game.Player2 || queued.TriggerEvent.CardID != host.CardInstanceID ||
				!reflect.DeepEqual(queued.InlineTrigger, &def.TriggeredAbilities[0]) {
				t.Fatal("queue refetched the source/body or replaced the original host event")
			}
		})
	}
}

func TestAttachedDeathWatcherRefusesOtherHostAndNonDeath(t *testing.T) {
	for _, outcome := range []string{"other host", "host exile", "prevented destruction", "replaced death"} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			aura := addCombatPermanent(g, game.Player3, attachedDeathWatcher(t))
			aura.Controller = game.Player2
			hostDef := vanillaCreature("Original Host", 2, 3)
			if outcome == "prevented destruction" {
				hostDef.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
			}
			host := addCombatPermanent(g, game.Player4, hostDef)
			other := addCombatPermanent(g, game.Player1, vanillaCreature("Other Host", 2, 3))
			if !attachPermanent(g, aura, host) {
				t.Fatal("legal Aura attachment failed")
			}
			if outcome == "replaced death" {
				resolveInstruction(engine, g, &game.StackObject{
					Controller: game.Player2,
					Targets:    []game.Target{game.PermanentTarget(host.ObjectID)},
				}, game.CreateReplacement{
					Object: game.TargetPermanentReference(0), Duration: game.DurationThisTurn,
					Replacement: &game.ReplacementEffect{
						MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Battlefield,
						MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
					},
				}, nil)
			}
			if outcome == "other host" {
				destroyPermanent(g, other.ObjectID)
			} else if outcome == "host exile" {
				if !movePermanentToZone(g, host, zone.Exile) {
					t.Fatal("actual host diversion failed")
				}
			} else if outcome == "replaced death" {
				if _, moved := destroyPermanent(g, host.ObjectID); !moved ||
					!g.Players[host.Owner].Exile.Contains(host.CardInstanceID) ||
					g.Players[host.Owner].Graveyard.Contains(host.CardInstanceID) {
					t.Fatal("actual completed host departure did not use the replacement destination")
				}
			} else if _, destroyed := destroyPermanent(g, host.ObjectID); destroyed {
				t.Fatal("indestructible host was destroyed")
			}
			engine.applyStateBasedActions(g)
			if engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 0 {
				t.Fatal("foreign subject/non-death/prevented move invented an attached death trigger")
			}
		})
	}
}

func TestAttachedDeathWatcherCopiedBodyKeepsPhysicalSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := attachedDeathWatcher(t)
	template := addCombatPermanent(g, game.Player1, def)
	template.Controller = game.Player4
	otherHost := addCombatPermanent(g, game.Player1, vanillaCreature("Other Host", 2, 3))
	if !attachPermanent(g, template, otherHost) {
		t.Fatal("template Aura attachment failed")
	}
	sourceDef := legacyCard(t, carddata.ScryfallCard{
		Name: "Physical Copy Source", Layout: "normal", TypeLine: "Enchantment — Aura",
		OracleText: "Enchant creature",
	})
	source := addCombatPermanent(g, game.Player3, sourceDef)
	source.Controller = game.Player2
	resolveInstruction(engine, g, &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player2,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{game.PermanentTarget(template.ObjectID)},
	}, game.BecomeCopy{Object: game.TargetPermanentReference(0)}, nil)
	host := addCombatPermanent(g, game.Player4, vanillaCreature("Original Host", 2, 3))
	host.Controller = game.Player1
	if !attachPermanent(g, source, host) {
		t.Fatal("actual copied Aura did not have the legal attached-subject contract")
	}
	if _, destroyed := destroyPermanent(g, host.ObjectID); !destroyed {
		t.Fatal("natural host death failed")
	}
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player3].Graveyard.Contains(source.CardInstanceID) ||
		permanentForCard(g, template.CardInstanceID) != template {
		t.Fatal("copied watcher did not depart independently of its blueprint")
	}
	var captures []game.EventTriggeredAbility
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentDied && event.PermanentID == host.ObjectID {
			captures = append(captures, event.TriggeredAbilities...)
		}
	}
	if len(captures) != 1 || captures[0].SourceID != source.ObjectID ||
		captures[0].SourceCardID != source.CardInstanceID || captures[0].Controller != game.Player2 ||
		!reflect.DeepEqual(captures[0].Ability, &def.TriggeredAbilities[0]) {
		t.Fatal("copied watcher rebound its body/physical source/controller to the blueprint")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("copied event-time watcher was dropped or dispatched twice")
	}
	queued, _ := g.Stack.Peek()
	if queued.SourceID != source.ObjectID || queued.SourceCardID != source.CardInstanceID ||
		queued.Controller != game.Player2 || !reflect.DeepEqual(queued.InlineTrigger, &def.TriggeredAbilities[0]) {
		t.Fatal("queue refetched the physical card's non-watcher printed body")
	}
}
