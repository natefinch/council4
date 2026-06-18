package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestOneOrMoreTriggerCoalescesSimultaneousEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
		OneOrMore:             true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player1)
	simultaneousID := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventPermanentDied, Controller: game.Player1, PermanentID: first.ObjectID, CardID: first.CardInstanceID, SimultaneousID: simultaneousID})
	emitEvent(g, game.Event{Kind: game.EventPermanentDied, Controller: game.Player1, PermanentID: second.ObjectID, CardID: second.CardInstanceID, SimultaneousID: simultaneousID})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one coalesced trigger", got)
	}
}

func TestOneOrMoreTriggerDoesNotCoalesceSequentialEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
		OneOrMore:             true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player1)
	emitEvent(g, game.Event{Kind: game.EventPermanentDied, Controller: game.Player1, PermanentID: first.ObjectID, CardID: first.CardInstanceID, SimultaneousID: g.IDGen.Next()})
	emitEvent(g, game.Event{Kind: game.EventPermanentDied, Controller: game.Player1, PermanentID: second.ObjectID, CardID: second.CardInstanceID, SimultaneousID: g.IDGen.Next()})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more triggers were not put on stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want one trigger for each sequential event", got)
	}
}

func TestOneOrMoreTriggerDoesNotInferBatchFromQueue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:     game.EventCardDiscarded,
		Player:    game.TriggerPlayerYou,
		OneOrMore: true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1})
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more triggers were not put on stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want queued events without a batch ID to remain distinct", got)
	}
}

func TestOneOrMoreAttackTriggerCoalescesPerAttackTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                    game.EventAttackerDeclared,
		Controller:               game.TriggerControllerYou,
		AttackRecipient:          game.AttackRecipientPlayer,
		OneOrMore:                true,
		OneOrMorePerAttackTarget: true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	batchID := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1, Player: game.Player2, AttackTarget: game.AttackTarget{Player: game.Player2}, SimultaneousID: batchID})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1, Player: game.Player2, AttackTarget: game.AttackTarget{Player: game.Player2}, SimultaneousID: batchID})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1, Player: game.Player3, AttackTarget: game.AttackTarget{Player: game.Player3}, SimultaneousID: batchID})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more attack triggers were not put on stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want one trigger for each attacked player", got)
	}
}

func TestOneOrMoreZoneChangeTriggerCoalescesActualSimultaneousMove(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		OneOrMore:     true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player2)

	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{first, second}, zone.Exile) {
		t.Fatal("simultaneous move failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more zone-change trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one coalesced trigger", got)
	}
}

func TestOneOrMoreZoneChangeTriggerDoesNotCoalesceSequentialMoves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		OneOrMore:     true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player2)

	if !movePermanentToZone(g, first, zone.Exile) || !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first sequential move did not trigger")
	}
	if !movePermanentToZone(g, second, zone.Exile) || !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("second sequential move did not trigger")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want one trigger for each sequential move", got)
	}
}

func TestOneOrMoreZoneChangeTriggerDoesNotCoalesceQueuedSequentialMoves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Battlefield,
		OneOrMore:     true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player2)

	if !movePermanentToZone(g, first, zone.Exile) || !movePermanentToZone(g, second, zone.Exile) {
		t.Fatal("sequential moves failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("queued sequential moves did not trigger")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want one trigger for each queued sequential move", got)
	}
}

func TestOneOrMoreEnterTriggerUsesTokenCreationBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:     game.EventPermanentEnteredBattlefield,
		OneOrMore: true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	token := &game.CardDef{CardFace: game.CardFace{
		Name:       "Bat",
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 1}),
		Toughness:  opt.Val(game.PT{Value: 1}),
		OracleText: "Flying",
	}}

	if !createTokenPermanentsWithChoices(engine, g, game.Player1, token, 3, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("token creation batch failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token creation batch did not trigger")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one trigger for one token-creation batch", got)
	}
}

func TestOneOrMoreEnterTriggerDoesNotCoalesceQueuedSequentialEntries(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:     game.EventPermanentEnteredBattlefield,
		OneOrMore: true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	token := &game.CardDef{CardFace: game.CardFace{
		Name:       "Bat",
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 1}),
		Toughness:  opt.Val(game.PT{Value: 1}),
		OracleText: "Flying",
	}}

	if _, ok := createTokenPermanent(g, game.Player1, token); !ok {
		t.Fatal("first token creation failed")
	}
	if _, ok := createTokenPermanent(g, game.Player1, token); !ok {
		t.Fatal("second token creation failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("queued sequential entries did not trigger")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want one trigger for each queued sequential entry", got)
	}
}

func TestFightEventTriggersForControlledFighter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventFight,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		},
	}

	resolveFightTargets(g, obj, 0, 1)

	fights := 0
	for _, event := range g.Events {
		if event.Kind == game.EventFight {
			fights++
		}
	}
	if fights != 2 {
		t.Fatalf("fight events = %d, want one per fighter", fights)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("fight trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want only controlled fighter trigger", got)
	}
}

func TestOneOrMoreFightTriggerCoalescesBothControlledFighters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventFight,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
		OneOrMore:             true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 3)

	resolveFightPermanents(g, first, second)

	var fights []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventFight {
			fights = append(fights, event)
		}
	}
	if len(fights) != 2 {
		t.Fatalf("fight events = %d, want one per fighter", len(fights))
	}
	if fights[0].SimultaneousID == 0 || fights[0].SimultaneousID != fights[1].SimultaneousID {
		t.Fatalf("fight event batch IDs = %v and %v, want one shared nonzero ID", fights[0].SimultaneousID, fights[1].SimultaneousID)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more fight trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one trigger for the fight resolution", got)
	}
}

func TestOneOrMoreFightTriggerKeepsSeparateFightResolutions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventFight,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
		OneOrMore:             true,
	}, nil, nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	third := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	fourth := addCombatCreaturePermanentWithPower(g, game.Player1, 3)

	resolveFightPermanents(g, first, second)
	resolveFightPermanents(g, third, fourth)

	var fightBatchIDs []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventFight {
			fightBatchIDs = append(fightBatchIDs, event.SimultaneousID)
		}
	}
	if len(fightBatchIDs) != 4 {
		t.Fatalf("fight events = %d, want two per fight resolution", len(fightBatchIDs))
	}
	if fightBatchIDs[0] == 0 || fightBatchIDs[0] != fightBatchIDs[1] || fightBatchIDs[2] == 0 || fightBatchIDs[2] != fightBatchIDs[3] || fightBatchIDs[0] == fightBatchIDs[2] {
		t.Fatalf("fight event batch IDs = %v, want a distinct shared nonzero ID per fight resolution", fightBatchIDs)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more fight triggers were not put on stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want one trigger per fight resolution", got)
	}
}

func TestDamageTriggerCanRequireAttackingRecipient(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                      game.EventDamageDealt,
		DamageRecipient:            game.DamageRecipientPermanent,
		DamageRecipientCombatState: game.CombatStateAttacking,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	nonattacking := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	attacking := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacking.ObjectID, Target: game.AttackTarget{Player: game.Player1}}},
	}

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		Controller:      game.Player1,
		PermanentID:     nonattacking.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
		Amount:          1,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("damage trigger matched nonattacking recipient")
	}
	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		Controller:      game.Player1,
		PermanentID:     attacking.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
		Amount:          1,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("damage trigger did not match attacking recipient")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want attacking-recipient trigger to draw one card", got)
	}
}

func TestSpellCastTriggerExcludesCreatureSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventSpellCast,
		Controller:       game.TriggerControllerOpponent,
		ExcludeCardTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	spellID := addCardToHand(g, game.Player2, greenCreature())
	addBasicLandPermanent(g, game.Player2, types.Forest)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast creature failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("noncreature spell trigger fired for creature spell")
	}
}

func TestPermanentTriggerRequireExcludeTypeFilters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		Controller:            game.TriggerControllerOpponent,
		RequirePermanentTypes: []types.Card{types.Artifact},
		ExcludePermanentTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})

	destroyPermanent(g, artifact.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("artifact death trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want artifact death trigger draw", got)
	}
}
