package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-11-06 (verified copy retained in
// the unit's .cardwork/cr-20261106.txt), read 509.1c, 508.1d, 400.7 and
// 611.2c. 509.1c: the defending player's declaration is illegal if it obeys
// fewer requirements ("effects that say a creature must block") than the
// maximum possible without disobeying a restriction; a creature isn't required
// to pay a cost to block; a "blocks if able during a certain turn" requirement
// applies in each declare blockers step of that turn. 508.1d is the attack-side
// counterpart. 400.7: a creature that changes zones is a new object, so a
// requirement naming the old object no longer applies. Effect shapes (pinned
// corpus Oracle text): Hunt Down ("Target creature blocks target creature this
// turn if able."), Tangle Angler ("{G}: Target creature blocks this creature
// this turn if able."), Avalanche Tusker ("Whenever this creature attacks,
// target creature defending player controls blocks it this combat if able."),
// Mark for Death ("Target creature an opponent controls blocks this turn if
// able. ..."), Boros Battleshaper ("... up to one target creature attacks or
// blocks this combat if able ..."). Rulings (Scryfall, retained in
// .cardwork/rulings-*.json): Hunt Down 2007-10-01 — if the first creature
// can't block the second, the ability does nothing and the first creature
// blocks freely. Tangle Angler 2011-01-01 — activations stack; if Angler
// doesn't attack, or the target can't block it, the target blocks freely;
// tapped creatures, creatures that can't block and creatures not controlled
// by the defending player are exempt. Mark for Death 2013-01-24 — the
// controller still chooses which attacker it blocks; if it can't block the
// requirement does nothing. Boros Battleshaper 2013-04-15 — the creature
// attacks or blocks only if able when attackers or blockers are declared.

func mustBlockGame() *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
	g.Combat = &game.CombatState{}
	return g
}

func attackingDefender(g *game.Game, defender game.PlayerID, attackers ...*game.Permanent) {
	for _, attacker := range attackers {
		g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
			Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: defender},
		})
	}
}

// resolveBlockRequirement resolves a Hunt Down-shaped ApplyRule: target 0
// blocks target 1 if able, or blocks any attacker when attacker is nil.
func resolveBlockRequirement(g *game.Game, blocker, attacker *game.Permanent, duration game.EffectDuration) {
	effect := game.RuleEffect{Kind: game.RuleEffectMustBlock}
	targets := []game.Target{game.PermanentTarget(blocker.ObjectID)}
	if attacker != nil {
		effect.RequiredBlockAttackerRef = opt.Val(game.TargetPermanentReference(1))
		targets = append(targets, game.PermanentTarget(attacker.ObjectID))
	}
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Block Requirement Test", Types: []types.Card{types.Instant}}}),
		Controller: game.Player1,
		Targets:    targets,
	}
	resolveInstruction(NewEngine(nil), g, obj, game.ApplyRule{
		Object:      opt.Val(game.TargetPermanentReference(0)),
		RuleEffects: []game.RuleEffect{effect},
		Duration:    duration,
	}, nil)
}

func blockPairs(pairs ...[2]*game.Permanent) []game.BlockDeclaration {
	declarations := make([]game.BlockDeclaration, 0, len(pairs))
	for _, pair := range pairs {
		declarations = append(declarations, game.BlockDeclaration{Blocker: pair[0].ObjectID, Blocking: pair[1].ObjectID})
	}
	return declarations
}

func blocksLegal(g *game.Game, declarations []game.BlockDeclaration) bool {
	return combatEngine{}.blockersLegal(g, game.Player2, action.DeclareBlockersAction{Blockers: declarations})
}

func checkBlockLegality(t *testing.T, g *game.Game, cases map[string]struct {
	blocks []game.BlockDeclaration
	legal  bool
},
) {
	t.Helper()
	for name, c := range cases {
		if got := blocksLegal(g, c.blocks); got != c.legal {
			t.Errorf("%s: legal = %v, want %v", name, got, c.legal)
		}
	}
}

type legality = map[string]struct {
	blocks []game.BlockDeclaration
	legal  bool
}

// TestMustBlockDirectedRequiresThatPair resolves Hunt Down's shape: the first
// target must block the second, and every other declaration that leaves it
// unobeyed is illegal (CR 509.1c). Offered samples all obey it.
func TestMustBlockDirectedRequiresThatPair(t *testing.T) {
	g := mustBlockGame()
	hunted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	attackingDefender(g, game.Player2, hunted, other)
	forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	free := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	resolveBlockRequirement(g, forced, hunted, game.DurationThisTurn)

	checkBlockLegality(t, g, legality{
		"no blocks":                  {nil, false},
		"forced blocks other":        {blockPairs([2]*game.Permanent{forced, other}), false},
		"free blocks hunted":         {blockPairs([2]*game.Permanent{free, hunted}), false},
		"forced blocks hunted":       {blockPairs([2]*game.Permanent{forced, hunted}), true},
		"both block hunted":          {blockPairs([2]*game.Permanent{forced, hunted}, [2]*game.Permanent{free, hunted}), true},
		"forced hunted, free other":  {blockPairs([2]*game.Permanent{forced, hunted}, [2]*game.Permanent{free, other}), true},
		"free other, forced unmoved": {blockPairs([2]*game.Permanent{free, other}), false},
	})
	pair := game.BlockDeclaration{Blocker: forced.ObjectID, Blocking: hunted.ObjectID}
	samples := combatEngine{}.legalBlockers(g, game.Player2)
	if len(samples) == 0 {
		t.Fatal("no block declarations offered")
	}
	for _, sample := range samples {
		payload, _ := sample.DeclareBlockersPayload()
		if !slices.Contains(payload.Blockers, pair) {
			t.Fatalf("offered declaration %+v ignores the requirement", payload.Blockers)
		}
	}
}

// TestMustBlockDirectedDoesNothingWhenUnobeyable covers the rulings: a
// requirement whose named creature isn't attacking this player, can't be
// legally blocked, or is gone leaves the creature free (CR 509.1c).
func TestMustBlockDirectedDoesNothingWhenUnobeyable(t *testing.T) {
	t.Run("attacker attacks another player", func(t *testing.T) {
		g := mustBlockGame()
		hunted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		attackingDefender(g, game.Player3, hunted)
		attackingDefender(g, game.Player2, other)
		forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		resolveBlockRequirement(g, forced, hunted, game.DurationThisTurn)
		checkBlockLegality(t, g, legality{
			"no blocks":           {nil, true},
			"forced blocks other": {blockPairs([2]*game.Permanent{forced, other}), true},
		})
	})
	t.Run("restriction wins", func(t *testing.T) {
		g := mustBlockGame()
		flier := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Flying)
		attackingDefender(g, game.Player2, flier)
		forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		resolveBlockRequirement(g, forced, flier, game.DurationThisTurn)
		checkBlockLegality(t, g, legality{"no blocks": {nil, true}})
	})
	t.Run("named creature not attacking", func(t *testing.T) {
		g := mustBlockGame()
		angler := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		attackingDefender(g, game.Player2, other)
		forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		resolveBlockRequirement(g, forced, angler, game.DurationThisTurn)
		checkBlockLegality(t, g, legality{
			"no blocks":           {nil, true},
			"forced blocks other": {blockPairs([2]*game.Permanent{forced, other}), true},
		})
	})
	t.Run("tapped creature exempt", func(t *testing.T) {
		g := mustBlockGame()
		hunted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		attackingDefender(g, game.Player2, hunted)
		forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		resolveBlockRequirement(g, forced, hunted, game.DurationThisTurn)
		forced.Tapped = true
		checkBlockLegality(t, g, legality{"no blocks": {nil, true}})
	})
	t.Run("named creature gone at resolution", func(t *testing.T) {
		g := mustBlockGame()
		hunted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		forced := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		before := len(g.RuleEffects)
		g.Battlefield = slices.DeleteFunc(g.Battlefield, func(permanent *game.Permanent) bool { return permanent == hunted })
		resolveBlockRequirement(g, forced, hunted, game.DurationThisTurn)
		if len(g.RuleEffects) != before {
			t.Fatalf("rule effects = %+v, want the requirement dropped", g.RuleEffects[before:])
		}
	})
}

// TestMustBlockUndirectedAnyAttacker resolves Mark for Death's requirement:
// the creature must block, but its controller chooses which attacker.
func TestMustBlockUndirectedAnyAttacker(t *testing.T) {
	g := mustBlockGame()
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	attackingDefender(g, game.Player2, first, second)
	marked := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	free := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	resolveBlockRequirement(g, marked, nil, game.DurationThisTurn)
	checkBlockLegality(t, g, legality{
		"no blocks":            {nil, false},
		"marked blocks first":  {blockPairs([2]*game.Permanent{marked, first}), true},
		"marked blocks second": {blockPairs([2]*game.Permanent{marked, second}), true},
		"only free blocks":     {blockPairs([2]*game.Permanent{free, first}), false},
	})
}

// TestMustBlockCompetesWithLure maximizes obeyed requirements across a lure
// and directed requirements on a creature that can block only one attacker
// (CR 509.1c). Each Tangle Angler activation is a separate requirement.
func TestMustBlockCompetesWithLure(t *testing.T) {
	setup := func(activations int) (*game.Game, *game.Permanent, *game.Permanent, *game.Permanent) {
		g := mustBlockGame()
		lure := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		angler := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		attackingDefender(g, game.Player2, lure, angler)
		addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlockedByAllAble, lure.ObjectID)
		blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		for range activations {
			resolveBlockRequirement(g, blocker, angler, game.DurationThisTurn)
		}
		return g, lure, angler, blocker
	}
	g, lure, angler, blocker := setup(1)
	checkBlockLegality(t, g, legality{
		"no blocks":     {nil, false},
		"blocks lure":   {blockPairs([2]*game.Permanent{blocker, lure}), true},
		"blocks angler": {blockPairs([2]*game.Permanent{blocker, angler}), true},
	})
	g, lure, angler, blocker = setup(2)
	checkBlockLegality(t, g, legality{
		"blocks lure":   {blockPairs([2]*game.Permanent{blocker, lure}), false},
		"blocks angler": {blockPairs([2]*game.Permanent{blocker, angler}), true},
	})
}

// TestMustAttackOrBlockSatisfiedByEither resolves Boros Battleshaper's "attacks
// or blocks this combat if able" as both requirements on one creature: its
// controller is either attacking or defending, and either action obeys it. It
// ends with the combat.
func TestMustAttackOrBlockSatisfiedByEither(t *testing.T) {
	battleshaper := func(g *game.Game, target *game.Permanent) {
		obj := &game.StackObject{
			ID:         g.IDGen.Next(),
			Kind:       game.StackTriggeredAbility,
			Controller: game.Player1,
			Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
		}
		resolveInstruction(NewEngine(nil), g, obj, game.ApplyRule{
			Object:      opt.Val(game.TargetPermanentReference(0)),
			RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectMustAttack}, {Kind: game.RuleEffectMustBlock}},
			Duration:    game.DurationUntilEndOfCombat,
		}, nil)
	}
	t.Run("blocks", func(t *testing.T) {
		g := mustBlockGame()
		attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		attackingDefender(g, game.Player2, attacker)
		target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		battleshaper(g, target)
		checkBlockLegality(t, g, legality{
			"no blocks": {nil, false},
			"blocks":    {blockPairs([2]*game.Permanent{target, attacker}), true},
		})
		expireEndOfCombatRuleEffects(g)
		checkBlockLegality(t, g, legality{"no blocks after combat": {nil, true}})
	})
	t.Run("attacks", func(t *testing.T) {
		g := mustBlockGame()
		g.Turn.Step = game.StepDeclareAttackers
		target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		battleshaper(g, target)
		engine := combatEngine{}
		if engine.attackersLegal(g, game.Player1, action.DeclareAttackersAction{}) {
			t.Fatal("declaring no attackers is legal, want the creature forced to attack")
		}
		attack := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{attackAt(target, game.AttackTarget{Player: game.Player2})}}
		if !engine.attackersLegal(g, game.Player1, attack) {
			t.Fatal("attacking with the creature is illegal")
		}
	})
}

// TestMustBlockFollowsObjectAndDuration: a requirement names one object, so a
// blinked creature is free (CR 400.7). "This combat" ends with the combat;
// "this turn" applies in each declare blockers step that turn (CR 509.1c).
func TestMustBlockFollowsObjectAndDuration(t *testing.T) {
	setup := func(duration game.EffectDuration) (*game.Game, *game.Permanent, *game.Permanent) {
		g := mustBlockGame()
		attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		attackingDefender(g, game.Player2, attacker)
		blocker := addCombatPermanent(g, game.Player2, vanillaCreature("Blinked Blocker", 2, 2))
		resolveBlockRequirement(g, blocker, attacker, duration)
		return g, attacker, blocker
	}
	t.Run("blinked blocker", func(t *testing.T) {
		g, attacker, blocker := setup(game.DurationThisTurn)
		obj := &game.StackObject{
			ID:         g.IDGen.Next(),
			Kind:       game.StackSpell,
			SourceID:   addCardInstance(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Blink Spell"}}),
			Controller: game.Player2,
			Targets:    []game.Target{game.PermanentTarget(blocker.ObjectID)},
		}
		instructions := immediateBlinkInstructions(game.PutOnBattlefield{})
		for index := range instructions {
			NewEngine(nil).resolveInstructionWithChoices(g, obj, &instructions[index], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}
		var returned *game.Permanent
		for _, permanent := range g.Battlefield {
			if permanent.CardInstanceID == blocker.CardInstanceID {
				returned = permanent
			}
		}
		if returned == nil || returned.ObjectID == blocker.ObjectID {
			t.Fatal("blocker was not blinked into a new object")
		}
		checkBlockLegality(t, g, legality{
			"no blocks":       {nil, true},
			"returned blocks": {blockPairs([2]*game.Permanent{returned, attacker}), true},
		})
	})
	t.Run("this combat ends", func(t *testing.T) {
		g, _, _ := setup(game.DurationUntilEndOfCombat)
		checkBlockLegality(t, g, legality{"no blocks": {nil, false}})
		expireEndOfCombatRuleEffects(g)
		checkBlockLegality(t, g, legality{"no blocks after combat": {nil, true}})
	})
	t.Run("this turn spans combats", func(t *testing.T) {
		g, _, _ := setup(game.DurationThisTurn)
		expireEndOfCombatRuleEffects(g)
		checkBlockLegality(t, g, legality{"no blocks in a later combat": {nil, false}})
	})
}

// tuskerShapeCreature is Avalanche Tusker's attack trigger on a 4/4: "Whenever
// this creature attacks, target creature defending player controls blocks it
// this combat if able."
func tuskerShapeCreature() *game.CardDef {
	def := vanillaCreature("Tusker Shape Test", 4, 4)
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{
			Type:    game.TriggerWhenever,
			Pattern: game.TriggerPattern{Event: game.EventAttackerDeclared, Source: game.TriggerSourceSelf},
		},
		Content: game.Mode{
			Targets: []game.TargetSpec{creatureTargetSpec()},
			Sequence: []game.Instruction{{Primitive: game.ApplyRule{
				Object: opt.Val(game.TargetPermanentReference(0)),
				RuleEffects: []game.RuleEffect{{
					Kind:                     game.RuleEffectMustBlock,
					RequiredBlockAttackerRef: opt.Val(game.SourcePermanentReference()),
				}},
				Duration: game.DurationUntilEndOfCombat,
			}}},
		}.Ability(),
	}}
	return def
}

// TestMustBlockAttackTriggerForcesBlockInCombat plays a full combat: the
// attack trigger targets a defending creature, the defender tries to block
// with a different creature instead, the engine rejects that declaration and
// the targeted creature blocks the attacker.
func TestMustBlockAttackTriggerForcesBlockInCombat(t *testing.T) {
	g := bandCombatGame()
	tusker := addCombatPermanent(g, game.Player1, tuskerShapeCreature())
	forced := addCombatPermanent(g, game.Player2, vanillaCreature("Forced Blocker", 2, 2))
	chump := addCombatPermanent(g, game.Player2, vanillaCreature("Chump Blocker", 1, 1))
	life := g.Players[game.Player2].Life
	policies := simPassPolicies()
	policies[game.Player1] = &bandBlockingAgent{target: forced.ObjectID}
	policies[game.Player2] = &bandBlockingAgent{blocks: blockPairs([2]*game.Permanent{chump, tusker})}

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attackPlayer2(tusker), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if after.Execution.InvalidActions == 0 {
		t.Fatal("the defender's declaration ignoring the requirement was accepted")
	}
	declared := blockEvents(after.Events, game.EventBlockerDeclared)
	if _, blocked := declared[[2]id.ID{forced.ObjectID, tusker.ObjectID}]; !blocked {
		t.Fatalf("blocker declarations = %+v, want the targeted creature blocking the tusker", declared)
	}
	if after.Players[game.Player2].Life != life {
		t.Fatalf("defender life = %d, want %d: the tusker was blocked", after.Players[game.Player2].Life, life)
	}
	if _, alive := permanentByObjectID(after, forced.ObjectID); alive {
		t.Fatal("the forced 2/2 blocker survived the 4/4 tusker")
	}
	if slices.ContainsFunc(after.RuleEffects, func(effect game.RuleEffect) bool { return effect.Kind == game.RuleEffectMustBlock }) {
		t.Fatal("the this-combat requirement outlived the combat")
	}
}
