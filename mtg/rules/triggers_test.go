package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestSelfETBTriggerGoesOnStackAndResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spellID := addCardToHand(g, game.Player1, triggeredCreature(&game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast triggered creature failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ETB trigger was not put on stack")
	}

	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility || obj.SourceCardID != spellID {
		t.Fatalf("top of stack = %+v, want triggered ability from %v", obj, spellID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want trigger to draw one card", got)
	}
}

func TestDeathTriggerGoesOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		RequirePermanentTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	creature := addCombatCreaturePermanent(g, game.Player2)

	destroyPermanent(g, creature.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("death trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want death trigger to draw one card", got)
	}
}

func TestInlineTriggeredAbilityRechecksProtectionFromSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Red Trigger Source",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Red},
	}})
	target := addProtectionFromColorPermanent(g, game.Player2, color.Red)
	trigger := game.TriggeredAbility{
		Content: game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
			Sequence: []game.Instruction{{
				Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)},
			}},
		}.Ability(),
	}
	obj := &game.StackObject{
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    game.Player1,
		InlineTrigger: &trigger,
		Targets:       []game.Target{game.PermanentTarget(target.ObjectID)},
	}

	if got := engine.resolveStackObject(g, obj, &TurnLog{}); got != "countered by rules" {
		t.Fatalf("resolution = %q, want countered by rules", got)
	}
	if target.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want no damage through protection", target.MarkedDamage)
	}
}

func TestTriggerMovesCountersFromEventPermanentLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	destination := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Target Relic",
		Types: []types.Card{types.Artifact}},
	})
	addCounterTransferTriggerSource(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dying Relic",
		Types: []types.Card{types.Artifact}},
	})
	source.Counters.Add(counter.PlusOnePlusOne, 2)
	source.Counters.Add(counter.Charge, 3)

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("counter transfer trigger was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.HasTriggerEvent || obj.TriggerEvent.PermanentID != source.ObjectID {
		t.Fatalf("trigger event = %+v, want event for source %v", obj, source.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := destination.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("destination +1/+1 counters = %d, want 2", got)
	}
	if got := destination.Counters.Get(counter.Charge); got != 3 {
		t.Fatalf("destination charge counters = %d, want 3", got)
	}
}

func TestTriggerEffectCanReferenceEventPermanentOnBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentEnteredBattlefield,
		RequirePermanentTypes: []types.Card{types.Creature},
	}, []game.Instruction{{
		Primitive: game.ApplyContinuous{
			Object:   opt.Val(game.EventPermanentReference()),
			Duration: game.DurationUntilEndOfTurn,
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:      game.LayerPowerToughnessModify,
					PowerDelta: 2,
				},
				{
					Layer:       game.LayerAbility,
					AddKeywords: []game.Keyword{game.Haste},
				},
			},
		},
	}}, nil)
	cardID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Entering Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	card := g.CardInstances[cardID]
	g.Players[game.Player2].Hand.Remove(cardID)
	permanent, ok := createCardPermanentFaceWithChoices(engine, g, card, game.Player2, zone.Hand, game.FaceFront, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !ok {
		t.Fatal("create permanent failed")
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ETB trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, permanent); got != 3 {
		t.Fatalf("effective power = %d, want 3", got)
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("event permanent did not gain haste")
	}
}

func TestTriggerPatternCanRequireStackSpellTargetingSelf(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventObjectBecameTarget,
		Source:               game.TriggerSourceSelf,
		MatchStackObjectKind: true,
		StackObjectKind:      game.StackSpell,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	g.Stack.Push(spell)

	emitTargetEvents(g, spell)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("spell-target trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want spell-target trigger to draw one card", got)
	}
}

func TestTriggerPatternStackSpellDoesNotMatchAbilityTargetingSelf(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventObjectBecameTarget,
		Source:               game.TriggerSourceSelf,
		MatchStackObjectKind: true,
		StackObjectKind:      game.StackSpell,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	ability := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	g.Stack.Push(ability)

	emitTargetEvents(g, ability)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("spell-target trigger matched an activated ability")
	}
}

func TestExaltedTriggersForCreatureAttackingAlone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Exalted Source",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 0}),
		Toughness:       opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.ExaltedStaticBody}},
	})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	emitEvent(g, game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: attacker.ObjectID,
	})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("exalted trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, attacker); got != 3 {
		t.Fatalf("effective power = %d, want 3 after exalted", got)
	}
}

func TestExaltedDoesNotTriggerForMultipleAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Exalted Source",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 0}),
		Toughness:       opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.ExaltedStaticBody}},
	})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	emitEvent(g, game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: first.ObjectID,
	})

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("exalted trigger was put on stack for multiple attackers")
	}
}

func TestDeathTriggerCanUseEventPermanentLKIAndReturnEventCardAsEnchantment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	triggerSource := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventPermanentDied,
	}, []game.Instruction{{
		Primitive: game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:    game.LayerType,
				SetTypes: []types.Card{types.Enchantment},
			}},
		},
	}}, nil)
	triggerCard, ok := g.GetCardInstance(triggerSource.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card not found")
	}
	triggerCard.Def.TriggeredAbilities[0].Trigger.InterveningCondition = opt.Val(game.Condition{
		Object: opt.Val(game.EventPermanentReference()),
		Types:  []types.Card{types.Creature},
	})
	source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Enduring Creature",
		Types:     []types.Card{types.Creature, types.Enchantment},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3})},
	})
	cardID := source.CardInstanceID

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("death trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent := permanentByCardID(g, cardID)
	if permanent == nil {
		t.Fatal("event card was not returned to the battlefield")
	}
	if permanent.Controller != game.Player2 {
		t.Fatalf("returned permanent controller = %v, want owner %v", permanent.Controller, game.Player2)
	}
	if !permanentHasType(g, permanent, types.Enchantment) {
		t.Fatal("returned permanent is not an enchantment")
	}
	if permanentHasType(g, permanent, types.Creature) {
		t.Fatal("returned permanent is still a creature")
	}
}

func TestSelfDiesTriggerMovesEventCardFromGraveyardToOwnersHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, selfDiesEventCardDefinition(game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceEvent},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}))
	cardID := source.CardInstanceID

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-dies return trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("event card was not moved to its owner's hand")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("event card remained in graveyard")
	}
}

func TestSelfDiesEventCardMoveRequiresExpectedZoneAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, selfDiesEventCardDefinition(game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceEvent},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}))
	cardID := source.CardInstanceID

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-dies return trigger was not put on stack")
	}
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("moving event card before trigger resolution failed")
	}
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) {
		t.Fatal("returning event card before trigger resolution failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("event-card effect moved a new graveyard object")
	}
	if g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("event card incorrectly moved to hand from exile")
	}
}

func TestSelfDiesTriggerGrantsOnlyEventCardAdventureCastFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, selfDiesAdventureDefinition())
	cardID := source.CardInstanceID
	otherID := addCardToGraveyard(g, game.Player1, selfDiesAdventureDefinition())

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-dies cast-permission trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	alternateCast := action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceAlternate, nil, 0, nil)
	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, alternateCast) {
		t.Fatalf("legal actions = %+v, want event card Adventure cast", legal)
	}
	if actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceFront, nil, 0, nil)) {
		t.Fatal("permission allowed event card's front face from graveyard")
	}
	if actionsContain(legal, action.CastSpellFaceFromZone(otherID, zone.Graveyard, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("permission allowed a different Adventure card from graveyard")
	}
	if !engine.applyAction(g, game.Player1, alternateCast) {
		t.Fatal("casting permitted Adventure face failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.SourceZone != zone.Graveyard || obj.Face != game.FaceAlternate {
		t.Fatalf("stack object = %+v, want same event card's Adventure face", obj)
	}
}

func TestSelfDiesAdventureCastPermissionExpiresAfterNextTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, selfDiesAdventureDefinition())
	cardID := source.CardInstanceID
	g.Turn.TurnNumber = 4

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-dies cast-permission trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	g.Turn.TurnNumber = 5
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !engine.canCastSpellFaceFromZoneWithKicker(g, game.Player1, cardID, zone.Graveyard, game.FaceAlternate, nil, 0, nil, false) {
		t.Fatal("Adventure permission did not last through controller's next turn")
	}
	expireRuleEffects(g)
	if engine.canCastSpellFaceFromZoneWithKicker(g, game.Player1, cardID, zone.Graveyard, game.FaceAlternate, nil, 0, nil, false) {
		t.Fatal("Adventure permission remained after controller's next-turn cleanup")
	}
}

func TestSelfDiesAdventureCastPermissionEndsWhenCardLeavesGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, selfDiesAdventureDefinition())
	cardID := source.CardInstanceID

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-dies cast-permission trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) ||
		!moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) {
		t.Fatal("moving permitted card out of and back into graveyard failed")
	}

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if engine.canCastSpellFaceFromZoneWithKicker(g, game.Player1, cardID, zone.Graveyard, game.FaceAlternate, nil, 0, nil, false) {
		t.Fatal("cast permission followed a card through a zone change")
	}
}

func TestCounterTransferInterveningIfUsesLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Target Relic",
		Types: []types.Card{types.Artifact}},
	})
	addCounterTransferTriggerSource(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dying Relic",
		Types: []types.Card{types.Artifact}},
	})

	movePermanentToZone(g, source, zone.Graveyard)

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("counter transfer trigger was put on stack for artifact with no counters")
	}
}

func TestSelfDiesCounterAbsenceInterveningIfUsesLKI(t *testing.T) {
	tests := []struct {
		name        string
		counterKind counter.Kind
		add         []counter.Kind
		wantTrigger bool
	}{
		{
			name:        "absent",
			counterKind: counter.PlusOnePlusOne,
			wantTrigger: true,
		},
		{
			name:        "same kind present",
			counterKind: counter.PlusOnePlusOne,
			add:         []counter.Kind{counter.PlusOnePlusOne},
		},
		{
			name:        "different kind present",
			counterKind: counter.PlusOnePlusOne,
			add:         []counter.Kind{counter.Charge},
			wantTrigger: true,
		},
		{
			name:        "minus kind present",
			counterKind: counter.MinusOneMinusOne,
			add:         []counter.Kind{counter.MinusOneMinusOne},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addSelfDiesCounterAbsenceTrigger(g, test.counterKind)
			for _, kind := range test.add {
				source.Counters.Add(kind, 1)
			}

			movePermanentToZone(g, source, zone.Graveyard)
			if got := engine.putTriggeredAbilitiesOnStack(g); got != test.wantTrigger {
				t.Fatalf("putTriggeredAbilitiesOnStack = %v, want %v", got, test.wantTrigger)
			}
		})
	}
}

func TestSelfDiesCounterAbsenceInterveningIfFailsClosedAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addSelfDiesCounterAbsenceTrigger(g, counter.PlusOnePlusOne)

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("counter-absence trigger was not put on stack")
	}
	delete(g.LastKnownInformation, source.ObjectID)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want no draw when LKI is unavailable at resolution", got)
	}
}

func TestCounterTransferUpToOneTargetMayHaveNoTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCounterTransferTriggerSource(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dying Relic",
		Types: []types.Card{types.Artifact}},
	})
	source.Counters.Add(counter.PlusOnePlusOne, 1)

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("counter transfer trigger with no legal target was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.Targets) != 0 {
		t.Fatalf("trigger targets = %+v, want no targets", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
}

func TestCounterTransferUpToOneTargetCanBeDeclined(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	destination := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Target Relic",
		Types: []types.Card{types.Artifact}},
	})
	addCounterTransferTriggerSource(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dying Relic",
		Types: []types.Card{types.Artifact}},
	})
	source.Counters.Add(counter.PlusOnePlusOne, 1)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("counter transfer trigger was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.Targets) != 0 {
		t.Fatalf("trigger targets = %+v, want declined target choice", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := destination.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("destination +1/+1 counters = %d, want declined transfer", got)
	}
}

func TestSelfDeathTriggerUsesLeftBattlefieldSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	permanent := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	destroyPermanent(g, permanent.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-death trigger was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != permanent.ObjectID || obj.SourceCardID != permanent.CardInstanceID {
		t.Fatalf("top of stack = %+v, want self-death trigger source %+v", obj, permanent)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want self-death trigger to draw one card", got)
	}
}

func TestTokenTriggersUseTokenDefinition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	token := triggeredCreature(&game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	permanent, ok := createTokenPermanent(g, game.Player1, token)
	if !ok {
		t.Fatal("token was not created")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token ETB trigger was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != permanent.ObjectID || obj.SourceCardID != 0 || obj.SourceTokenDef != token {
		t.Fatalf("top of stack = %+v, want token trigger source %+v", obj, permanent)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want token ETB trigger to draw one card", got)
	}
}

func TestTokenSelfETBTriggerDoesNotMatchOtherToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Drawn"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Drawn"}})
	token := triggeredCreature(&game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	if _, ok := createTokenPermanent(g, game.Player1, token); !ok {
		t.Fatal("first token was not created")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first token ETB trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	second, ok := createTokenPermanent(g, game.Player1, token)
	if !ok {
		t.Fatal("second token was not created")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("second token ETB trigger was not put on stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 1 {
		t.Fatalf("stack size = %d, want only the second token's self trigger", len(objects))
	}
	if objects[0].SourceID != second.ObjectID {
		t.Fatalf("trigger source = %v, want second token %v", objects[0].SourceID, second.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want exactly two token ETB triggers resolved", got)
	}
}

func TestTriggerPatternExcludeSelfSkipsSourcePermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentEnteredBattlefield,
		Controller:            game.TriggerControllerYou,
		ExcludeSelf:           true,
		RequirePermanentTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		Player:      game.Player1,
		CardID:      source.CardInstanceID,
		PermanentID: source.ObjectID,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ExcludeSelf trigger fired for its own source permanent")
	}

	other := addCombatCreaturePermanent(g, game.Player1)
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		Player:      game.Player1,
		CardID:      other.CardInstanceID,
		PermanentID: other.ObjectID,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ExcludeSelf trigger did not fire for another matching permanent")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want trigger from source %v", obj, source.ObjectID)
	}
}

func TestTokenSelfDeathTriggerUsesLeftBattlefieldSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	token := triggeredCreature(&game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	permanent, ok := createTokenPermanent(g, game.Player1, token)
	if !ok {
		t.Fatal("token was not created")
	}
	destroyPermanent(g, permanent.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token self-death trigger was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != permanent.ObjectID || obj.SourceCardID != 0 || obj.SourceTokenDef != token {
		t.Fatalf("top of stack = %+v, want token self-death trigger source %+v", obj, permanent)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want token self-death trigger to draw one card", got)
	}
}

func TestTokenSelfDeathTriggerDoesNotMatchOtherToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	token := triggeredCreature(&game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	first, ok := createTokenPermanent(g, game.Player1, token)
	if !ok {
		t.Fatal("first token was not created")
	}
	if _, ok := createTokenPermanent(g, game.Player1, token); !ok {
		t.Fatal("second token was not created")
	}
	destroyPermanent(g, first.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token self-death trigger was not put on stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 1 {
		t.Fatalf("stack size = %d, want only the dying token's self trigger", len(objects))
	}
	if objects[0].SourceID != first.ObjectID {
		t.Fatalf("trigger source = %v, want dying token %v", objects[0].SourceID, first.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want exactly one token self-death trigger resolved", got)
	}
}

func TestDamageTriggerGoesOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:           game.EventDamageDealt,
		Player:          game.TriggerPlayerOpponent,
		DamageRecipient: game.DamageRecipientPlayer,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	dealPlayerDamage(g, 0, 0, game.Player2, game.Player2, 1, false)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want damage trigger to draw one card", got)
	}
}

func TestCombatDamageTriggerRequiresCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:               game.EventDamageDealt,
		Source:              game.TriggerSourceSelf,
		Subject:             game.TriggerSubjectDamageSource,
		DamageRecipient:     game.DamageRecipientPlayer,
		RequireCombatDamage: true,
	}
	event := game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  source.ObjectID,
		Controller:      game.Player1,
		Player:          game.Player2,
		DamageRecipient: game.DamageRecipientPlayer,
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("combat-damage trigger matched non-combat damage")
	}
	event.CombatDamage = true
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("combat-damage trigger did not match combat damage")
	}
}

func TestDamageSourceSubjectDoesNotMatchDamageRecipient(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	recipient := addCombatCreaturePermanent(g, game.Player1)
	source := addCombatCreaturePermanent(g, game.Player2)
	pattern := &game.TriggerPattern{
		Event:                game.EventDamageDealt,
		Source:               game.TriggerSourceSelf,
		Subject:              game.TriggerSubjectDamageSource,
		DamageRecipient:      game.DamageRecipientPermanent,
		DamageRecipientTypes: []types.Card{types.Creature},
		RequireCombatDamage:  true,
	}
	event := game.Event{
		Kind:            game.EventDamageDealt,
		SourceID:        source.CardInstanceID,
		SourceObjectID:  source.ObjectID,
		Controller:      game.Player2,
		CardID:          recipient.CardInstanceID,
		PermanentID:     recipient.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
		CombatDamage:    true,
	}
	if triggerMatchesEvent(g, recipient, pattern, event) {
		t.Fatal("damage-source trigger matched the damage recipient")
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("damage-source trigger did not match the damage source")
	}
	nonCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Attacked Battle",
		Types: []types.Card{types.Battle},
	}})
	event.PermanentID = nonCreature.ObjectID
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("damage-to-creature trigger matched a noncreature permanent recipient")
	}
}

func TestCombatDamageSourceTriggerUsesLKIAfterSourceDies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	attacker := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventDamageDealt,
		Source:               game.TriggerSourceSelf,
		Subject:              game.TriggerSubjectDamageSource,
		DamageRecipient:      game.DamageRecipientPermanent,
		DamageRecipientTypes: []types.Card{types.Creature},
		RequireCombatDamage:  true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		Blockers: []game.BlockDeclaration{{
			Blocker:  blocker.ObjectID,
			Blocking: attacker.ObjectID,
		}},
		BlockedAttackers: map[id.ID]bool{attacker.ObjectID: true},
		BlockerOrder:     map[id.ID][]id.ID{attacker.ObjectID: {blocker.ObjectID}},
	}

	combatEngine{}.resolveDamagePass(g, normalCombatDamage, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, attacker.ObjectID); ok {
		t.Fatal("attacker survived combat damage; test requires LKI trigger source")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger from dead source was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want dead source trigger to draw one card", got)
	}
}

func TestBecomesBlockedTriggerFiresOnceForMultipleBlockers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	attacker := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventAttackerBecameBlocked,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	firstBlocker := addCombatCreaturePermanent(g, game.Player2)
	secondBlocker := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
	}
	declare := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: firstBlocker.ObjectID, Blocking: attacker.ObjectID},
		{Blocker: secondBlocker.ObjectID, Blocking: attacker.ObjectID},
	}))

	if !engine.applyDeclareBlockers(g, game.Player2, declare) {
		t.Fatal("applyDeclareBlockers() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("becomes-blocked trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one becomes-blocked trigger", got)
	}
}

func TestDrawTriggerChoosesDeterministicLegalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}}}, []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"},
	})

	if _, ok := engine.drawCard(g, game.Player1); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 39 {
		t.Fatalf("player 2 life = %d, want deterministic opponent target to lose 1", got)
	}
}

func TestTriggerTargetChoiceCanBeMadeByAgent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}}}, []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"},
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player1); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("draw trigger was not put on stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("player 2 life = %d, want agent to choose another target", got)
	}
	if got := g.Players[game.Player3].Life; got != 39 {
		t.Fatalf("player 3 life = %d, want chosen target to lose 1", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceTarget || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want recorded target choice without fallback", log.Choices)
	}
}

func TestOptionalTriggeredAbilityChoiceHappensOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Triggering Drawn"}})
	addOptionalTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventCardDrawn,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player2); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("optional trigger was not put on stack")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want optional trigger on stack before may choice", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("player 1 hand size = %d, want optional trigger declined on resolution", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceMay || log.Choices[0].Selected[0] != 0 {
		t.Fatalf("choices = %+v, want no may choice recorded", log.Choices)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "declined" {
		t.Fatalf("resolves = %+v, want declined optional trigger", log.Resolves)
	}
}

func TestCastTriggerGoesOnStackAboveCastSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	spellID := addCardToHand(g, game.Player1, greenInstant())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cast trigger was not put on stack")
	}

	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility {
		t.Fatalf("top of stack = %+v, want cast trigger above cast spell", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want cast trigger to draw one card", got)
	}
}

func TestBeginningOfUpkeepTriggerResolvesBeforeDrawStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Upkeep Draw"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepUpkeep,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want upkeep trigger plus draw step draw", got)
	}
}

func TestBeginningOfEndStepTriggerResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "End Step Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepEnd,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want end-step trigger draw", got)
	}
}

func TestBeginningOfDrawStepTriggerResolvesAfterTurnDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Trigger Draw"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Turn Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepDraw,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want turn draw plus draw-step trigger", got)
	}
}

func TestBeginningOfCombatTriggerResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Combat Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepBeginningOfCombat,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runCombatPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want beginning-of-combat trigger draw", got)
	}
}

func TestBeginningOfStepTriggerRequiresExplicitStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Should Not Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want only turn draw without broad step trigger", got)
	}
}

func TestStateTriggerLatchesUntilConditionBecomesFalse(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.Type = game.TriggerState
	card.Def.TriggeredAbilities[0].Trigger.State = opt.Val(game.StateTriggerCondition{MatchControllerLifeLessOrEqual: true, ControllerLifeLessOrEqual: 10})
	g.Players[game.Player1].Life = 10

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger re-fired while condition remained true")
	}
	g.Players[game.Player1].Life = 11
	engine.putTriggeredAbilitiesOnStack(g)
	g.Players[game.Player1].Life = 10
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger did not re-arm after condition became false")
	}
}

func TestSpellCastTriggerFiltersCardTypesAndController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventSpellCast,
		Controller:       game.TriggerControllerOpponent,
		RequireCardTypes: []types.Card{types.Instant},
		ExcludeCardTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	spellID := addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, types.Forest)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent instant cast trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want spell-cast trigger draw", got)
	}
}

func TestBlockedAttackerSubjectMatchesAttachedPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	blocker := addCombatCreaturePermanent(g, game.Player2)
	equipment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment}},
	})
	if !attachPermanent(g, equipment, attacker) {
		t.Fatal("attachPermanent failed")
	}
	event := game.Event{
		Kind:              game.EventBlockerDeclared,
		Controller:        game.Player2,
		PermanentID:       blocker.ObjectID,
		BlockedAttackerID: attacker.ObjectID,
	}
	pattern := &game.TriggerPattern{
		Event:                 game.EventBlockerDeclared,
		Controller:            game.TriggerControllerYou,
		Source:                game.TriggerSourceAttachedPermanent,
		Subject:               game.TriggerSubjectBlockedAttacker,
		RequirePermanentTypes: []types.Card{types.Creature},
	}
	if !triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached equipment did not match blocked attacker subject")
	}
	pattern.Subject = game.TriggerSubjectDefault
	if triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached equipment matched blocker as default subject")
	}
	pattern.Subject = game.TriggerSubjectBlockedAttacker
	nonCreatureBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Land", Types: []types.Card{types.Land}}})
	event.PermanentID = nonCreatureBlocker.ObjectID
	if triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("creature type filter matched blocked attacker instead of blocker")
	}
}

func TestSpellTargetTriggerPredicates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	ownCreature := addCombatCreaturePermanent(g, game.Player1)
	opponentCreature := addCombatCreaturePermanent(g, game.Player2)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(source.ObjectID),
			game.PermanentTarget(opponentCreature.ObjectID),
		},
	}
	g.Stack.Push(obj)
	event := game.Event{Kind: game.EventSpellCast, StackObjectID: obj.ID, Controller: game.Player1}
	if !triggerMatchesEvent(g, source, &game.TriggerPattern{
		Event:              game.EventSpellCast,
		Controller:         game.TriggerControllerYou,
		SpellTargetsSource: true,
	}, event) {
		t.Fatal("spell-targets-source trigger did not match source target")
	}
	if !triggerMatchesEvent(g, source, &game.TriggerPattern{
		Event:            game.EventSpellCast,
		Controller:       game.TriggerControllerYou,
		SpellTargetAllow: game.TargetAllowPermanent,
		SpellTargetPattern: opt.Val(game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
			Controller:     game.ControllerNotYou,
		}),
	}, event) {
		t.Fatal("spell target predicate did not match opponent creature target")
	}
	obj.Targets = []game.Target{game.PermanentTarget(ownCreature.ObjectID)}
	if triggerMatchesEvent(g, source, &game.TriggerPattern{
		Event:            game.EventSpellCast,
		Controller:       game.TriggerControllerYou,
		SpellTargetAllow: game.TargetAllowPermanent,
		SpellTargetPattern: opt.Val(game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
			Controller:     game.ControllerNotYou,
		}),
	}, event) {
		t.Fatal("spell target predicate matched own creature target")
	}
}

func TestTriggeredAbilityMaxTriggersPerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.TriggeredAbilities[0].MaxTriggersPerTurn = 1

	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size after two same-turn events = %d, want 1", got)
	}
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger exceeded max triggers per turn")
	}
	engine.advanceToNextTurn(g)
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger did not reset next turn")
	}
}

func TestOneOrMoreTriggerCoalescesDetectionBatch(t *testing.T) {
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
	emitEvent(g, game.Event{Kind: game.EventPermanentDied, Controller: game.Player1, PermanentID: first.ObjectID, CardID: first.CardInstanceID})
	emitEvent(g, game.Event{Kind: game.EventPermanentDied, Controller: game.Player1, PermanentID: second.ObjectID, CardID: second.CardInstanceID})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one coalesced trigger", got)
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

func TestInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	t.Run("not put on stack when false at trigger time", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].Life = 40
		addTriggeredPermanentWithCondition(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}})
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		if _, ok := engine.drawCard(g, game.Player2); !ok {
			t.Fatal("drawCard() = false, want true")
		}
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("intervening-if false trigger was put on stack")
		}
	})
	t.Run("does not resolve when false at resolution time", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].Life = 41
		addTriggeredPermanentWithCondition(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}})
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Event Drawn"}})
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Trigger Drawn"}})
		if _, ok := engine.drawCard(g, game.Player2); !ok {
			t.Fatal("drawCard() = false, want true")
		}
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("intervening-if true trigger was not put on stack")
		}
		g.Players[game.Player1].Life = 40
		log := TurnLog{}
		engine.resolveTopOfStack(g, &log)
		if g.Players[game.Player1].Hand.Size() != 0 {
			t.Fatal("intervening-if false on resolution still applied effect")
		}
		if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
			t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
		}
	})
}

func TestKickedInterveningIfChecksEnterEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	trigger := game.TriggerCondition{InterveningIfEventPermanentWasKicked: true}
	if triggerInterveningIf(g, nil, game.Player1, &trigger, &game.Event{}) {
		t.Fatal("unkicked enter event satisfied kicked intervening-if")
	}
	if !triggerInterveningIf(g, nil, game.Player1, &trigger, &game.Event{KickerPaid: true}) {
		t.Fatal("kicked enter event did not satisfy kicked intervening-if")
	}
}

func TestWasCastInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	t.Run("not put on stack when permanent was not cast", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addSelfEnterInterveningTrigger(g, &game.TriggerCondition{
			InterveningIfEventPermanentWasCast: true,
		})
		emitEvent(g, game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			CardID:      source.CardInstanceID,
			PermanentID: source.ObjectID,
		})
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("non-cast permanent enter trigger was put on stack")
		}
	})

	t.Run("cast fact is rechecked on resolution", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addSelfEnterInterveningTrigger(g, &game.TriggerCondition{
			InterveningIfEventPermanentWasCast: true,
		})
		emitEvent(g, game.Event{
			Kind:         game.EventPermanentEnteredBattlefield,
			CardID:       source.CardInstanceID,
			PermanentID:  source.ObjectID,
			EnterWasCast: true,
		})
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("cast permanent enter trigger was not put on stack")
		}
		obj, ok := g.Stack.Peek()
		if !ok {
			t.Fatal("missing triggered ability")
		}
		obj.TriggerEvent.EnterWasCast = false
		log := TurnLog{}
		engine.resolveTopOfStack(g, &log)
		if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
			t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
		}
	})
}

func TestControlsPermanentInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addSelfEnterInterveningTrigger(g, &game.TriggerCondition{
		InterveningCondition: opt.Val(game.Condition{
			ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
			}),
		}),
	})
	enter := game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		CardID:      source.CardInstanceID,
		PermanentID: source.ObjectID,
	}
	emitEvent(g, enter)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controls-artifact enter trigger was put on stack without an artifact")
	}

	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	}})
	artifact.PhasedOut = true
	emitEvent(g, enter)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controls-artifact enter trigger counted a phased-out artifact")
	}

	artifact.PhasedOut = false
	emitEvent(g, enter)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controls-artifact enter trigger was not put on stack")
	}
	artifact.PhasedOut = true
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
		t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
	}
}

func TestInterveningIfUsesEffectiveControllerAtTriggerTime(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 10
	g.Players[game.Player2].Life = 41
	triggerSource := addTriggeredPermanentWithCondition(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}})
	newController := game.Player2
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: triggerSource.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(newController),
	})
	addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Event Drawn"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Trigger Drawn"}})

	if _, ok := engine.drawCard(g, game.Player3); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger controlled by Player2 should use Player2 life for intervening-if")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Controller != game.Player2 {
		t.Fatalf("trigger controller = %+v, want Player2", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Hand.Size() != 1 {
		t.Fatal("trigger did not resolve for effective controller")
	}
}

func TestTriggeredAbilitiesUseAPNAPStackOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	addTriggeredPermanent(g, game.Player2, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)

	if _, ok := engine.drawCard(g, game.Player3); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw triggers were not put on stack")
	}

	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack size = %d, want two triggers", len(objects))
	}
	if objects[0].Controller != game.Player1 || objects[1].Controller != game.Player2 {
		t.Fatalf("stack controllers bottom-to-top = %v, %v; want active player's trigger below next player's trigger", objects[0].Controller, objects[1].Controller)
	}
}

func TestTriggeredAbilitiesUseAgentOrderWithinController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	first := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	second := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1, 0}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player2); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("draw triggers were not put on stack")
	}

	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack size = %d, want two triggers", len(objects))
	}
	if objects[0].SourceID != second.ObjectID || objects[1].SourceID != first.ObjectID {
		t.Fatalf("stack sources bottom-to-top = %v, %v; want agent order %v, %v", objects[0].SourceID, objects[1].SourceID, second.ObjectID, first.ObjectID)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceOrder || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want recorded order choice without fallback", log.Choices)
	}
}

func addTriggeredPermanent(g *game.Game, controller game.PlayerID, pattern *game.TriggerPattern, instructions []game.Instruction, targets []game.TargetSpec) *game.Permanent {
	return addCombatPermanent(g, controller, triggeredCreature(pattern, instructions, targets))
}

func addOptionalTriggeredPermanent(g *game.Game, controller game.PlayerID, pattern *game.TriggerPattern, instructions []game.Instruction, targets []game.TargetSpec) *game.Permanent {
	card := triggeredCreature(pattern, instructions, targets)
	card.TriggeredAbilities[0].Optional = true
	return addCombatPermanent(g, controller, card)
}

func addTriggeredPermanentWithCondition(g *game.Game, controller game.PlayerID, pattern *game.TriggerPattern, lifeAtLeast int, instructions []game.Instruction) *game.Permanent {
	permanent := addTriggeredPermanent(g, controller, pattern, instructions, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningIfControllerLifeAtLeast = lifeAtLeast
	return permanent
}

func addSelfEnterInterveningTrigger(g *game.Game, condition *game.TriggerCondition) *game.Permanent {
	permanent := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, nil, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	condition.Type = game.TriggerWhen
	condition.Pattern = game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}
	card.Def.TriggeredAbilities[0].Trigger = *condition
	return permanent
}

func addSelfDiesCounterAbsenceTrigger(g *game.Game, kind counter.Kind) *game.Permanent {
	permanent := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningIfEventPermanentHadNoCounterKind = opt.Val(kind)
	return permanent
}

func selfDiesEventCardDefinition(primitive game.Primitive) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Returning Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
				Event:  game.EventPermanentDied,
				Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability(),
		}},
	}}
}

func selfDiesAdventureDefinition() *game.CardDef {
	def := selfDiesEventCardDefinition(game.GrantCastPermission{
		Card:     game.CardReference{Kind: game.CardReferenceEvent},
		FromZone: zone.Graveyard,
		Face:     game.FaceAlternate,
		Duration: game.DurationUntilEndOfYourNextTurn,
	})
	def.Layout = game.LayoutAdventure
	def.Alternate = opt.Val(game.CardFace{
		Name:         "Returning Adventure",
		Types:        []types.Card{types.Sorcery},
		Subtypes:     []types.Sub{types.Adventure},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}}.Ability()),
	})
	return def
}

func addCounterTransferTriggerSource(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Counter Transfer Source",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{Event: game.EventZoneChanged, Controller: game.TriggerControllerYou, RequirePermanentTypes: []types.Card{types.Artifact}, MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Graveyard}, InterveningIf: "it had counters on it", InterveningIfEventPermanentHadCounters: true},
				Content: game.Mode{
					Targets: []game.TargetSpec{
						{MinTargets: 0, MaxTargets: 1, Constraint: "artifact or creature you control"},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.MoveCounters{
								Object: game.TargetPermanentReference(0),
								Source: game.CounterSourceSpec{
									Kind: game.CounterSourceEventPermanent,
								},
							},
						},
					},
				}.Ability(),
			},
		}},
	})
}

func TestTriggerPatternRequireNonToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pattern := &game.TriggerPattern{
		Event:                 game.EventPermanentEnteredBattlefield,
		Controller:            game.TriggerControllerYou,
		RequirePermanentTypes: []types.Card{types.Creature},
		RequireNonToken:       true,
	}
	source := addTriggeredPermanent(g, game.Player1, pattern, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	token, ok := createTokenPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Token", Types: []types.Card{types.Creature}}})
	if !ok {
		t.Fatal("createTokenPermanent failed")
	}
	card := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Nontoken", Types: []types.Card{types.Creature}}})

	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: token.ObjectID,
		TokenName:   "Token",
		TokenDef:    token.TokenDef,
	}) {
		t.Fatal("non-token trigger matched token event")
	}
	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: card.ObjectID,
		CardID:      card.CardInstanceID,
	}) {
		t.Fatal("non-token trigger did not match nontoken event")
	}
}

func triggeredCreature(pattern *game.TriggerPattern, instructions []game.Instruction, targets []game.TargetSpec) *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{Name: "Triggered Creature",
		Types:     []types.Card{types.Creature},
		ManaCost:  greenCost(),
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: *pattern},
				Content: game.Mode{
					Targets:  targets,
					Sequence: instructions,
				}.Ability(),
			},
		}},
	}
}

type choiceOnlyAgent struct {
	choices [][]int
	next    int
}

func (*choiceOnlyAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	return action.Pass()
}

func (a *choiceOnlyAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if a.next >= len(a.choices) {
		return nil
	}
	choice := append([]int(nil), a.choices[a.next]...)
	a.next++
	return choice
}

func TestSpellCastTriggerMatchesColorSelection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			ColorsAny: []color.Color{color.Green},
		},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	greenSpell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Green},
	}}
	spellID := addCardToHand(g, game.Player1, greenSpell)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast green instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("green-spell cast trigger did not fire for green instant")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want green-spell trigger to draw one card", got)
	}
}

func TestSpellCastTriggerColorSelectionExcludesWrongColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{
			ColorsAny: []color.Color{color.Blue},
		},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	greenSpell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Green},
	}}
	spellID := addCardToHand(g, game.Player1, greenSpell)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast green instant failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("blue-spell trigger incorrectly fired for green instant")
	}
}

func TestSpellCastEventPopulatesColors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	greenSpell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Green},
	}}
	spellID := addCardToHand(g, game.Player1, greenSpell)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast green instant failed")
	}
	var castEvent *game.Event
	for i := range g.Events {
		if g.Events[i].Kind == game.EventSpellCast {
			castEvent = &g.Events[i]
			break
		}
	}
	if castEvent == nil {
		t.Fatal("no EventSpellCast found")
	}
	if len(castEvent.Colors) != 1 || castEvent.Colors[0] != color.Green {
		t.Fatalf("EventSpellCast.Colors = %v, want [Green]", castEvent.Colors)
	}
}
