package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-11-06 (verified copy retained in
// the unit's .cardwork/sources/cr-20261106.txt), read 510.1a-d, 510.2, 510.4,
// 702.19b, 611.2a and 400.7. 510.1a: each attacking and blocking creature
// assigns combat damage equal to its power, and a creature that would assign 0
// or less doesn't assign combat damage at all. 510.4: a first/double strike
// creature assigns in the first-strike step, a double strike creature again in
// the second. 611.2a: a resolved continuous effect lasts as long as stated.
// 400.7: a permanent that changes zones becomes a new object. Effect shapes
// (pinned corpus Oracle text): Keeper of Tresserhorn ("Whenever this creature
// attacks and isn't blocked, it assigns no combat damage this turn and
// defending player loses 2 life."), Laccolith Titan ("Whenever this creature
// becomes blocked, you may have it deal damage equal to its power to target
// creature. If you do, this creature assigns no combat damage this turn."),
// Carrion Rats ("Whenever this creature attacks or blocks, any player may
// exile a card from their graveyard. If a player does, this creature assigns
// no combat damage this turn."), Master of Cruelties (first strike, deathtouch;
// "... that player's life total becomes 1. This creature assigns no combat
// damage this combat."), Laccolith Rig (an Aura: "... the first creature
// assigns no combat damage this turn."). Rulings (Scryfall, retained in
// .cardwork/sources/*-rulings.json): Master of Cruelties 2013-04-15 —
// assigning no combat damage isn't the same as preventing that damage, so
// effects that make damage unpreventable have no effect; another attacking
// creature assigns combat damage normally. Laccolith Rig 2006-10-15 — moving
// the enchantment after the ability triggers doesn't change which creature is
// affected.

func noCombatDamageRule(source bool) game.RuleEffect {
	return game.RuleEffect{Kind: game.RuleEffectAssignNoCombatDamage, AffectedSource: source}
}

// selfNoCombatDamageTrigger is the "it assigns no combat damage <duration>"
// half of a self-referential combat trigger, followed by any rider.
func selfNoCombatDamageTrigger(event game.EventKind, duration game.EffectDuration, rider ...game.Instruction) game.TriggeredAbility {
	sequence := append([]game.Instruction{{Primitive: game.ApplyRule{
		RuleEffects: []game.RuleEffect{noCombatDamageRule(true)},
		Duration:    duration,
	}}}, rider...)
	return game.TriggeredAbility{
		Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{Event: event, Source: game.TriggerSourceSelf}},
		Content: game.Mode{Sequence: sequence}.Ability(),
	}
}

func noCombatDamageCreature(g *game.Game, controller game.PlayerID, power, toughness int, triggers []game.TriggeredAbility, keywords ...game.Keyword) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:               "No Combat Damage Creature",
		Types:              []types.Card{types.Creature},
		Power:              opt.Val(game.PT{Value: power}),
		Toughness:          opt.Val(game.PT{Value: toughness}),
		StaticAbilities:    []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(keywords...)}},
		TriggeredAbilities: triggers,
	}})
}

// applyNoCombatDamage resolves an ApplyRule from a separate source permanent
// (Laccolith Rig's shape) naming creature as its object.
func applyNoCombatDamage(g *game.Game, source, creature *game.Permanent, duration game.EffectDuration) {
	resolveInstruction(NewEngine(nil), g, &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   source.Controller,
		Targets:      []game.Target{game.PermanentTarget(creature.ObjectID)},
	}, game.ApplyRule{
		Object:      opt.Val(game.TargetPermanentReference(0)),
		RuleEffects: []game.RuleEffect{noCombatDamageRule(false)},
		Duration:    duration,
	}, nil)
}

// TestKeeperOfTresserhornShapeAssignsNoCombatDamage plays Keeper's trigger
// through a natural combat: the unblocked attacker's own trigger stops its
// combat damage, the life loss still happens, a second attacker deals its
// damage, and neither lifelink nor a "damage can't be prevented" effect
// produces damage that was never assigned.
func TestKeeperOfTresserhornShapeAssignsNoCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	keeper := noCombatDamageCreature(g, game.Player1, 6, 6, []game.TriggeredAbility{
		selfNoCombatDamageTrigger(game.EventAttackerBecameUnblocked, game.DurationThisTurn,
			game.Instruction{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.DefendingPlayerReference()}}),
	}, game.Lifelink)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	resolveInstruction(engine, g, &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1},
		game.ApplyRule{RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectDamageCantBePrevented}}, Duration: game.DurationThisTurn}, nil)
	log := TurnLog{}

	engine.runCombatPhase(g, [game.NumPlayers]PlayerAgent{game.Player1: widestAttackAgent{}, game.Player2: firstLegalAgent{}, game.Player3: firstLegalAgent{}, game.Player4: firstLegalAgent{}}, &log)

	if !sawAttackWith(log.Actions, keeper.ObjectID) || !sawAttackWith(log.Actions, other.ObjectID) {
		t.Fatal("both creatures did not attack")
	}
	if got := g.Players[game.Player2].Life; got != 35 {
		t.Fatalf("defending life = %d, want 35 (40 - 2 life loss - 3 from the other attacker)", got)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("attacking player life = %d, want 40: lifelink gained from unassigned damage", got)
	}
	for _, damage := range log.CombatDamage {
		if damage.SourceID == keeper.ObjectID || damage.Attacker == keeper.ObjectID {
			t.Fatalf("Keeper assigned combat damage %+v", damage)
		}
	}
	if len(log.CombatDamage) != 1 {
		t.Fatalf("combat damage logs = %+v, want only the other attacker's", log.CombatDamage)
	}
}

// TestLaccolithShapeBlockedAttackerAssignsNoCombatDamage: a blocked attacker
// whose "becomes blocked" trigger applies the effect assigns nothing to its
// blocker or, with trample, to the player, while its blocker still deals its
// damage (CR 510.1c-d).
func TestLaccolithShapeBlockedAttackerAssignsNoCombatDamage(t *testing.T) {
	for _, keywords := range [][]game.Keyword{nil, {game.Trample}} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		titan := noCombatDamageCreature(g, game.Player1, 6, 6, []game.TriggeredAbility{
			selfNoCombatDamageTrigger(game.EventAttackerBecameBlocked, game.DurationThisTurn),
		}, keywords...)
		blocker := noCombatDamageCreature(g, game.Player2, 2, 3, nil)
		log := TurnLog{}

		engine.runCombatPhase(g, allFirstLegalAgents(), &log)

		if blocker.MarkedDamage != 0 {
			t.Fatalf("keywords %v: blocker marked damage = %d, want 0", keywords, blocker.MarkedDamage)
		}
		if _, ok := permanentByObjectID(g, blocker.ObjectID); !ok {
			t.Fatalf("keywords %v: blocker died to unassigned damage", keywords)
		}
		if titan.MarkedDamage != 2 {
			t.Fatalf("keywords %v: titan marked damage = %d, want its blocker's 2", keywords, titan.MarkedDamage)
		}
		if got := g.Players[game.Player2].Life; got != 40 {
			t.Fatalf("keywords %v: defending life = %d, want 40", keywords, got)
		}
	}
}

// TestCarrionRatsShapeBlockerAssignsNoCombatDamage: a blocker with the effect
// assigns nothing to the attacker it blocks, which still deals its damage.
func TestCarrionRatsShapeBlockerAssignsNoCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := noCombatDamageCreature(g, game.Player1, 2, 2, nil)
	rats := noCombatDamageCreature(g, game.Player2, 2, 1, []game.TriggeredAbility{
		selfNoCombatDamageTrigger(game.EventBlockerDeclared, game.DurationThisTurn),
	})
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if attacker.MarkedDamage != 0 {
		t.Fatalf("attacker marked damage = %d, want 0", attacker.MarkedDamage)
	}
	if _, ok := permanentByObjectID(g, attacker.ObjectID); !ok {
		t.Fatal("attacker died to the rats' unassigned damage")
	}
	if _, ok := permanentByObjectID(g, rats.ObjectID); ok {
		t.Fatal("rats survived the attacker's combat damage")
	}
}

// TestMasterOfCrueltiesShapeAssignsNoCombatDamageThisCombat: "this combat"
// covers both combat damage steps of a double strike creature (CR 510.4) and
// lapses at the end of combat.
func TestMasterOfCrueltiesShapeAssignsNoCombatDamageThisCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	noCombatDamageCreature(g, game.Player1, 3, 4, []game.TriggeredAbility{
		selfNoCombatDamageTrigger(game.EventAttackerBecameUnblocked, game.DurationUntilEndOfCombat),
	}, game.DoubleStrike, game.Deathtouch)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if got := g.Players[game.Player2].Life; got != 40 || len(log.CombatDamage) != 0 {
		t.Fatalf("defending life = %d, combat damage %+v; want 40 and none in either step", got, log.CombatDamage)
	}
	for i := range g.RuleEffects {
		if g.RuleEffects[i].Kind == game.RuleEffectAssignNoCombatDamage {
			t.Fatal("this-combat no-damage effect outlived the combat")
		}
	}
}

// TestNoCombatDamageFollowsExactObject: the effect from another permanent
// (Laccolith Rig) keeps applying after that source leaves, and applies only
// to the named object, not to the card's new object after a zone change
// (CR 400.7).
func TestNoCombatDamageFollowsExactObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	rig := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Rig", Types: []types.Card{types.Enchantment}}})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	applyNoCombatDamage(g, rig, creature, game.DurationThisTurn)
	removePermanentFromBattlefield(g, rig.ObjectID)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: creature.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("defending life = %d, want 40 after the source left", got)
	}

	blink := immediateBlinkInstructions(game.PutOnBattlefield{})
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
		Targets: []game.Target{game.PermanentTarget(creature.ObjectID)}}
	for i := range blink {
		engine.resolveInstructionWithChoices(g, obj, &blink[i], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
	returned := permanentForCard(g, creature.CardInstanceID)
	if returned == nil || returned.ObjectID == creature.ObjectID {
		t.Fatal("creature did not return as a new object")
	}
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: returned.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != 36 {
		t.Fatalf("defending life = %d, want 36 from the new object", got)
	}
}

// widestAttackAgent declares the legal attack with the most attackers.
type widestAttackAgent struct{ firstLegalAgent }

func (a widestAttackAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	best, most := -1, 0
	for i := range legal {
		if payload, ok := legal[i].DeclareAttackersPayload(); ok && len(payload.Attackers) > most {
			best, most = i, len(payload.Attackers)
		}
	}
	if best < 0 {
		return a.firstLegalAgent.ChooseAction(obs, legal)
	}
	return legal[best]
}
