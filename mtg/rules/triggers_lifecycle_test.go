package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
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

// TestSelfETBTriggerSearchFetchesBasicLandTapped plays the enter-the-battlefield
// ramp tutor cluster (Wood Elves, the basic-land monuments) end to end: a
// creature whose ETB self-trigger carries a basic-land library search. When the
// creature enters and the trigger resolves, the basic land must leave the
// library and enter the battlefield tapped, proving the embedded (lowercase)
// search the parser now reconstructs behaves identically to a sentence-initial
// tutor inside a triggered-ability shell.
func TestSelfETBTriggerSearchFetchesBasicLandTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fetched := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Mountain",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Mountain},
	}})
	spellID := addCardToHand(g, game.Player1, triggeredCreature(&game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:   zone.Library,
			Destination:  zone.Battlefield,
			EntersTapped: true,
			Filter: game.Selection{
				RequiredTypes: []types.Card{types.Land},
				Supertypes:    []types.Super{types.Basic},
			},
		},
	}}}, nil))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast triggered creature failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ETB search trigger was not put on stack")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: selectAllAgent{}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Library.Contains(fetched) {
		t.Fatal("fetched basic land was not removed from the library")
	}
	permanent := permanentForCard(g, fetched)
	if permanent == nil {
		t.Fatal("fetched basic land did not enter the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("fetched basic land entered untapped, want tapped")
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

// TestExaltedGrantedByContinuousEffectTriggers proves that exalted granted to a
// permanent by a static keyword-grant continuous effect (the "<subject> has
// exalted" declaration) creates the exalted attacks-alone trigger, exactly like
// printed exalted.
func TestExaltedGrantedByContinuousEffectTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Exalted Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords: []game.Keyword{game.Exalted},
			}},
		}},
	}})
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
		t.Fatal("granted exalted trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, attacker); got != 3 {
		t.Fatalf("effective power = %d, want 3 after granted exalted", got)
	}
}

// TestExaltedStacksAcrossMultipleSources proves that each exalted instance
// triggers independently (CR 702.83b): with two exalted sources in play, a sole
// attacker gets +1/+1 from each, for +2/+2 total.
func TestExaltedStacksAcrossMultipleSources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 2 {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Exalted Source",
			Types:           []types.Card{types.Creature},
			Power:           opt.Val(game.PT{Value: 0}),
			Toughness:       opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{game.ExaltedStaticBody}},
		})
	}
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
		t.Fatal("exalted triggers were not put on stack")
	}
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStack(g, &TurnLog{})
	}

	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("effective power = %d, want 4 after two exalted triggers", got)
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
