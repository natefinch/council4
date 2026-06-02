package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestShieldCounterPreventsDamageBeforeMutationAndEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sourceID := addColoredSourceCard(g, game.Player1, mana.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 3, false)

	if dealt != 0 {
		t.Fatalf("dealt damage = %d, want 0", dealt)
	}
	if target.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", target.MarkedDamage)
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.GameEvent) bool {
		return event.SourceID == sourceID &&
			event.PermanentID == target.ObjectID &&
			event.Amount == 3 &&
			event.DamageRecipient == game.DamageRecipientPermanent
	})
	assertNoEvent(t, g.Events, game.EventDamageDealt, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestShieldCounterReplacesDestroyBeforeZoneChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)

	removed, ok := destroyPermanent(g, target.ObjectID)

	if ok || removed != nil {
		t.Fatalf("destroyPermanent() = %+v, %v, want nil, false for replaced destroy", removed, ok)
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("shield-replaced permanent left the battlefield")
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	if g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("shield-replaced permanent moved to graveyard")
	}
	assertEvent(t, g.Events, game.EventDestroyReplaced, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID &&
			event.FromZone == game.ZoneBattlefield &&
			event.ToZone == game.ZoneGraveyard
	})
	assertNoEvent(t, g.Events, game.EventPermanentDied, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestPreventedCombatDamageDoesNotGrantLifelinkOrMarkDeathtouch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Deathtouch, game.Lifelink)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker.Counters.Add(counter.Shield, 1)
	g.Players[game.Player1].Life = 40
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("lifelink controller life = %d, want 40", g.Players[game.Player1].Life)
	}
	if blocker.MarkedDamage != 0 || blocker.MarkedDeathtouchDamage {
		t.Fatalf("blocker damage = %d deathtouch = %v, want no marked damage", blocker.MarkedDamage, blocker.MarkedDeathtouchDamage)
	}
}

func TestProtectionFromColorPreventsDamageAndTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, mana.Red)
	protected := addProtectionFromColorPermanent(g, game.Player2, mana.Red)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, protected, 2, false)

	if dealt != 0 {
		t.Fatalf("dealt damage = %d, want 0", dealt)
	}
	if protected.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", protected.MarkedDamage)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.GameEvent) bool {
		return event.SourceID == sourceID &&
			event.PermanentID == protected.ObjectID &&
			event.Amount == 2
	})

	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:   "Red Strike",
		Types:  []types.Card{types.Instant},
		Colors: []mana.Color{mana.Red},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
				},
				Effects: []game.Effect{{Type: game.EffectDamage, Amount: 1, TargetIndex: 0}},
			},
		},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if engine.canCastSpell(g, game.Player1, spellID, []game.Target{game.PermanentTarget(protected.ObjectID)}, 0, nil) {
		t.Fatal("red spell could target a permanent with protection from red")
	}
}

func TestHexproofPreventsOpponentTargetsButAllowsControllerTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hexproof := addHexproofPermanent(g, game.Player2)
	opponentSpell := addCardToHand(g, game.Player1, targetCreatureInstant())
	controllerSpell := addCardToHand(g, game.Player2, targetCreatureInstant())

	g.Turn.PriorityPlayer = game.Player1
	if engine.canCastSpell(g, game.Player1, opponentSpell, []game.Target{game.PermanentTarget(hexproof.ObjectID)}, 0, nil) {
		t.Fatal("opponent spell could target hexproof permanent")
	}

	g.Turn.PriorityPlayer = game.Player2
	if !engine.canCastSpell(g, game.Player2, controllerSpell, []game.Target{game.PermanentTarget(hexproof.ObjectID)}, 0, nil) {
		t.Fatal("controller spell could not target own hexproof permanent")
	}
}

func TestLegalActionsOmitOpponentHexproofTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hexproof := addHexproofPermanent(g, game.Player2)
	targetable := addCombatCreaturePermanent(g, game.Player3)
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)

	if actionsContain(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(hexproof.ObjectID)}, 0, nil)) {
		t.Fatalf("legal actions include hexproof target: %+v", legal)
	}
	if !actionsContain(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(targetable.ObjectID)}, 0, nil)) {
		t.Fatalf("legal actions omit non-hexproof target: %+v", legal)
	}
}

func TestHexproofCounterPreventsOpponentTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player2)
	target.Counters.Add(counter.Hexproof, 1)
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	if engine.canCastSpell(g, game.Player1, spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil) {
		t.Fatal("opponent spell could target permanent with hexproof counter")
	}
}

func TestPreventionShieldPreventsTrackedAmountAndExpires(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, mana.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}

	engine.resolveEffect(g, obj, &game.Effect{Type: game.EffectPrevent, Amount: 2, TargetIndex: 0}, nil)
	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 5, false)

	if dealt != 3 {
		t.Fatalf("dealt damage = %d, want 3 after prevention shield", dealt)
	}
	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want consumed", g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID && event.Amount == 2
	})

	engine.resolveEffect(g, obj, &game.Effect{Type: game.EffectPrevent, Amount: 1, TargetIndex: 0}, nil)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields after cleanup = %+v, want expired", g.PreventionShields)
	}
}

func TestMultiplePreventionShieldsRecordDeterministicReplacementOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, mana.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	engine.resolveEffect(g, obj, &game.Effect{Type: game.EffectPrevent, Amount: 1, TargetIndex: 0}, nil)
	engine.resolveEffect(g, obj, &game.Effect{Type: game.EffectPrevent, Amount: 1, TargetIndex: 0}, nil)

	dealPermanentDamage(g, sourceID, 0, game.Player1, target, 3, false)

	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one deterministic prevention order", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player2 || !decision.UsedFallback || len(decision.Selected) != 2 || decision.Selected[0] != 0 || decision.Selected[1] != 1 {
		t.Fatalf("replacement decision = %+v, want Player2 fallback order [0 1]", decision)
	}
}

func TestRegenerationReplacesDestroyAndRemovesFromCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker.MarkedDamage = 2
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: []id.ID{blocker.ObjectID},
		},
	}

	engine.resolveEffect(g, &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(blocker.ObjectID)},
	}, &game.Effect{Type: game.EffectRegenerate, TargetIndex: 0}, nil)
	removed, ok := destroyPermanent(g, blocker.ObjectID)

	if ok || removed != nil {
		t.Fatalf("destroyPermanent() = %+v, %v, want regenerated replacement", removed, ok)
	}
	if _, ok := permanentByObjectID(g, blocker.ObjectID); !ok {
		t.Fatal("regenerated blocker left battlefield")
	}
	if !blocker.Tapped || blocker.MarkedDamage != 0 || blocker.RegenerationShields != 0 {
		t.Fatalf("regenerated blocker tapped=%v damage=%d shields=%d, want tapped, no damage, no shields", blocker.Tapped, blocker.MarkedDamage, blocker.RegenerationShields)
	}
	if len(g.Combat.Blockers) != 0 || len(g.Combat.BlockerOrder[attacker.ObjectID]) != 0 {
		t.Fatalf("combat after regeneration blockers=%+v order=%+v, want blocker removed", g.Combat.Blockers, g.Combat.BlockerOrder)
	}
	assertEvent(t, g.Events, game.EventDestroyReplaced, func(event game.GameEvent) bool {
		return event.PermanentID == blocker.ObjectID
	})
}

func TestRegenerationShieldExpiresDuringCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.RegenerationShields = 1

	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if creature.RegenerationShields != 0 {
		t.Fatalf("regeneration shields = %d, want cleanup expiry", creature.RegenerationShields)
	}
}

func TestShieldAndRegenerationReplacementOrderIsRecorded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.Shield, 1)
	creature.RegenerationShields = 1

	destroyPermanent(g, creature.ObjectID)

	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one shield/regeneration order", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player1 || !decision.UsedFallback || len(decision.Selected) != 2 {
		t.Fatalf("replacement decision = %+v, want Player1 fallback order", decision)
	}
	if creature.Counters.Get(counter.Shield) != 0 || creature.RegenerationShields != 1 {
		t.Fatalf("shield counters=%d regeneration=%d, want shield used before regeneration", creature.Counters.Get(counter.Shield), creature.RegenerationShields)
	}
}

func TestRegenerationReplacesLethalDamageSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.MarkedDamage = 2
	creature.RegenerationShields = 1

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 0 {
		t.Fatalf("deaths = %+v, want regeneration to replace lethal-damage destruction", deaths)
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok || !creature.Tapped || creature.MarkedDamage != 0 {
		t.Fatalf("creature after regeneration = %+v, want tapped on battlefield with no damage", creature)
	}
}

func TestPermanentEntersTappedAndWithCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{
		Name:         "Tapped Walker",
		Types:        []types.Card{types.Creature},
		Power:        optPT(game.PT{Value: 1}),
		Toughness:    optPT(game.PT{Value: 1}),
		EntersTapped: true,
		EntersWithCounters: []game.CounterPlacement{
			{Kind: counter.PlusOnePlusOne, Amount: 2},
		},
	}

	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanent(g, card, game.Player1, game.ZoneHand)

	if !ok || !permanent.Tapped {
		t.Fatalf("permanent = %+v, want enters tapped", permanent)
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}
}

func TestEntersTappedUnlessPaidPaysLifeByDefault(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, [game.NumPlayers]PlayerAgent{}, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if permanent.Tapped {
		t.Fatalf("permanent = %+v, want untapped after paying life", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 37 {
		t.Fatalf("life = %d, want 37", got)
	}
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want one ETB payment choice", log.Choices)
	}
	choice := log.Choices[0]
	if choice.Request.Kind != game.ChoiceMay || choice.Request.Prompt != "Pay 3 life?" || len(choice.Selected) != 1 || choice.Selected[0] != 1 || !choice.UsedFallback {
		t.Fatalf("choice = %+v, want fallback yes for ETB payment", choice)
	}
}

func TestEntersTappedUnlessPaidDeclinedEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, agents, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped after declining payment", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("life = %d, want 40", got)
	}
	if len(log.Choices) != 1 || len(log.Choices[0].Selected) != 1 || log.Choices[0].Selected[0] != 0 || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want explicit no", log.Choices)
	}
}

func TestEntersTappedUnlessPaidCannotPayEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 2
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, payLifeETBModalLand())
	engine := NewEngine(nil)
	log := &TurnLog{}

	if !engine.applyPlayLandFaceWithChoices(g, game.Player1, cardID, game.FaceBack, [game.NumPlayers]PlayerAgent{}, log) {
		t.Fatal("applyPlayLandFaceWithChoices() = false")
	}

	permanent := g.Battlefield[len(g.Battlefield)-1]
	if !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped when payment is not payable", permanent)
	}
	if got := g.Players[game.Player1].Life; got != 2 {
		t.Fatalf("life = %d, want 2", got)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no prompt for unpayable ETB payment", log.Choices)
	}
}

func TestGenericReplacementChangesZoneDestination(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	engine.resolveEffect(g, &game.StackObject{Controller: game.Player1}, &game.Effect{
		Type: game.EffectReplace,
		Replacement: optReplacement(&game.ReplacementEffect{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      game.ZoneBattlefield,
			MatchToZone:   true,
			ToZone:        game.ZoneGraveyard,
			ReplaceToZone: game.ZoneExile,
		}),
	}, nil)

	if !movePermanentToZone(g, target, game.ZoneGraveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}

	if g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("replacement did not redirect away from graveyard")
	}
	if !g.Players[game.Player1].Exile.Contains(target.CardInstanceID) {
		t.Fatal("replacement did not move card to exile")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID && event.ToZone == game.ZoneExile
	})
}

func payLifeETBModalLand() *game.CardDef {
	return &game.CardDef{
		Name:   "Front Spell // Pay Life Land",
		Layout: game.LayoutModalDFC,
		Types:  []types.Card{types.Sorcery},
		Back: opt.Val(game.CardFace{
			Name:  "Pay Life Land",
			Types: []types.Card{types.Land},
			EntersTappedUnlessPaid: opt.Val(game.ResolutionPayment{
				Prompt: "Pay 3 life?",
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostPayLife, Amount: 3, Text: "Pay 3 life"},
				},
			}),
		}),
	}
}

func TestGenericETBReplacementAppliesTappedAndCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	engine.resolveEffect(g, &game.StackObject{Controller: game.Player1}, &game.Effect{
		Type: game.EffectReplace,
		Replacement: optReplacement(&game.ReplacementEffect{
			Description:  "enter modified",
			MatchEvent:   game.EventPermanentEnteredBattlefield,
			MatchToZone:  true,
			ToZone:       game.ZoneBattlefield,
			EntersTapped: true,
			EntersWithCounters: []game.CounterPlacement{
				{Kind: counter.PlusOnePlusOne, Amount: 1},
			},
		}),
	}, nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:      "Entering Creature",
		Types:     []types.Card{types.Creature},
		Power:     optPT(game.PT{Value: 1}),
		Toughness: optPT(game.PT{Value: 1}),
	})
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)

	permanent, ok := createCardPermanent(g, card, game.Player1, game.ZoneHand)

	if !ok || !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped by replacement", permanent)
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
}

func TestMultipleGenericReplacementsRecordOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	for _, replacement := range []game.ReplacementEffect{
		{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      game.ZoneBattlefield,
			MatchToZone:   true,
			ToZone:        game.ZoneGraveyard,
			ReplaceToZone: game.ZoneExile,
		},
		{
			Description:   "hand instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      game.ZoneBattlefield,
			ReplaceToZone: game.ZoneHand,
		},
	} {
		engine.resolveEffect(g, &game.StackObject{Controller: game.Player1}, &game.Effect{
			Type:        game.EffectReplace,
			Replacement: optReplacement(&replacement),
		}, nil)
	}

	if !movePermanentToZone(g, target, game.ZoneGraveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}

	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %+v, want one order decision", g.ReplacementDecisions)
	}
	decision := g.ReplacementDecisions[0]
	if decision.Player != game.Player1 || len(decision.Selected) != 2 || decision.Selected[0] != 0 || decision.Selected[1] != 1 {
		t.Fatalf("replacement decision = %+v, want deterministic Player1 order", decision)
	}
	if !g.Players[game.Player1].Hand.Contains(target.CardInstanceID) {
		t.Fatal("second replacement in fallback order should move card to hand")
	}
}

func TestPermanentSourceReplacementStopsAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Replacement Source",
		Types: []types.Card{types.Enchantment},
	})
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	engine.resolveEffect(g, &game.StackObject{
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}, &game.Effect{
		Type: game.EffectReplace,
		Replacement: optReplacement(&game.ReplacementEffect{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      game.ZoneBattlefield,
			MatchToZone:   true,
			ToZone:        game.ZoneGraveyard,
			ReplaceToZone: game.ZoneExile,
		}),
	}, nil)

	if !movePermanentToZone(g, source, game.ZoneGraveyard) {
		t.Fatal("source should leave battlefield")
	}
	if !movePermanentToZone(g, target, game.ZoneGraveyard) {
		t.Fatal("target should move to graveyard")
	}

	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("replacement from departed source should not apply")
	}
}

func TestSkipStepEffectSkipsNextDrawStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Would Draw"})
	engine.resolveEffect(g, &game.StackObject{Controller: game.Player1}, &game.Effect{
		Type:        game.EffectSkipStep,
		TargetIndex: game.TargetIndexController,
		Step:        game.StepDraw,
	}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want skipped draw step", got)
	}
	if g.Players[game.Player1].Library.Size() != 1 {
		t.Fatalf("library size = %d, want card not drawn", g.Players[game.Player1].Library.Size())
	}
}

func addColoredSourceCard(g *game.Game, owner game.PlayerID, color mana.Color) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{
			Name:   "Colored Source",
			Types:  []types.Card{types.Instant},
			Colors: []mana.Color{color},
		},
		Owner: owner,
	}
	return cardID
}

func addProtectionFromColorPermanent(g *game.Game, controller game.PlayerID, color mana.Color) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{
		Name:      "Protected Creature",
		Types:     []types.Card{types.Creature},
		Power:     optPT(pt),
		Toughness: optPT(pt),
		Abilities: []game.AbilityDef{
			{
				Kind:                 game.StaticAbility,
				Keywords:             []game.Keyword{game.Protection},
				ProtectionFromColors: []mana.Color{color},
			},
		},
	})
}

func addHexproofPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{
		Name:      "Hexproof Creature",
		Types:     []types.Card{types.Creature},
		Power:     optPT(pt),
		Toughness: optPT(pt),
		Abilities: []game.AbilityDef{{
			Kind:     game.StaticAbility,
			Keywords: []game.Keyword{game.Hexproof},
		}},
	})
}

func targetCreatureInstant() *game.CardDef {
	return &game.CardDef{
		Name:  "Target Creature Instant",
		Types: []types.Card{types.Instant},
		Abilities: []game.AbilityDef{{
			Kind:    game.SpellAbility,
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
		}},
	}
}
