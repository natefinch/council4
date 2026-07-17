package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func deathKissAttackPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:           game.EventAttackerDeclared,
		Controller:      game.TriggerControllerOpponent,
		Player:          game.TriggerPlayerOpponent,
		AttackRecipient: game.AttackRecipientPlayer,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
}

func deathKissDoubleInstruction() []game.Instruction {
	return []game.Instruction{{Primitive: game.ApplyContinuous{
		Object:   opt.Val(game.EventPermanentReference()),
		Duration: game.DurationUntilEndOfTurn,
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:       game.LayerPowerToughnessModify,
			DoublePower: true,
		}},
	}}}
}

func TestDeathKissAttackRelations(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addTriggeredPermanent(g, game.Player1, deathKissAttackPattern(), nil, nil)
	pattern := deathKissAttackPattern()
	attacker := addCombatPermanent(g, game.Player2, vanillaCreature("Attacker", 2, 2))

	event := game.Event{
		Kind:           game.EventAttackerDeclared,
		Controller:     game.Player2,
		SourceObjectID: attacker.ObjectID,
		PermanentID:    attacker.ObjectID,
		Player:         game.Player3,
		AttackTarget:   game.AttackTarget{Player: game.Player3},
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("opponent creature attacking a direct opponent did not match")
	}
	event.Controller = game.Player1
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("controller's own attacking creature matched")
	}
	event.Controller = game.Player2
	event.Player = game.Player1
	event.AttackTarget = game.AttackTarget{Player: game.Player1}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("creature attacking the source controller matched")
	}
	event.Player = game.Player3
	event.AttackTarget = game.AttackTarget{Player: game.Player3, PlaneswalkerID: g.IDGen.Next()}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("creature attacking a planeswalker matched the direct-player trigger")
	}
}

func TestDeathKissDoubleUsesLivePermanentAcrossControlChangeAndStacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, deathKissAttackPattern(), deathKissDoubleInstruction(), nil)
	addTriggeredPermanent(g, game.Player1, deathKissAttackPattern(), deathKissDoubleInstruction(), nil)
	attacker := addCombatPermanent(g, game.Player2, vanillaCreature("Attacker", 2, 2))

	emitEvent(g, game.Event{
		Kind:           game.EventAttackerDeclared,
		Controller:     game.Player2,
		SourceObjectID: attacker.ObjectID,
		PermanentID:    attacker.ObjectID,
		Player:         game.Player3,
		AttackTarget:   game.AttackTarget{Player: game.Player3},
	})
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatalf("stack size = %d, want two Death Kiss triggers", g.Stack.Size())
	}

	addCountersToPermanent(g, attacker, counter.PlusOnePlusOne, 1)
	attacker.Controller = game.Player4
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, attacker); got != 12 {
		t.Fatalf("power = %d, want live 3 doubled twice to 12", got)
	}
	if got, ok := effectiveToughness(g, attacker); !ok || got != 3 {
		t.Fatalf("toughness = %d/%v, want unchanged 3", got, ok)
	}

	addCountersToPermanent(g, attacker, counter.PlusOnePlusOne, 1)
	if got := effectivePower(g, attacker); got != 16 {
		t.Fatalf("power after later change = %d, want live 4 doubled twice to 16", got)
	}
	if got, ok := effectiveToughness(g, attacker); !ok || got != 4 {
		t.Fatalf("toughness after later change = %d/%v, want unchanged 4", got, ok)
	}
}

func TestMonstrosityEmitsAnnouncedAndReplacementAdjustedAmounts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Doubler",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.CounterPlacementReplacement("double", 2, 0, counter.PlusOnePlusOne, game.TriggerControllerYou),
		},
	}})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Monster",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:  game.EventPermanentBecameMonstrous,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{}.Ability(),
		}},
	}})
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
		XValue:       2,
	}
	resolveInstruction(engine, g, obj, game.Monstrosity{
		Object: game.SourcePermanentReference(),
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
	}, &TurnLog{})

	if !source.Monstrous || source.Counters.Get(counter.PlusOnePlusOne) != 4 {
		t.Fatalf("monstrous=%v counters=%d, want true/4", source.Monstrous, source.Counters.Get(counter.PlusOnePlusOne))
	}
	event, ok := lastEventOfKind(g, game.EventPermanentBecameMonstrous)
	if !ok || event.PermanentID != source.ObjectID || event.Controller != game.Player1 ||
		event.XValue != 2 || event.Amount != 4 {
		t.Fatalf("monstrous event = %#v", event)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("becomes-monstrous trigger stack size = %d, want 1", g.Stack.Size())
	}

	eventCount := countEvents(g, game.EventPermanentBecameMonstrous)
	obj.XValue = 5
	resolveInstruction(engine, g, obj, game.Monstrosity{
		Object: game.SourcePermanentReference(),
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
	}, &TurnLog{})
	if source.Counters.Get(counter.PlusOnePlusOne) != 4 ||
		countEvents(g, game.EventPermanentBecameMonstrous) != eventCount {
		t.Fatal("already-monstrous permanent changed or emitted another event")
	}
}

func TestMonstrosityZeroStillBecomesMonstrous(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, vanillaCreature("Monster", 3, 3))
	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	resolveInstruction(engine, g, obj, game.Monstrosity{
		Object: game.SourcePermanentReference(),
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
	}, &TurnLog{})
	event, ok := lastEventOfKind(g, game.EventPermanentBecameMonstrous)
	if !source.Monstrous || source.Counters.Get(counter.PlusOnePlusOne) != 0 ||
		!ok || event.XValue != 0 || event.Amount != 0 {
		t.Fatalf("source/event = %#v / %#v", source, event)
	}
}

func TestEventXTargetCardinalityAndGoadResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetSpec := game.TargetSpec{
		MinTargets:                  0,
		MaxTargets:                  99,
		MaxTargetsFromTriggerEventX: true,
		Allow:                       game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerOpponent,
		}),
	}
	instructions := []game.Instruction{{Primitive: game.Goad{
		Object: game.AllTargetPermanentsReference(0),
	}}}
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentBecameMonstrous,
		Source: game.TriggerSourceSelf,
	}, instructions, []game.TargetSpec{targetSpec})
	first := addCombatPermanent(g, game.Player2, vanillaCreature("First", 2, 2))
	second := addCombatPermanent(g, game.Player3, vanillaCreature("Second", 2, 2))
	third := addCombatPermanent(g, game.Player4, vanillaCreature("Third", 2, 2))
	_ = addCombatPermanent(g, game.Player1, vanillaCreature("Own", 2, 2))
	body := game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: instructions,
	}.Ability()

	for _, test := range []struct {
		x       int
		wantMax int
	}{
		{x: 0, wantMax: 0},
		{x: 2, wantMax: 2},
		{x: 5, wantMax: 3},
	} {
		choices := targetChoicesForBodyFromSourceObjectWithModes(
			g, game.Player1, nil, source.ObjectID, game.Event{XValue: test.x}, &body, nil,
		)
		if choices.kind != targetLegalChoicesFound {
			t.Fatalf("X=%d choices kind = %v", test.x, choices.kind)
		}
		maxCount := 0
		for i, counts := range choices.targetCounts {
			count := len(choices.choices[i])
			if len(counts) > 0 {
				count = counts[0]
			}
			if count > maxCount {
				maxCount = count
			}
		}
		if maxCount != test.wantMax {
			t.Fatalf("X=%d max target count = %d, want %d", test.x, maxCount, test.wantMax)
		}
	}

	event := game.Event{
		Kind:           game.EventPermanentBecameMonstrous,
		SourceID:       source.CardInstanceID,
		SourceObjectID: source.ObjectID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
		XValue:         2,
	}
	choices := targetChoicesForBodyFromSourceObjectWithModes(
		g, game.Player1, nil, source.ObjectID, event, &body, nil,
	)
	choiceIndex := -1
	for i, choice := range choices.choices {
		if len(choice) == 2 &&
			((choice[0] == game.PermanentTarget(first.ObjectID) &&
				choice[1] == game.PermanentTarget(second.ObjectID)) ||
				(choice[0] == game.PermanentTarget(second.ObjectID) &&
					choice[1] == game.PermanentTarget(first.ObjectID))) {
			choiceIndex = i
			break
		}
	}
	if choiceIndex < 0 {
		t.Fatalf("target choices = %#v, want first+second combination", choices.choices)
	}
	emitEvent(g, event)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{choiceIndex}}},
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("becomes-monstrous trigger was not put on the stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.HasTriggerEvent || obj.TriggerEvent.XValue != 2 ||
		len(obj.Targets) != 2 || obj.TargetCounts[0] != 2 {
		t.Fatalf("trigger stack object = %#v", obj)
	}
	first.Controller = game.Player1
	source.Controller = game.Player4
	engine.resolveTopOfStack(g, &TurnLog{})

	if _, goaded := first.Goaded[game.Player1]; goaded {
		t.Fatal("target that became controlled by the trigger controller was goaded")
	}
	status, goaded := second.Goaded[game.Player1]
	if !goaded || status.ExpiresFor != game.Player1 {
		t.Fatalf("second goad status = %#v, want attributed to/expires for Player1", status)
	}
	if _, goaded := third.Goaded[game.Player1]; goaded {
		t.Fatal("unchosen creature was goaded")
	}
}

func lastEventOfKind(g *game.Game, kind game.EventKind) (game.Event, bool) {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind {
			return g.Events[i], true
		}
	}
	return game.Event{}, false
}
