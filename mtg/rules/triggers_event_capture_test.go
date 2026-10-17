package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestEventCapturePrimalCommandWellSurvivesSourceDeparture(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	registry := cards.NewDefaultRegistry()
	well := addCombatPermanent(g, game.Player1, registry.Lookup("Well of Lost Dreams"))
	for range 8 {
		addCardToLibrary(g, game.Player1, vanillaCreature("Library card", 2, 2))
	}
	addImplementationSpellToStack(g, game.Player1, registry.Lookup("Primal Command"), []game.Target{
		game.PlayerTarget(game.Player1), game.PermanentTarget(well.ObjectID),
	})
	spell, _ := g.Stack.Peek()
	spell.ChosenModes, spell.TargetCounts = []int{0, 1}, []int{1, 1}
	e.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 47 || !g.Players[game.Player1].Library.Contains(well.CardInstanceID) ||
		g.Stack.Size() != 0 {
		t.Fatal("Primal Command did not finish both modes before priority")
	}
	branch := g.Clone()
	for _, result := range []*game.Game{g, branch} {
		if !e.putTriggeredAbilitiesOnStack(result) || result.Stack.Size() != 1 {
			t.Fatal("Well's already-triggered ability was lost or duplicated")
		}
		trigger, _ := result.Stack.Peek()
		if trigger.SourceID != well.ObjectID || trigger.TriggerEvent.Amount != 7 {
			t.Fatal("Well lost its incarnation or captured life amount")
		}
		result.Players[game.Player1].ManaPool.Add(mana.C, 10)
		e.resolveTopOfStack(result, &TurnLog{})
		if result.Players[game.Player1].Hand.Size() != 7 || result.Players[game.Player1].ManaPool.Amount(mana.C) != 3 {
			t.Fatal("departed Well did not resolve using the event's seven-life payment bound")
		}
	}
}

func TestEventCaptureDoesNotUseLaterSourceState(t *testing.T) {
	for _, kind := range []game.EventKind{game.EventLifeGained, game.EventPermanentTapped, game.EventScry, game.EventTokenCreated} {
		for _, change := range []string{"leave", "lose ability", "change control", "late source", "late ability"} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: kind},
				[]game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}, nil)
			ability := g.CardInstances[source.CardInstanceID].Def.TriggeredAbilities[0]
			if change == "late source" || change == "late ability" {
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					ID: g.IDGen.Next(), Layer: game.LayerAbility, AffectedObjectID: source.ObjectID, RemoveAllAbilities: true,
				})
			}
			emitEvent(g, game.Event{Kind: kind, Player: game.Player1, Controller: game.Player1,
				PermanentID: source.ObjectID, CardID: source.CardInstanceID, Amount: 1})
			switch change {
			case "leave":
				movePermanentToZone(g, source, zone.Library)
			case "lose ability":
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					ID: g.IDGen.Next(), Layer: game.LayerAbility, AffectedObjectID: source.ObjectID, RemoveAllAbilities: true,
				})
			case "change control":
				source.Controller = game.Player2
			case "late source":
				movePermanentToZone(g, source, zone.Library)
				addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Late watcher", Types: []types.Card{types.Artifact}, TriggeredAbilities: []game.TriggeredAbility{ability},
				}})
			case "late ability":
				g.ContinuousEffects = nil
			default:
				t.Fatalf("unknown change %q", change)
			}
			hasTrigger := e.putTriggeredAbilitiesOnStack(g)
			want := change != "late source" && change != "late ability"
			if hasTrigger != want {
				t.Fatalf("event %v, source %s: triggered=%v, want %v", kind, change, hasTrigger, want)
			}
			if want {
				trigger, _ := g.Stack.Peek()
				if g.Stack.Size() != 1 || trigger.Controller != game.Player1 || trigger.SourceID != source.ObjectID {
					t.Fatalf("event %v reassigned/repeated its captured trigger: %+v", kind, trigger)
				}
			}
		}
	}
}

func TestEventCaptureEveryEmittedKindIsAuthoritative(t *testing.T) {
	for kind := game.EventUnknown; int(kind) < game.EventKindCount; kind++ {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		event := prepareEvent(g, game.Event{Kind: kind})
		if !event.TriggeredAbilitiesCaptured {
			t.Fatalf("event kind %v can still fall through to late detection", kind)
		}
	}
}

func TestEventCaptureDoesNotResurrectDamageWitnessThatLeftBeforeDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	registry := cards.NewDefaultRegistry()
	source := addCombatPermanent(g, game.Player1, registry.Lookup("Prodigal Pyromancer"))
	aura := addCombatPermanent(g, game.Player1, registry.Lookup("Curiosity"))
	if !attachPermanent(g, aura, source) {
		t.Fatal("could not attach Curiosity")
	}
	if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{game.PlayerTarget(game.Player2)}, 0)) {
		t.Fatal("could not activate Pyromancer")
	}
	movePermanentToZone(g, source, zone.Graveyard)
	e.applyStateBasedActions(g)
	if !g.Players[game.Player1].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("Curiosity did not leave before damage")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != 39 || e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("damage restored a trigger belonging to an absent source/attachment")
	}
}

func TestEventCaptureInterveningIfUsesEventThenResolution(t *testing.T) {
	for _, conditionInitiallyTrue := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := NewEngine(nil)
		if conditionInitiallyTrue {
			g.Players[game.Player1].Life = 41
		}
		addTriggeredPermanentWithCondition(g, game.Player1, &game.TriggerPattern{Event: game.EventScry},
			41, []game.Instruction{{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}}})
		emitEvent(g, game.Event{Kind: game.EventScry, Player: game.Player1, Amount: 1})
		g.Players[game.Player1].Life = 41
		if got := e.putTriggeredAbilitiesOnStack(g); got != conditionInitiallyTrue {
			t.Fatalf("intervening-if was reevaluated at placement: %v, want %v", got, conditionInitiallyTrue)
		}
		if conditionInitiallyTrue {
			addCardToLibrary(g, game.Player1, vanillaCreature("Not drawn", 2, 2))
			g.Players[game.Player1].Life = 40
			e.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player1].Hand.Size() != 0 {
				t.Fatal("intervening-if resolution recheck was skipped")
			}
		}
	}
}

func TestEventCaptureSnapshotsChoicesAndSourceMultipliers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventLifeGained}, nil, nil)
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: mana.G},
	}
	g.CardInstances[source.CardInstanceID].Def.TriggeredAbilities[0].Content = game.Mode{
		Sequence: []game.Instruction{{Primitive: game.AddMana{Amount: game.Fixed(1), EntryChoiceFrom: game.EntryColorChoiceKey}}},
	}.Ability()
	sourceDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Event-time source multiplier", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectAdditionalTriggerForControlledPermanent,
			AffectedController: game.ControllerYou,
		}}}},
	}}
	doubler := addCombatPermanent(g, game.Player1, sourceDef)
	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: 2})
	index := len(g.Events) - 1
	source.EntryChoices[game.EntryColorChoiceKey] = game.ResolutionChoiceResult{Kind: game.ResolutionChoiceMana, Color: mana.R}
	movePermanentToZone(g, source, zone.Library)
	movePermanentToZone(g, doubler, zone.Graveyard)
	branch := g.Clone()
	branch.Events[index].TriggeredAbilities[0].EntryChoices[game.EntryColorChoiceKey] =
		game.ResolutionChoiceResult{Kind: game.ResolutionChoiceMana, Color: mana.U}
	for _, result := range []*game.Game{g, branch} {
		if !e.putTriggeredAbilitiesOnStack(result) || result.Stack.Size() != 2 {
			t.Fatal("event-time multiplier was lost after source departure")
		}
		e.resolveTopOfStack(result, &TurnLog{})
		e.resolveTopOfStack(result, &TurnLog{})
		want := mana.G
		if result == branch {
			want = mana.U
		}
		if result.Players[game.Player1].ManaPool.Amount(want) != 2 {
			t.Fatal("captured entry choices were lost or aliased across clones")
		}
	}
}
