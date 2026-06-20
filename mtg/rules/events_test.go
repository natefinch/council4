package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDrawCardEmitsDrawAndZoneChangeEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})

	drawn, ok := engine.drawCard(g, game.Player1)

	if !ok || drawn != cardID {
		t.Fatalf("drawCard() = %v, %v, want %v, true", drawn, ok, cardID)
	}
	assertEvent(t, g.Events, game.EventCardDrawn, func(event game.Event) bool {
		return event.Player == game.Player1 &&
			event.CardID == cardID &&
			event.FromZone == zone.Library &&
			event.ToZone == zone.Hand &&
			event.Amount == 1
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == cardID &&
			event.FromZone == zone.Library &&
			event.ToZone == zone.Hand
	})
	if zoneIndex := eventIndex(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == cardID &&
			event.FromZone == zone.Library &&
			event.ToZone == zone.Hand
	}); zoneIndex > eventIndex(g.Events, game.EventCardDrawn, func(event game.Event) bool {
		return event.CardID == cardID
	}) {
		t.Fatalf("draw zone-change event should precede draw-specific event: %+v", g.Events)
	}
}

func TestCastAndResolvePermanentSpellEmitsEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}

	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == spellID &&
			event.Controller == game.Player1 &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Stack
	})
	if zoneIndex := eventIndex(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == spellID &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Stack
	}); zoneIndex > eventIndex(g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == spellID
	}) {
		t.Fatalf("cast zone-change event should precede cast-specific event: %+v", g.Events)
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	assertEvent(t, g.Events, game.EventSpellResolved, func(event game.Event) bool {
		return event.CardID == spellID && event.Controller == game.Player1
	})
	zoneIndex := eventIndex(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == spellID &&
			event.FromZone == zone.Stack &&
			event.ToZone == zone.Battlefield
	})
	etbIndex := eventIndex(g.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.CardID == spellID &&
			event.Controller == game.Player1 &&
			event.FromZone == zone.Stack &&
			event.ToZone == zone.Battlefield &&
			event.PermanentID != 0
	})
	if zoneIndex == -1 || etbIndex == -1 || zoneIndex > etbIndex {
		t.Fatalf("zone change index = %d, ETB index = %d, want zone change before ETB in %+v", zoneIndex, etbIndex, g.Events)
	}
}

func TestPlayLandEmitsHandToBattlefieldZoneChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, basicLand())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.PlayLand(landID)) {
		t.Fatal("applyAction(PlayLand) = false, want true")
	}

	zoneIndex := eventIndex(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == landID &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Battlefield
	})
	etbIndex := eventIndex(g.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.CardID == landID &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Battlefield
	})
	if zoneIndex == -1 || etbIndex == -1 || zoneIndex > etbIndex {
		t.Fatalf("land play zone change index = %d, ETB index = %d, want hand-to-battlefield zone change before ETB in %+v", zoneIndex, etbIndex, g.Events)
	}
	assertNoEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == landID &&
			event.FromZone == zone.Stack
	})
}

func TestFaceDownEntryEventsPreserveFaceDownState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Hidden Creature",
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
		ManaCost:   opt.Val(cost.Mana{cost.O(3)}),
		OracleText: "Morph {3}",
	}})
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	if _, ok := createCardPermanentFaceDown(g, card, game.Player1, zone.Hand, game.FaceFront, game.FaceDownMorph, true); !ok {
		t.Fatal("createCardPermanentFaceDown failed")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == cardID && event.FaceDown
	})
	assertEvent(t, g.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.CardID == cardID && event.FaceDown
	})
}

func TestDestroyPermanentEmitsZoneChangeAndDeathEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatCreaturePermanent(g, game.Player2)
	permanent.FaceDown = true
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: permanent.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(game.Player1),
	})

	_, ok := destroyPermanent(g, permanent.ObjectID)

	if !ok {
		t.Fatal("destroyPermanent() ok = false, want true")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.PermanentID == permanent.ObjectID &&
			event.CardID == permanent.CardInstanceID &&
			event.Controller == game.Player1 &&
			event.Player == game.Player2 &&
			event.FaceDown &&
			event.FromZone == zone.Battlefield &&
			event.ToZone == zone.Graveyard
	})
	assertEvent(t, g.Events, game.EventPermanentDied, func(event game.Event) bool {
		return event.PermanentID == permanent.ObjectID &&
			event.CardID == permanent.CardInstanceID &&
			event.Controller == game.Player1 &&
			event.Player == game.Player2 &&
			event.FaceDown
	})
}

func TestDamageEffectEmitsDamageEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Fixed(3),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.Player == game.Player2 &&
			event.Controller == game.Player1 &&
			event.Amount == 3 &&
			event.DamageRecipient == game.DamageRecipientPlayer &&
			!event.CombatDamage
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == sourceID &&
			event.FromZone == zone.Stack &&
			event.ToZone == zone.Graveyard
	})
	assertEvent(t, g.Events, game.EventSpellResolved, func(event game.Event) bool {
		return event.CardID == sourceID
	})
}

func TestCounteredSpellEmitsStackToGraveyardZoneChangeButNoResolveEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player2)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Fixed(3),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	// Set target spec on the spell's content to require a creature target
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}
	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == sourceID &&
			event.FromZone == zone.Stack &&
			event.ToZone == zone.Graveyard
	})
	assertNoEvent(t, g.Events, game.EventSpellResolved, func(event game.Event) bool {
		return event.CardID == sourceID
	})
}

func TestMassDamageEffectEmitsDamageEventForEachPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature2 := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount: game.Fixed(2),
		Recipient: game.GroupDamageRecipient(
			game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.PermanentID == creature1.ObjectID &&
			event.Amount == 2 &&
			event.DamageRecipient == game.DamageRecipientPermanent
	})
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.PermanentID == creature2.ObjectID &&
			event.Amount == 2 &&
			event.DamageRecipient == game.DamageRecipientPermanent
	})
}

func TestActivatedAbilityDamageEventUsesPermanentSourceObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Pinger",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target player"}},
				Sequence: []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}}},
			}.Ability(),
		}}},
	})
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		AbilityIndex: 0,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PlayerTarget(game.Player2)},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.SourceID == source.CardInstanceID &&
			event.SourceObjectID == source.ObjectID &&
			event.Player == game.Player2 &&
			event.Amount == 1
	})
}

func TestCombatDamageToPermanentEmitsCombatDamageEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	engine.resolveCombatDamage(g, &TurnLog{})

	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.SourceObjectID == attacker.ObjectID &&
			event.PermanentID == blocker.ObjectID &&
			event.Amount == 3 &&
			event.DamageRecipient == game.DamageRecipientPermanent &&
			event.CombatDamage
	})
}

func TestTokenCreationEmitsZoneChangeBeforeETBEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token",
		Types: []types.Card{types.Creature}},
	}

	permanent, ok := createTokenPermanent(g, game.Player1, token)
	if !ok {
		t.Fatal("token was not created")
	}

	zoneIndex := eventIndex(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.PermanentID == permanent.ObjectID &&
			event.TokenName == token.Name &&
			event.FromZone == zone.None &&
			event.ToZone == zone.Battlefield
	})
	etbIndex := eventIndex(g.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.PermanentID == permanent.ObjectID &&
			event.TokenName == token.Name &&
			!event.EnterWasCast &&
			event.FromZone == zone.None &&
			event.ToZone == zone.Battlefield
	})
	if zoneIndex == -1 || etbIndex == -1 || zoneIndex > etbIndex {
		t.Fatalf("zone change index = %d, ETB index = %d, want zone change before ETB in %+v", zoneIndex, etbIndex, g.Events)
	}
	tokenCreatedIndex := eventIndex(g.Events, game.EventTokenCreated, func(event game.Event) bool {
		return event.PermanentID == permanent.ObjectID &&
			event.TokenName == token.Name &&
			event.Controller == game.Player1 &&
			event.Player == game.Player1
	})
	if tokenCreatedIndex <= etbIndex {
		t.Fatalf("token created index = %d, want after ETB index %d in %+v", tokenCreatedIndex, etbIndex, g.Events)
	}
}

func TestDiscardToMaximumHandSizeEmitsDiscardAndZoneChangeEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range maximumHandSize + 1 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}

	discardToMaximumHandSize(g, game.Player1)

	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.Player == game.Player1 &&
			event.CardID != 0 &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Graveyard &&
			event.Amount == 1
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID != 0 &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Graveyard
	})
	if zoneIndex := eventIndex(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.FromZone == zone.Hand &&
			event.ToZone == zone.Graveyard
	}); zoneIndex > eventIndex(g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.FromZone == zone.Hand &&
			event.ToZone == zone.Graveyard
	}) {
		t.Fatalf("discard zone-change event should precede discard-specific event: %+v", g.Events)
	}
}

func TestDeclareAttackersAndBlockersEmitEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanent(g, game.Player1)
	blocker := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	attackers, ok := action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}).DeclareAttackersPayload()
	if !ok || !engine.applyDeclareAttackers(g, game.Player1, attackers) {
		t.Fatal("applyDeclareAttackers() = false, want true")
	}
	g.Turn.Step = game.StepDeclareBlockers
	blockers, ok := action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
	}).DeclareBlockersPayload()
	if !ok || !engine.applyDeclareBlockers(g, game.Player2, blockers) {
		t.Fatal("applyDeclareBlockers() = false, want true")
	}

	assertEvent(t, g.Events, game.EventAttackerDeclared, func(event game.Event) bool {
		return event.PermanentID == attacker.ObjectID &&
			event.Controller == game.Player1 &&
			event.AttackTarget.Player == game.Player2
	})
	assertEvent(t, g.Events, game.EventBlockerDeclared, func(event game.Event) bool {
		return event.PermanentID == blocker.ObjectID &&
			event.Controller == game.Player2 &&
			event.BlockedAttackerID == attacker.ObjectID
	})
}

func TestCombatEventsCarryDeclarationAndDamageBatches(t *testing.T) {
	t.Run("attackers and blockers", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		firstAttacker := addCombatCreaturePermanent(g, game.Player1)
		secondAttacker := addCombatCreaturePermanent(g, game.Player1)
		firstBlocker := addCombatCreaturePermanent(g, game.Player2)
		secondBlocker := addCombatCreaturePermanent(g, game.Player2)
		g.Turn.Phase = game.PhaseCombat
		g.Turn.Step = game.StepDeclareAttackers
		g.Combat = &game.CombatState{}

		attackers, _ := action.DeclareAttackers([]game.AttackDeclaration{
			{Attacker: firstAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: secondAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		}).DeclareAttackersPayload()
		if !engine.applyDeclareAttackers(g, game.Player1, attackers) {
			t.Fatal("applyDeclareAttackers() = false, want true")
		}
		var attackEvents []game.Event
		for _, event := range g.Events {
			if event.Kind == game.EventAttackerDeclared {
				attackEvents = append(attackEvents, event)
			}
		}
		if len(attackEvents) != 2 {
			t.Fatalf("attack events = %+v, want two", attackEvents)
		}
		attackBatch := attackEvents[0].SimultaneousID
		if attackBatch == 0 || attackEvents[1].SimultaneousID != attackBatch {
			t.Fatalf("attack events do not share a nonzero batch: %+v", attackEvents)
		}
		for _, event := range attackEvents {
			if event.Player != game.Player2 {
				t.Fatalf("attack recipient player = %v, want Player2", event.Player)
			}
		}

		g.Turn.Step = game.StepDeclareBlockers
		eventStart := len(g.Events)
		blockers, _ := action.DeclareBlockers([]game.BlockDeclaration{
			{Blocker: firstBlocker.ObjectID, Blocking: firstAttacker.ObjectID},
			{Blocker: secondBlocker.ObjectID, Blocking: secondAttacker.ObjectID},
		}).DeclareBlockersPayload()
		if !engine.applyDeclareBlockers(g, game.Player2, blockers) {
			t.Fatal("applyDeclareBlockers() = false, want true")
		}
		blockBatch := g.Combat.BlockDeclarationBatchID
		if blockBatch == 0 {
			t.Fatal("block declaration batch ID = 0")
		}
		for _, event := range g.Events[eventStart:] {
			if (event.Kind == game.EventBlockerDeclared || event.Kind == game.EventAttackerBecameBlocked) &&
				(event.SimultaneousID != blockBatch || event.RelatedPermanentID == 0) {
				t.Fatalf("block event lacks shared batch or related combatant: %+v", event)
			}
		}
	})

	t.Run("combat damage", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
			{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		}}

		combatEngine{}.resolveDamagePass(g, normalCombatDamage, &TurnLog{})

		var damageEvents []game.Event
		for _, event := range g.Events {
			if event.Kind == game.EventDamageDealt && event.CombatDamage {
				damageEvents = append(damageEvents, event)
			}
		}
		if len(damageEvents) != 2 ||
			damageEvents[0].SimultaneousID == 0 ||
			damageEvents[0].SimultaneousID != damageEvents[1].SimultaneousID {
			t.Fatalf("combat damage events = %+v, want one nonzero shared batch", damageEvents)
		}
	})
}

func TestEventsArePartitionedByTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: 1})

	engine.advanceToNextTurn(g)
	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player2, Amount: 2})

	turnOne := g.EventsForTurn(1)
	if len(turnOne) != 1 || turnOne[0].Kind != game.EventLifeGained {
		t.Fatalf("turn one events = %+v, want life gained event", turnOne)
	}
	turnTwo := g.EventsForTurn(2)
	if len(turnTwo) != 1 || turnTwo[0].Kind != game.EventLifeLost {
		t.Fatalf("turn two events = %+v, want life lost event", turnTwo)
	}
	if got := g.EventsPreviousTurn(); len(got) != 1 || got[0].Kind != game.EventLifeGained {
		t.Fatalf("previous turn events = %+v, want turn one events", got)
	}
	if got := g.EventsThisTurn(); len(got) != 1 || got[0].Kind != game.EventLifeLost {
		t.Fatalf("this turn events = %+v, want turn two events", got)
	}
}

func TestLifeGainAndLossEmitEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	if gained := gainLife(g, game.Player1, 3); gained != 3 {
		t.Fatalf("gainLife() = %d, want 3", gained)
	}
	if lost := loseLife(g, game.Player2, 4); lost != 4 {
		t.Fatalf("loseLife() = %d, want 4", lost)
	}

	assertEvent(t, g.Events, game.EventLifeGained, func(event game.Event) bool {
		return event.Player == game.Player1 && event.Amount == 3
	})
	assertEvent(t, g.Events, game.EventLifeLost, func(event game.Event) bool {
		return event.Player == game.Player2 && event.Amount == 4
	})
}

func TestTapUntapAndTargetEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatCreaturePermanent(g, game.Player1)

	setPermanentTapped(g, permanent, true)
	setPermanentTapped(g, permanent, false)

	assertEvent(t, g.Events, game.EventPermanentTapped, func(event game.Event) bool {
		return event.PermanentID == permanent.ObjectID && event.Controller == game.Player1
	})
	assertEvent(t, g.Events, game.EventPermanentUntapped, func(event game.Event) bool {
		return event.PermanentID == permanent.ObjectID && event.Controller == game.Player1
	})

	spellID := addCardToHand(g, game.Player1, permanentTargetSpell("creature"))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(permanent.ObjectID)}, 0, nil)) {
		t.Fatal("targeted spell cast failed")
	}
	assertEvent(t, g.Events, game.EventObjectBecameTarget, func(event game.Event) bool {
		return event.SourceID == spellID &&
			event.Controller == game.Player1 &&
			event.PermanentID == permanent.ObjectID &&
			event.Target == game.PermanentTarget(permanent.ObjectID)
	})
}

func TestLifePaymentAndDamageEmitLifeLostEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Use a simple creature with an activated ability that costs life
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Pain Creature",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalPayLife, Amount: 2},
			},
			Content: game.Mode{
				Sequence: []game.Instruction{
					{Primitive: game.LoseLife{Player: game.ControllerReference(), Amount: game.Fixed(3)}},
				},
			}.Ability(),
		}}}},
	)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(creature.ObjectID, 0, nil, 0)) {
		t.Fatal("life-payment ability activation failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	assertEvent(t, g.Events, game.EventLifeLost, func(event game.Event) bool {
		return event.Player == game.Player1 && event.Amount == 2
	})
	assertEvent(t, g.Events, game.EventLifeLost, func(event game.Event) bool {
		return event.Player == game.Player1 && event.Amount == 3
	})
}

func TestPlayerEventsRecordOrdinalThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if gainLife(g, game.Player1, 1) != 1 || gainLife(g, game.Player1, 2) != 2 {
		t.Fatal("gain life failed")
	}
	if loseLife(g, game.Player1, 1) != 1 {
		t.Fatal("lose life failed")
	}
	assertEvent(t, g.Events, game.EventLifeGained, func(event game.Event) bool {
		return event.Player == game.Player1 && event.Amount == 1 && event.PlayerEventOrdinalThisTurn == 1
	})
	assertEvent(t, g.Events, game.EventLifeGained, func(event game.Event) bool {
		return event.Player == game.Player1 && event.Amount == 2 && event.PlayerEventOrdinalThisTurn == 2
	})
	assertEvent(t, g.Events, game.EventLifeLost, func(event game.Event) bool {
		return event.Player == game.Player1 && event.PlayerEventOrdinalThisTurn == 1
	})
}

func assertEvent(t *testing.T, events []game.Event, kind game.EventKind, matches func(game.Event) bool) {
	t.Helper()
	for _, event := range events {
		if event.Kind == kind && matches(event) {
			return
		}
	}
	t.Fatalf("missing event kind %v in events: %+v", kind, events)
}

func assertNoEvent(t *testing.T, events []game.Event, kind game.EventKind, matches func(game.Event) bool) {
	t.Helper()
	for _, event := range events {
		if event.Kind == kind && matches(event) {
			t.Fatalf("unexpected event kind %v in events: %+v", kind, events)
		}
	}
}

func eventIndex(events []game.Event, kind game.EventKind, matches func(game.Event) bool) int {
	for i, event := range events {
		if event.Kind == kind && matches(event) {
			return i
		}
	}
	return -1
}
