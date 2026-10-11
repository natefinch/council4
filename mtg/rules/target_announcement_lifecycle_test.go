package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestTargetAnnouncementSnapshotSurvivesGrantDepartureAndCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	grant := addCombatPermanent(g, game.Player4, targetAnnouncementGrant(t))
	grant.Controller = game.Player2
	recipient := addCombatPermanent(g, game.Player1, vanillaCreature("Recipient", 2, 3))
	recipient.Controller = game.Player2
	observed := addCardToLibrary(g, game.Player2, vanillaCreature("Original Observation", 1, 1))
	copiedObservation := addCardToLibrary(g, game.Player3, vanillaCreature("Copy Observation", 1, 1))
	targeter := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player3,
		Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)},
	}
	g.Stack.Push(targeter)
	emitTargetEvents(g, targeter)
	if !movePermanentToZone(g, grant, zone.Graveyard) {
		t.Fatal("grant departure failed")
	}
	recipient.Controller = game.Player4
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("event-time granted trigger disappeared before queue")
	}
	trigger, _ := g.Stack.Peek()
	copiedTrigger := game.NewStackObjectCopy(trigger, g.IDGen.Next())
	copiedTrigger.Controller = game.Player3
	g.Stack.Push(copiedTrigger)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player3].Hand.Contains(copiedObservation) ||
		!g.Players[game.Player2].Library.Contains(observed) {
		t.Fatal("copied captured body lost its own controller/local observed product")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Hand.Contains(observed) {
		t.Fatal("original captured body used the new source controller or physical owner")
	}
	g.Stack.Pop()
	g.Stack.Push(targeter)
	emitTargetEvents(g, targeter)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("absent grant triggered on a later event")
	}
}

func TestTargetAnnouncementRetargetAndCopyPublishExactlyOnce(t *testing.T) {
	for _, route := range []string{"ability", "spell"} {
		for _, retargetCopy := range []bool{false, true} {
			t.Run(route+"/"+map[bool]string{false: "retained copy target", true: "changed copy target"}[retargetCopy], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				g.Turn.PriorityPlayer = game.Player3
				grant := addCombatPermanent(g, game.Player4, targetAnnouncementGrant(t))
				grant.Controller = game.Player2
				a := addCombatPermanent(g, game.Player1, vanillaCreature("Recipient A", 2, 3))
				a.Controller = game.Player2
				b := addCombatPermanent(g, game.Player3, vanillaCreature("Recipient B", 2, 3))
				b.Controller = game.Player2
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Retarget Actuator", Layout: "normal", TypeLine: "Artifact",
					OracleText: "{0}: Target creature gets +1/+1 until end of turn.",
				})
				actuator := addCombatPermanent(g, game.Player3, def)
				if route == "ability" {
					if !engine.applyActivateAbilityWithChoices(g, game.Player3, action.ActivateAbilityAction{
						SourceID: actuator.ObjectID, Targets: []game.Target{game.PermanentTarget(a.ObjectID)},
					}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
						t.Fatal("legal original targeted activation failed")
					}
				} else {
					spellDef := compileUnlessCard(t, cardgen.ScryfallCard{
						Name: "Retarget Spell", Layout: "normal", TypeLine: "Instant", ManaCost: "{0}",
						OracleText: "Target creature gets +1/+1 until end of turn.",
					})
					cardID := addCardToHand(g, game.Player3, spellDef)
					if !engine.applyCastSpellWithChoices(g, game.Player3, action.CastSpellAction{
						CardID: cardID, Targets: []game.Target{game.PermanentTarget(a.ObjectID)},
					}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
						t.Fatal("legal original targeted spell cast failed")
					}
				}
				obj, _ := g.Stack.Peek()
				agents := [game.NumPlayers]PlayerAgent{game.Player3: &targetAnnouncementCostAgent{target: b.ObjectID}}
				if !engine.retargetStackObjectChoice(g, game.Player3, obj, agents, &TurnLog{}) {
					t.Fatal("legal retarget failed")
				}
				if targetAnnouncementCount(g, obj.ID) != 2 {
					t.Fatal("changed target did not publish one new occurrence")
				}
				if !engine.retargetStackObjectChoice(g, game.Player3, obj, agents, &TurnLog{}) ||
					targetAnnouncementCount(g, obj.ID) != 2 {
					t.Fatal("unchanged target emitted a duplicate occurrence")
				}
				crimes := func() int {
					count := 0
					for _, event := range g.Events {
						if event.Kind == game.EventCrimeCommitted {
							count++
						}
					}
					return count
				}
				if crimes() != 1 {
					t.Fatal("original announcement/retarget did not preserve exactly one crime")
				}
				copyChooser := &targetAnnouncementCostAgent{target: a.ObjectID}
				addEffectSpellToStack(g, game.Player4, game.CopyStackObject{
					Object: game.TargetStackObjectReference(0), MayChooseNewTargets: retargetCopy,
				}, []game.Target{game.StackObjectTarget(obj.ID)})
				engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player4: copyChooser}, &TurnLog{})
				copiedObject, _ := g.Stack.Peek()
				if !copiedObject.Copy || copiedObject.ID == obj.ID || copiedObject.Controller != game.Player4 ||
					targetAnnouncementCount(g, copiedObject.ID) != 1 {
					t.Fatal("normal copy producer did not publish exactly one final target occurrence")
				}
				if crimes() != 1 {
					t.Fatal("copy target announcement incorrectly committed another crime")
				}
				want := b.ObjectID
				if retargetCopy {
					want = a.ObjectID
				}
				if len(copiedObject.Targets) != 1 || copiedObject.Targets[0].PermanentID != want {
					t.Fatal("copy target selection changed the original or ignored legal new targets")
				}
				events := prepareTargetEvents(g, copiedObject)
				if len(events) != 1 || len(events[0].TriggeredAbilities) != 1 ||
					events[0].TriggeredAbilities[0].Controller != game.Player2 ||
					events[0].TriggeredAbilities[0].SourceID != want {
					t.Fatal("copier controller replaced recipient-local granted trigger ownership")
				}
			})
		}
	}
}

func TestTargetAnnouncementRecipientLimitResetsAtActualNextTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	grant := addCombatPermanent(g, game.Player4, targetAnnouncementGrant(t))
	grant.Controller = game.Player2
	a := addCombatPermanent(g, game.Player1, vanillaCreature("A", 2, 3))
	a.Controller = game.Player2
	b := addCombatPermanent(g, game.Player3, vanillaCreature("B", 2, 3))
	b.Controller = game.Player2
	check := func(recipient *game.Permanent, expected bool) {
		t.Helper()
		obj := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player3,
			Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)},
		}
		g.Stack.Push(obj)
		emitTargetEvents(g, obj)
		if engine.putTriggeredAbilitiesOnStack(g) != expected {
			t.Fatalf("recipient %d trigger expected=%v", recipient.ObjectID, expected)
		}
		if expected {
			engine.resolveTopOfStack(g, &TurnLog{})
		}
		g.Stack.Pop()
	}
	for _, recipient := range []*game.Permanent{a, b} {
		check(recipient, true)
		check(recipient, true)
		check(recipient, false)
	}
	if len(g.Events) == 0 || g.Players[game.Player2].Hand.Size() != 0 {
		t.Fatal("empty observations invented cards")
	}
	engine.advanceToNextTurn(g)
	check(a, true)
	check(b, true)
}
