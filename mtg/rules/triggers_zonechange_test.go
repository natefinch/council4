package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestNonSelfDiesTriggerControllerFilterFiresOnlyForCorrectController
// exercises the TriggerControllerYou + ExcludeSelf + SubjectSelection path:
// the trigger on Player1's creature must NOT fire when an opponent's creature
// dies and MUST fire when a different Player1-controlled creature dies.
func TestNonSelfDiesTriggerControllerFilterFiresOnlyForCorrectController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	// Source permanent: watches "another creature you control dies".
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:       game.EventPermanentDied,
		Controller:  game.TriggerControllerYou,
		ExcludeSelf: true,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// Opponent's creature dies — trigger must NOT fire.
	opponentCreature := addCombatCreaturePermanent(g, game.Player2)
	destroyPermanent(g, opponentCreature.ObjectID)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("non-self dies trigger fired for an opponent-controlled creature")
	}

	// Another Player1-controlled creature dies — trigger MUST fire once.
	friendlyCreature := addCombatCreaturePermanent(g, game.Player1)
	destroyPermanent(g, friendlyCreature.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("non-self dies trigger did not fire for another friendly creature")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v", obj, source.ObjectID)
	}
}

func TestNonSelfDiesTriggerFiresForSimultaneousDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:       game.EventPermanentDied,
		ExcludeSelf: true,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	other := addCombatCreaturePermanent(g, game.Player1)

	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{source, other}, zone.Graveyard) {
		t.Fatal("simultaneous move failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("departed source did not trigger for another simultaneous death")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.TriggerEvent.PermanentID != other.ObjectID {
		t.Fatalf("top of stack = %+v, want source %v triggered by %v", obj, source.ObjectID, other.ObjectID)
	}
}

// TestNonSelfDiesTriggerExcludeSelfDoesNotFireForSource verifies that a
// non-self dies trigger with ExcludeSelf=true does not fire when the source
// permanent itself dies.
func TestNonSelfDiesTriggerExcludeSelfDoesNotFireForSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:       game.EventPermanentDied,
		ExcludeSelf: true,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	destroyPermanent(g, source.ObjectID)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ExcludeSelf non-self dies trigger fired for its own source permanent")
	}
}

// TestNonSelfDiesTriggerSubjectSelectionCreatureDoesNotMatchNonCreature checks
// that a trigger with SubjectSelection{RequiredTypes: [Creature]} does not fire
// when a non-creature permanent dies.
func TestNonSelfDiesTriggerSubjectSelectionCreatureDoesNotMatchNonCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventPermanentDied,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// Add a non-creature (artifact) and destroy it.
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Some Artifact",
			Types: []types.Card{types.Artifact},
		},
	})
	destroyPermanent(g, artifact.ObjectID)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature SubjectSelection trigger fired when a non-creature permanent died")
	}
}

// TestNonSelfDiesTriggerSubjectSelectionFiresForMatchingCreature verifies that
// a SubjectSelection{RequiredTypes: [Creature]} trigger fires when any creature
// dies — confirming the happy path with LKI-based type matching.
func TestNonSelfDiesTriggerSubjectSelectionFiresForMatchingCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventPermanentDied,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	creature := addCombatCreaturePermanent(g, game.Player2)
	destroyPermanent(g, creature.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("SubjectSelection creature trigger did not fire when a creature died")
	}
}

// TestNonSelfDiesTriggerOpponentControllerFiresOnlyForOpponentCreature
// exercises the TriggerControllerOpponent + SubjectSelection path used by cards
// like Assault Intercessor ("Whenever a creature an opponent controls dies, that
// player loses 2 life"): the trigger must fire when an opponent's creature dies
// and must NOT fire when the source controller's own creature dies.
func TestNonSelfDiesTriggerOpponentControllerFiresOnlyForOpponentCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventPermanentDied,
		Controller: game.TriggerControllerOpponent,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}, []game.Instruction{{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.EventPlayerReference()}}}, nil)

	// The source controller's own creature dies — trigger must NOT fire.
	ownCreature := addCombatCreaturePermanent(g, game.Player1)
	destroyPermanent(g, ownCreature.ObjectID)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent-controller dies trigger fired for the source controller's own creature")
	}

	// An opponent's creature dies — trigger MUST fire once.
	opponentCreature := addCombatCreaturePermanent(g, game.Player2)
	destroyPermanent(g, opponentCreature.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent-controller dies trigger did not fire for an opponent's creature")
	}
}

func TestZoneChangeTriggerSubjectSelectionUsesLastKnownCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	subject := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Legendary Dragon",
		ManaCost:        opt.Val(cost.Mana{cost.O(4)}),
		Colors:          []color.Color{color.Green},
		Supertypes:      []types.Super{types.Legendary},
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Dragon},
		Power:           opt.Val(game.PT{Value: 4}),
		Toughness:       opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}})
	subject.Tapped = true
	if !movePermanentToZone(g, subject, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}
	event := g.Events[len(g.Events)-2]
	if event.Kind != game.EventZoneChanged || event.PermanentID != subject.ObjectID {
		t.Fatalf("zone-change event = %+v", event)
	}
	pattern := &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		MatchToZone:   true,
		ToZone:        zone.Graveyard,
		SubjectSelection: game.Selection{
			Supertypes:  []types.Super{types.Legendary},
			SubtypesAny: []types.Sub{types.Dragon},
			ColorsAny:   []color.Color{color.Green},
			Tapped:      game.TriTrue,
			Keyword:     game.Flying,
			ManaValue:   opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
			Power:       opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
			Toughness:   opt.Val(compare.Int{Op: compare.Equal, Value: 4}),
		},
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("zone-change pattern did not match last-known characteristics")
	}
	pattern.SubjectSelection.Power = opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5})
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("zone-change pattern matched a last-known power near miss")
	}
}

func TestZoneChangeTriggerMatchesFaceDownCombatLKIAndExcludedDestination(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	attacker := addCombatCreaturePermanent(g, game.Player2)
	attacker.FaceDown = true
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID}},
	}
	if !movePermanentToZone(g, attacker, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}
	diedEvent := g.Events[len(g.Events)-1]
	if diedEvent.Kind != game.EventPermanentDied || !diedEvent.FaceDown {
		t.Fatalf("died event = %+v, want face-down permanent-died event", diedEvent)
	}
	pattern := &game.TriggerPattern{
		Event:         game.EventPermanentDied,
		MatchFaceDown: true,
		FaceDown:      true,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateAttacking,
		},
	}
	if !triggerMatchesEvent(g, source, pattern, diedEvent) {
		t.Fatal("died pattern did not match face-down attacking last-known information")
	}
	pattern.FaceDown = false
	if triggerMatchesEvent(g, source, pattern, diedEvent) {
		t.Fatal("face-up near-miss pattern matched face-down event")
	}

	zoneEvent := g.Events[len(g.Events)-2]
	leaveWithoutDying := &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		ExcludeToZone: true,
		ToZone:        zone.Graveyard,
	}
	if triggerMatchesEvent(g, source, leaveWithoutDying, zoneEvent) {
		t.Fatal("leave-without-dying pattern matched a move to the graveyard")
	}
	exiled := addCombatCreaturePermanent(g, game.Player2)
	if !movePermanentToZone(g, exiled, zone.Exile) {
		t.Fatal("movePermanentToZone to exile failed")
	}
	exileEvent := g.Events[len(g.Events)-1]
	if !triggerMatchesEvent(g, source, leaveWithoutDying, exileEvent) {
		t.Fatal("leave-without-dying pattern did not match a move to exile")
	}
}

func TestPermanentZoneChangeTriggerRejectsCardOnlyZoneChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	event := game.Event{
		Kind:       game.EventZoneChanged,
		Controller: game.Player2,
		Player:     game.Player2,
		FromZone:   zone.Battlefield,
		ToZone:     zone.Graveyard,
	}
	pattern := &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("permanent zone-change pattern matched a card-only zone change")
	}
}

func TestSelfZoneChangeTriggerUsesDepartedSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		Source:        game.TriggerSourceSelf,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		MatchToZone:   true,
		ToZone:        zone.Exile,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	if !movePermanentToZone(g, source, zone.Exile) {
		t.Fatal("movePermanentToZone failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("departed source zone-change trigger was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.TriggerEvent.PermanentID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want departed source trigger", obj)
	}
}

// addAllyCreaturePermanent adds a creature with the Ally subtype so
// self-or-another subject-selection tests can distinguish matching from
// non-matching permanents.
func addAllyCreaturePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Ally Creature",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Sub("Ally")},
	}})
}

// selfOrAnotherEntersPattern models "Whenever this creature or another Ally you
// control enters, ...": the union of the source itself and another Ally the
// source's controller controls.
func selfOrAnotherEntersPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:                  game.EventPermanentEnteredBattlefield,
		Controller:             game.TriggerControllerYou,
		SubjectSelectionOrSelf: true,
		SubjectSelection: game.Selection{
			SubtypesAny: []types.Sub{types.Sub("Ally")},
		},
	}
}

// TestSelfOrAnotherEntersTriggerFiresForMatchingOther verifies the union fires
// for a different Ally the controller controls.
func TestSelfOrAnotherEntersTriggerFiresForMatchingOther(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, selfOrAnotherEntersPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	other := addAllyCreaturePermanent(g, game.Player1)
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: other.ObjectID,
		CardID:      other.CardInstanceID,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-or-another trigger did not fire for another matching Ally")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.TriggerEvent.PermanentID != other.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v for %v", obj, source.ObjectID, other.ObjectID)
	}
}

// TestSelfOrAnotherEntersTriggerFiresForSource verifies the union fires for the
// source itself, even though the source is not an Ally and the self-excluding
// "another" wording would otherwise reject it.
func TestSelfOrAnotherEntersTriggerFiresForSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, selfOrAnotherEntersPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: source.ObjectID,
		CardID:      source.CardInstanceID,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-or-another trigger did not fire for its own source")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.TriggerEvent.PermanentID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want self trigger from source %v", obj, source.ObjectID)
	}
}

// TestSelfOrAnotherEntersTriggerDoesNotFireForNonMatching verifies the union
// does not fire for a controlled creature that does not match the selection and
// is not the source.
func TestSelfOrAnotherEntersTriggerDoesNotFireForNonMatching(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, selfOrAnotherEntersPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	nonAlly := addCombatCreaturePermanent(g, game.Player1)
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: nonAlly.ObjectID,
		CardID:      nonAlly.CardInstanceID,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-or-another trigger fired for a non-matching, non-source creature")
	}
}

// TestSelfOrAnotherEntersTriggerRespectsController verifies the "you control"
// relation still rejects a matching Ally an opponent controls.
func TestSelfOrAnotherEntersTriggerRespectsController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, selfOrAnotherEntersPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	opponentAlly := addAllyCreaturePermanent(g, game.Player2)
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player2,
		PermanentID: opponentAlly.ObjectID,
		CardID:      opponentAlly.CardInstanceID,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-or-another trigger fired for an Ally an opponent controls")
	}
}

func TestAttachedZoneChangeTriggerUsesDepartedSubjectLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	subject := addCombatCreaturePermanent(g, game.Player2)
	equipment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
	}})
	if !attachPermanent(g, equipment, subject) {
		t.Fatal("attachPermanent failed")
	}
	if !movePermanentToZone(g, subject, zone.Hand) {
		t.Fatal("movePermanentToZone failed")
	}
	event := g.Events[len(g.Events)-1]
	pattern := &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		Source:        game.TriggerSourceAttachedPermanent,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		MatchToZone:   true,
		ToZone:        zone.Hand,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
	if !triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached zone-change trigger did not use departed subject LKI")
	}
}

func TestDrawTriggerYouFiresForControllerDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// Controller draws → trigger fires
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger did not fire for controller draw")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 after controller draw", got)
	}

	// Opponent draws → trigger must not fire
	g.Stack = game.Stack{}
	g.Events = nil
	g.TriggerEventCursor = 0
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player2})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger fired for opponent draw but should not")
	}
}

func TestDrawTriggerOpponentFiresForOpponentDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerOpponent,
	}, []game.Instruction{{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.EventPlayerReference()}}}, nil)

	// Opponent draws → trigger fires
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player2})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger did not fire for opponent draw")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 after opponent draw", got)
	}
	before := g.Players[game.Player2].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != before-2 {
		t.Fatalf("opponent life = %d, want %d after event-player life loss", got, before-2)
	}

	// Controller draws → trigger must not fire
	g.Stack = game.Stack{}
	g.Events = nil
	g.TriggerEventCursor = 0
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger fired for controller draw but should not")
	}
}

func TestSpellCastTriggerEventPlayerUsesCaster(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerOpponent,
	}, []game.Instruction{{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.EventPlayerReference()}}}, nil)

	before := g.Players[game.Player2].Life
	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player2})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("spell-cast trigger did not fire for opponent")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != before-2 {
		t.Fatalf("caster life = %d, want %d after event-player life loss", got, before-2)
	}
}

func TestDrawTriggerAnyPlayerFiresForBoth(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerAny,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// Controller draws → trigger fires
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger did not fire for controller draw with TriggerPlayerAny")
	}

	g.Stack = game.Stack{}
	g.Events = nil
	g.TriggerEventCursor = 0

	// Opponent draws → trigger fires
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player2})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger did not fire for opponent draw with TriggerPlayerAny")
	}
}

func TestDiscardTriggerYouFiresForControllerDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDiscarded,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// Controller discards → trigger fires
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("discard trigger did not fire for controller discard")
	}

	// Opponent discards → trigger must not fire
	g.Stack = game.Stack{}
	g.Events = nil
	g.TriggerEventCursor = 0
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player2})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("discard trigger fired for opponent discard but should not")
	}
}

func TestDiscardTriggerOpponentFiresForOpponentDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDiscarded,
		Player: game.TriggerPlayerOpponent,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// Opponent discards → trigger fires
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player2})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("discard trigger did not fire for opponent discard")
	}

	// Controller discards → trigger must not fire
	g.Stack = game.Stack{}
	g.Events = nil
	g.TriggerEventCursor = 0
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("discard trigger fired for controller discard but should not")
	}
}

func TestDiscardOneOrMoreCoalesces(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:     game.EventCardDiscarded,
		Player:    game.TriggerPlayerYou,
		OneOrMore: true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// Three events from one explicit discard batch trigger exactly once.
	simultaneousID := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1, SimultaneousID: simultaneousID})
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1, SimultaneousID: simultaneousID})
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1, SimultaneousID: simultaneousID})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("discard one-or-more trigger did not fire")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 coalesced trigger for three discards", got)
	}
}

// TestZoneChangeTriggerChosenTypeSubjectGatesOnSourceEntryChoice exercises the
// "a creature you control of the chosen type" subject filter (Kindred
// Discovery): the trigger must fire only when the entering creature shares the
// subtype the trigger's source chose as it entered, read from the source's
// EntryChoices[EntryTypeChoiceKey].
func TestZoneChangeTriggerChosenTypeSubjectGatesOnSourceEntryChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pattern := &game.TriggerPattern{
		Event:      game.EventPermanentEnteredBattlefield,
		Controller: game.TriggerControllerYou,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			SubtypeChoice: game.SubtypeChoiceSourceEntry,
		},
	}
	source := addTriggeredPermanent(g, game.Player1, pattern, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Goblin},
	}

	goblin := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Goblin Guy", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Goblin}}})
	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: goblin.ObjectID,
		CardID:      goblin.CardInstanceID,
	}) {
		t.Fatal("chosen-type trigger did not match a creature of the chosen subtype")
	}

	elf := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Elf Guy", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf}}})
	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: elf.ObjectID,
		CardID:      elf.CardInstanceID,
	}) {
		t.Fatal("chosen-type trigger matched a creature whose subtype is not the chosen type")
	}
}

// addArtifactPermanent adds an Artifact permanent so the self-or-another
// battlefield-to-graveyard union tests can distinguish a matching artifact from
// a non-artifact creature.
func addArtifactPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Scrap Artifact",
		Types: []types.Card{types.Artifact},
	}})
}

// selfGraveyardOrAnotherArtifactPattern models Scrap Trawler's "Whenever this
// creature dies or another artifact you control is put into a graveyard from the
// battlefield, ..." as the compiler lowers it: a battlefield-to-graveyard zone
// change filtered to artifacts the source's controller controls, widened to the
// source itself through SubjectSelectionOrSelf.
func selfGraveyardOrAnotherArtifactPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:                  game.EventZoneChanged,
		Controller:             game.TriggerControllerYou,
		SubjectSelectionOrSelf: true,
		MatchFromZone:          true,
		FromZone:               zone.Battlefield,
		MatchToZone:            true,
		ToZone:                 zone.Graveyard,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Artifact},
		},
	}
}

// TestSelfGraveyardOrAnotherArtifactTriggerFiresForAnotherArtifact verifies the
// union fires when a different artifact the controller controls is put into a
// graveyard from the battlefield.
func TestSelfGraveyardOrAnotherArtifactTriggerFiresForAnotherArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, selfGraveyardOrAnotherArtifactPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	other := addArtifactPermanent(g, game.Player1)
	if !movePermanentToZone(g, other, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-or-another graveyard trigger did not fire for another artifact")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.TriggerEvent.PermanentID != other.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v for %v", obj, source.ObjectID, other.ObjectID)
	}
}

// TestSelfGraveyardOrAnotherArtifactTriggerFiresForSource verifies the union
// fires when the source itself is put into a graveyard from the battlefield
// (i.e. dies), even though the self-excluding "another" wording would otherwise
// reject it.
func TestSelfGraveyardOrAnotherArtifactTriggerFiresForSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, selfGraveyardOrAnotherArtifactPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-or-another graveyard trigger did not fire for its own source dying")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.TriggerEvent.PermanentID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want self trigger from source %v", obj, source.ObjectID)
	}
}

// TestSelfGraveyardOrAnotherArtifactTriggerDoesNotFireForNonArtifact verifies
// the union does not fire when a non-artifact creature the controller controls
// dies, since it neither matches the artifact selection nor is the source.
func TestSelfGraveyardOrAnotherArtifactTriggerDoesNotFireForNonArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, selfGraveyardOrAnotherArtifactPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	creature := addCombatCreaturePermanent(g, game.Player1)
	if !movePermanentToZone(g, creature, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-or-another graveyard trigger fired for a non-artifact, non-source creature")
	}
}

// graveyardLeaveYouPattern models "Whenever one or more cards leave your
// graveyard, ..." as the compiler lowers it: an any-card zone change whose
// origin graveyard belongs to the source's controller.
func graveyardLeaveYouPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		Player:        game.TriggerPlayerYou,
		MatchFromZone: true,
		FromZone:      zone.Graveyard,
		OneOrMore:     true,
	}
}

// TestGraveyardLeaveYouTriggerFiresWhenControllerCardLeavesGraveyard verifies
// the trigger fires when a card the controller owns leaves the controller's
// graveyard for another zone.
func TestGraveyardLeaveYouTriggerFiresWhenControllerCardLeavesGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, graveyardLeaveYouPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Graveyard Card"}})
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Hand) {
		t.Fatal("moveCardBetweenZones failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("graveyard-leave trigger did not fire when controller's card left their graveyard")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v", obj, source.ObjectID)
	}
}

// TestGraveyardLeaveYouTriggerDoesNotFireForOpponentGraveyard verifies the
// "your graveyard" scoping: a card leaving an opponent's graveyard must not
// fire the controller's trigger.
func TestGraveyardLeaveYouTriggerDoesNotFireForOpponentGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, graveyardLeaveYouPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opponent Graveyard Card"}})
	if !moveCardBetweenZones(g, game.Player2, cardID, zone.Graveyard, zone.Hand) {
		t.Fatal("moveCardBetweenZones failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("graveyard-leave trigger fired when an opponent's card left their graveyard")
	}
}

// graveyardLeaveCreatureCardPattern models "Whenever one or more creature cards
// leave your graveyard, ..." — the any-creature-card subject form whose origin
// graveyard belongs to the source's controller.
func graveyardLeaveCreatureCardPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		Player:        game.TriggerPlayerYou,
		MatchFromZone: true,
		FromZone:      zone.Graveyard,
		OneOrMore:     true,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
}

// TestGraveyardLeaveCreatureCardTriggerFiresOnlyForCreatureCards verifies the
// typed-subject path: the trigger fires when a creature card leaves the
// controller's graveyard but not when a noncreature card does.
func TestGraveyardLeaveCreatureCardTriggerFiresOnlyForCreatureCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, graveyardLeaveCreatureCardPattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	instant := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Instant",
		Types: []types.Card{types.Instant},
	}})
	if !moveCardBetweenZones(g, game.Player1, instant, zone.Graveyard, zone.Hand) {
		t.Fatal("moveCardBetweenZones failed for instant")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature-card graveyard-leave trigger fired for a noncreature card")
	}

	creature := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Creature",
		Types: []types.Card{types.Creature},
	}})
	if !moveCardBetweenZones(g, game.Player1, creature, zone.Graveyard, zone.Hand) {
		t.Fatal("moveCardBetweenZones failed for creature")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature-card graveyard-leave trigger did not fire for a creature card")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v", obj, source.ObjectID)
	}
}

// graveyardPutIntoFromAnywhereCreaturePattern models "Whenever a creature card
// is put into your graveyard from anywhere, ..." — a card move into the
// source controller's graveyard with no origin-zone constraint, so it fires for
// deaths, mills, and discards alike.
func graveyardPutIntoFromAnywhereCreaturePattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:       game.EventZoneChanged,
		Player:      game.TriggerPlayerYou,
		MatchToZone: true,
		ToZone:      zone.Graveyard,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
}

// TestGraveyardPutIntoFromAnywhereTriggerFiresOnMillAndDeath verifies the
// "put into your graveyard from anywhere" form fires when a creature card is
// milled (library to graveyard) and when a creature dies (battlefield to
// graveyard), but not for a noncreature card or an opponent's graveyard.
func TestGraveyardPutIntoFromAnywhereTriggerFiresOnMillAndDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, graveyardPutIntoFromAnywhereCreaturePattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// A noncreature card milled into your graveyard must NOT fire.
	land := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Wastes", Types: []types.Card{types.Land}}})
	if !moveCardBetweenZones(g, game.Player1, land, zone.Library, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones failed for land")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("put-into-graveyard creature trigger fired for a land card")
	}

	// A creature card milled into an opponent's graveyard must NOT fire.
	oppCreature := addCardToLibrary(g, game.Player2, greenCreature())
	if !moveCardBetweenZones(g, game.Player2, oppCreature, zone.Library, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones failed for opponent creature")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("put-into-your-graveyard trigger fired for an opponent's graveyard")
	}

	// A creature card milled into your graveyard MUST fire.
	creature := addCardToLibrary(g, game.Player1, greenCreature())
	if !moveCardBetweenZones(g, game.Player1, creature, zone.Library, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones failed for creature")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("put-into-graveyard trigger did not fire for a milled creature card")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v", obj, source.ObjectID)
	}
	g.Stack.Pop()

	// A creature dying (battlefield to your graveyard) MUST also fire.
	dying := addCombatCreaturePermanent(g, game.Player1)
	destroyPermanent(g, dying.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("put-into-graveyard trigger did not fire for a creature death")
	}
}

// graveyardPutIntoExcludingBattlefieldCreaturePattern models "Whenever a
// creature card is put into a graveyard from anywhere other than the
// battlefield, ..." — a card move into the graveyard whose origin must not be
// the battlefield, so it fires for mills and discards but not for deaths.
func graveyardPutIntoExcludingBattlefieldCreaturePattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:           game.EventZoneChanged,
		MatchToZone:     true,
		ToZone:          zone.Graveyard,
		ExcludeFromZone: true,
		FromZone:        zone.Battlefield,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
}

// TestGraveyardPutIntoExcludingBattlefieldTriggerSkipsDeaths verifies the "from
// anywhere other than the battlefield" form fires when a creature card is milled
// (library to graveyard) but not when a creature dies (battlefield to
// graveyard).
func TestGraveyardPutIntoExcludingBattlefieldTriggerSkipsDeaths(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, graveyardPutIntoExcludingBattlefieldCreaturePattern(),
		[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	// A creature dying (battlefield to graveyard) must NOT fire.
	dying := addCombatCreaturePermanent(g, game.Player1)
	destroyPermanent(g, dying.ObjectID)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("excluding-battlefield trigger fired for a creature death")
	}

	// A creature card milled into the graveyard (library to graveyard) MUST fire.
	creature := addCardToLibrary(g, game.Player1, greenCreature())
	if !moveCardBetweenZones(g, game.Player1, creature, zone.Library, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones failed for milled creature")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("excluding-battlefield trigger did not fire for a milled creature card")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v", obj, source.ObjectID)
	}
}
