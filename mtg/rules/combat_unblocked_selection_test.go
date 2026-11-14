package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3893 (part of #2226). Official CR effective
// 2026-11-06 (verified local copy MagicCompRules-20261106.txt), read 506.3d,
// 506.4, 508.4, 508.4d, 509.1, 509.1h and 611.2c. 509.1h: an attacking creature
// with no creatures declared as blockers becomes unblocked, and one that was
// blocked stays blocked even if all its blockers are removed from combat.
// 508.4d/506.3d: a creature put onto the battlefield attacking during the
// declare blockers step or later enters unblocked. 506.4: a creature removed
// from combat stops being unblocked. 611.2c: a resolved effect's affected set
// is locked when it begins. Card (pinned corpus): Inkfathom Witch (oracle
// 60b6e96a-63e6-4520-b2dd-b20b7dfbcd57, {1}{U/B} Creature — Merfolk Wizard 1/1)
// "Fear (This creature can't be blocked except by artifact creatures and/or
// black creatures.) {2}{U}{B}: Each unblocked creature has base power and
// toughness 4/1 until end of turn." Rulings: 2008-05-01 an unblocked creature
// is one that attacked and wasn't blocked; before the declare blockers step
// (or after combat) the ability has no effect; 2009-10-01 it doesn't override
// later P/T-setting effects or any modifications such as counters. The card is
// authored here only as a test definition; a later card unit registers it.

func inkfathomWitchTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Inkfathom Witch",
		ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.HybridMana(mana.U, mana.B)}),
		Colors:    []color.Color{color.Blue, color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FearStaticBody,
		},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:           "{2}{U}{B}: Each unblocked creature has base power and toughness 4/1 until end of turn.",
			ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.U, cost.B}),
			ZoneOfFunction: zone.Battlefield,
			Content: game.AbilityContent{
				Modes: []game.Mode{{
					Sequence: []game.Instruction{{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{{
								Duration:     game.DurationPermanent,
								Group:        game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateUnblocked}),
								Layer:        game.LayerPowerToughnessSet,
								SetPower:     opt.Val(game.PT{Value: 4}),
								SetToughness: opt.Val(game.PT{Value: 1}),
							}},
							Duration: game.DurationUntilEndOfTurn,
						},
					}},
				}},
				MinModes: 1,
				MaxModes: 1,
			},
		}},
	}}
}

func unblockedTestCreature(g *game.Game, controller game.PlayerID) *game.Permanent {
	creature := addCombatCreaturePermanentWithPower(g, controller, 2)
	creature.SummoningSick = false
	return creature
}

func declareUnblockedTestAttacks(t *testing.T, g *game.Game, engine *Engine, defender game.PlayerID, attackers ...*game.Permanent) {
	t.Helper()
	declaration := action.DeclareAttackersAction{}
	for _, attacker := range attackers {
		declaration.Attackers = append(declaration.Attackers, attackAt(attacker, game.AttackTarget{Player: defender}))
	}
	if !engine.applyDeclareAttackers(g, g.Turn.ActivePlayer, declaration) {
		t.Fatal("attack declaration was rejected")
	}
}

func declareUnblockedTestBlocks(t *testing.T, g *game.Game, engine *Engine, defender game.PlayerID, blocks ...game.BlockDeclaration) {
	t.Helper()
	g.Turn.Step = game.StepDeclareBlockers
	if !engine.applyDeclareBlockers(g, defender, action.DeclareBlockersAction{Blockers: blocks}) {
		t.Fatal("block declaration was rejected")
	}
}

func activateInkfathomWitch(t *testing.T, g *game.Game, engine *Engine, witch *game.Permanent) {
	t.Helper()
	controller := witch.Controller
	g.Turn.PriorityPlayer = controller
	pool := &g.Players[controller].ManaPool
	pool.Add(mana.U, 1)
	pool.Add(mana.B, 1)
	pool.Add(mana.C, 2)
	agents := simPassPolicies()
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, controller, action.ActivateAbility(witch.ObjectID, 0, nil, 0), agents, log) {
		t.Fatal("Inkfathom Witch activation failed")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if !g.Stack.IsEmpty() {
		t.Fatal("Inkfathom Witch ability did not resolve")
	}
}

func assertUnblockedTestPT(t *testing.T, g *game.Game, label string, permanent *game.Permanent, power, toughness int) {
	t.Helper()
	gotToughness, _ := effectiveToughness(g, permanent)
	if gotPower := effectivePower(g, permanent); gotPower != power || gotToughness != toughness {
		t.Errorf("%s = %d/%d, want %d/%d", label, gotPower, gotToughness, power, toughness)
	}
}

// TestInkfathomWitchSetsEachUnblockedAttacker plays issue #3893 scenario 1:
// Player1 attacks with three creatures, one carrying a +1/+1 counter, and
// Player2 blocks one. After blocks, Inkfathom Witch makes both unblocked
// attackers base 4/1; the counter still applies over the set (layer 7c), and
// the blocked attacker, the blocker and non-attacking creatures are unchanged.
// Combat damage to Player2 reflects the new power.
func TestInkfathomWitchSetsEachUnblockedAttacker(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	witch := addCombatPermanent(g, game.Player1, inkfathomWitchTestDef())
	countered := unblockedTestCreature(g, game.Player1)
	countered.Counters.Add(counter.PlusOnePlusOne, 1)
	plain := unblockedTestCreature(g, game.Player1)
	blocked := unblockedTestCreature(g, game.Player1)
	bystander := unblockedTestCreature(g, game.Player1)
	blocker := unblockedTestCreature(g, game.Player2)
	idle := unblockedTestCreature(g, game.Player2)

	declareUnblockedTestAttacks(t, g, engine, game.Player2, countered, plain, blocked)
	declareUnblockedTestBlocks(t, g, engine, game.Player2, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: blocked.ObjectID})
	activateInkfathomWitch(t, g, engine, witch)

	assertUnblockedTestPT(t, g, "unblocked attacker", plain, 4, 1)
	assertUnblockedTestPT(t, g, "unblocked attacker with a +1/+1 counter", countered, 5, 2)
	assertUnblockedTestPT(t, g, "blocked attacker", blocked, 2, 2)
	assertUnblockedTestPT(t, g, "blocker", blocker, 2, 2)
	assertUnblockedTestPT(t, g, "non-attacking creature", bystander, 2, 2)
	assertUnblockedTestPT(t, g, "defender's idle creature", idle, 2, 2)
	assertUnblockedTestPT(t, g, "Inkfathom Witch", witch, 1, 1)

	life := g.Players[game.Player2].Life
	engine.resolveCombatDamage(g, &TurnLog{})
	if lost := life - g.Players[game.Player2].Life; lost != 9 {
		t.Fatalf("Player2 lost %d life, want 9 from the 4/1 and 5/2 unblocked attackers", lost)
	}
}

// TestInkfathomWitchSkipsAttackerWhoseBlockerLeftCombat plays issue #3893
// scenario 2: an attacker whose only blocker was removed from combat after
// blocks stays blocked (CR 509.1h), so the Witch doesn't affect it.
func TestInkfathomWitchSkipsAttackerWhoseBlockerLeftCombat(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	witch := addCombatPermanent(g, game.Player1, inkfathomWitchTestDef())
	blocked := unblockedTestCreature(g, game.Player1)
	unblocked := unblockedTestCreature(g, game.Player1)
	blocker := unblockedTestCreature(g, game.Player2)

	declareUnblockedTestAttacks(t, g, engine, game.Player2, blocked, unblocked)
	declareUnblockedTestBlocks(t, g, engine, game.Player2, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: blocked.ObjectID})
	removePermanentFromCombat(g, blocker.ObjectID)
	activateInkfathomWitch(t, g, engine, witch)

	assertUnblockedTestPT(t, g, "attacker whose blocker left combat", blocked, 2, 2)
	assertUnblockedTestPT(t, g, "unblocked attacker", unblocked, 4, 1)
}

// TestInkfathomWitchBeforeBlocksAffectsNothing plays issue #3893 scenario 3:
// activated in the declare attackers step, before blockers are declared, no
// creature is unblocked, and the set is locked at resolution (CR 611.2c), so
// attackers left unblocked afterward are still unaffected.
func TestInkfathomWitchBeforeBlocksAffectsNothing(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	witch := addCombatPermanent(g, game.Player1, inkfathomWitchTestDef())
	attacker := unblockedTestCreature(g, game.Player1)

	declareUnblockedTestAttacks(t, g, engine, game.Player2, attacker)
	activateInkfathomWitch(t, g, engine, witch)
	assertUnblockedTestPT(t, g, "attacker before blocks", attacker, 2, 2)

	declareUnblockedTestBlocks(t, g, engine, game.Player2)
	if !combatStateMatches(g, attacker, game.CombatStateUnblocked) {
		t.Fatal("attacker with no blockers is not unblocked after blocks")
	}
	assertUnblockedTestPT(t, g, "attacker left unblocked after the ability resolved", attacker, 2, 2)
}

// TestInkfathomWitchAffectsOpponentsUnblockedAttackers plays issue #3893
// scenario 4: Player2 activates Inkfathom Witch during Player1's combat. "Each
// unblocked creature" isn't limited to its controller, so Player1's unblocked
// attacker becomes 4/1 while the blocked one and Player2's creatures don't.
func TestInkfathomWitchAffectsOpponentsUnblockedAttackers(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	witch := addCombatPermanent(g, game.Player2, inkfathomWitchTestDef())
	unblocked := unblockedTestCreature(g, game.Player1)
	blocked := unblockedTestCreature(g, game.Player1)
	blocker := unblockedTestCreature(g, game.Player2)

	declareUnblockedTestAttacks(t, g, engine, game.Player2, unblocked, blocked)
	declareUnblockedTestBlocks(t, g, engine, game.Player2, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: blocked.ObjectID})
	activateInkfathomWitch(t, g, engine, witch)

	assertUnblockedTestPT(t, g, "opponent's unblocked attacker", unblocked, 4, 1)
	assertUnblockedTestPT(t, g, "opponent's blocked attacker", blocked, 2, 2)
	assertUnblockedTestPT(t, g, "Witch controller's blocker", blocker, 2, 2)
	assertUnblockedTestPT(t, g, "Inkfathom Witch", witch, 1, 1)
}

// TestUnblockedCreatureTargetRecheckedOnResolution checks "target unblocked
// creature": only an unblocked attacker is a legal choice, and a target removed
// from combat before resolution is no longer unblocked (CR 506.4), so the spell
// doesn't resolve (CR 608.2b).
func TestUnblockedCreatureTargetRecheckedOnResolution(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	unblocked := unblockedTestCreature(g, game.Player1)
	blocked := unblockedTestCreature(g, game.Player1)
	blocker := unblockedTestCreature(g, game.Player2)
	spell := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Unblocked Shock",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target unblocked creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateUnblocked}),
			}},
			Sequence: []game.Instruction{{
				Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)},
			}},
		}.Ability()),
	}})

	declareUnblockedTestAttacks(t, g, engine, game.Player2, unblocked, blocked)
	g.Turn.PriorityPlayer = game.Player2
	castAt := func(target *game.Permanent) action.Action {
		return action.CastSpell(spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	}
	if actionsContain(engine.legalActions(g, game.Player2), castAt(unblocked)) {
		t.Fatal("an attacker was a legal unblocked target before blockers were declared")
	}

	declareUnblockedTestBlocks(t, g, engine, game.Player2, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: blocked.ObjectID})
	g.Turn.PriorityPlayer = game.Player2
	legal := engine.legalActions(g, game.Player2)
	if actionsContain(legal, castAt(blocked)) {
		t.Fatal("a blocked attacker was a legal unblocked target")
	}
	if !actionsContain(legal, castAt(unblocked)) {
		t.Fatal("the unblocked attacker was not a legal target")
	}
	agents := simPassPolicies()
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player2, castAt(unblocked), agents, log) {
		t.Fatal("casting at the unblocked attacker failed")
	}
	removePermanentFromCombat(g, unblocked.ObjectID)
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if unblocked.MarkedDamage != 0 {
		t.Fatalf("creature removed from combat took %d damage, want the spell to fail to resolve", unblocked.MarkedDamage)
	}
}

// TestCreatureEnteringAttackingAfterBlocksIsUnblocked checks CR 508.4d: a
// creature put onto the battlefield attacking a player nothing else attacked,
// after blockers were declared, is unblocked.
func TestCreatureEnteringAttackingAfterBlocksIsUnblocked(t *testing.T) {
	g := bandCombatGame()
	engine := NewEngine(nil)
	attacker := unblockedTestCreature(g, game.Player1)
	declareUnblockedTestAttacks(t, g, engine, game.Player2, attacker)
	declareUnblockedTestBlocks(t, g, engine, game.Player2)

	late := unblockedTestCreature(g, game.Player1)
	declareTokenAttackingDefender(g, game.Player1, late, game.Player3)
	if !combatStateMatches(g, late, game.CombatStateUnblocked) {
		t.Fatal("creature put onto the battlefield attacking after blocks is not unblocked")
	}
	if !combatStateMatches(g, attacker, game.CombatStateUnblocked) {
		t.Fatal("declared attacker with no blockers is not unblocked")
	}
}
