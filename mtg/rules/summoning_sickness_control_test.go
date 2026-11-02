package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3507. Official CR effective 2026-09-25 (verified local
// copy), read 302.6, 611.2a, 613.1b, 702.10b-c and 702.26d. 302.6: a
// creature's activated ability with {T} or {Q} in its cost can't be activated,
// and the creature can't attack, unless it has been under its controller's
// control continuously since their most recent turn began. 702.10b-c: haste
// lifts both restrictions. 702.26d: phasing doesn't change control. Cards
// (pinned corpus): Control Magic (oracle cd0d7141-46d2-4aa3-bc77-6b3b4513803e,
// {2}{U}{U} Aura) "Enchant creature / You control enchanted creature." (ruling
// 2016-06-08: gaining control of a creature doesn't gain control of attached
// Auras or Equipment). Llanowar Elves (oracle
// 68954295-54e3-4303-a6bc-fc4547a4e3a3, {G} 1/1) "{T}: Add {G}." (no rulings).
// Word of Seizing (oracle b80d2841-03b0-42d3-ac9b-51e793b1dc24, {3}{R}{R}
// instant) "Split second ... Untap target permanent and gain control of it until
// end of turn. It gains haste until end of turn." Disenchant (oracle
// a7e97fa9-4b72-4548-b854-5be5f18a6f1a, {1}{W} instant) "Destroy target
// artifact or enchantment." Patrol Signaler (oracle
// 1a8d7714-a566-4195-8605-550929584c61, {1}{W} 1/1) "{1}{W}, {Q}: Create a 1/1
// white Kithkin Soldier creature token." (rulings 2008-08-01: the summoning
// sickness rule applies to {Q}, ignored if the creature has haste; an untapped
// permanent can't pay {Q}).

// controlledSinceTurnBegan records, for a fixture that skips the untap step,
// that the permanent's current controller has controlled it continuously since
// their most recent turn began.
func controlledSinceTurnBegan(g *game.Game, permanent *game.Permanent) {
	permanent.SicknessController = opt.Val(effectiveController(g, permanent))
	permanent.SummoningSick = false
}

func sicknessMainPhaseGame(active game.PlayerID) *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSicknessMainPhase(g, active)
	return g
}

func setSicknessMainPhase(g *game.Game, active game.PlayerID) {
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = active, active
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	g.Combat = nil
}

// beginSicknessTurn starts player's next turn through the real beginning phase
// (untap, upkeep and draw) and leaves the game in that turn's precombat main
// phase.
func beginSicknessTurn(g *game.Game, engine *Engine, player game.PlayerID) {
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = player, player
	addCardToLibrary(g, player, &game.CardDef{CardFace: game.CardFace{Name: "Draw"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	setSicknessMainPhase(g, player)
}

// creatureCanAttack reports whether player's declare-attackers enumeration
// offers an attack with creature, and checks the real declaration driver
// agrees. The combat phase it opens and the attack's tapping are discarded
// afterward.
func creatureCanAttack(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, creature *game.Permanent) bool {
	t.Helper()
	phase, step, tapped := g.Turn.Phase, g.Turn.Step, creature.Tapped
	defer func() {
		g.Turn.Phase, g.Turn.Step = phase, step
		g.Combat = nil
		creature.Tapped = tapped
	}()
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	offered := slices.ContainsFunc(legalDeclareAttackersActions(g, player), func(act action.Action) bool {
		payload, ok := act.DeclareAttackersPayload()
		return ok && slices.ContainsFunc(payload.Attackers, func(d game.AttackDeclaration) bool {
			return d.Attacker == creature.ObjectID
		})
	})
	opponent := game.Player1
	if player == game.Player1 {
		opponent = game.Player2
	}
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		{Attacker: creature.ObjectID, Target: game.AttackTarget{Player: opponent}},
	}}
	accepted := engine.applyDeclareAttackers(g, player, declaration)
	if offered != accepted {
		t.Fatalf("attack enumeration offered %v but declaration driver accepted %v", offered, accepted)
	}
	return accepted
}

// elvesPayForElves reports whether player may cast a Llanowar Elves from hand
// with only the given Llanowar Elves permanent's {T}: Add {G} as a mana source.
func elvesPayForElves(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID) bool {
	t.Helper()
	spell := addCardToHand(g, player, registryCardDef(t, "Llanowar Elves"))
	defer g.Players[player].Hand.Remove(spell)
	return containsAction(engine.legalActions(g, player), action.CastSpell(spell, nil, 0, nil))
}

// TestControlMagicStolenCreatureIsSummoningSick is the #3507 scenario. Player2
// casts the registered Control Magic on Player1's established Llanowar Elves
// during Player2's turn. The Elves hasn't been under Player2's control since
// Player2's turn began, so it can't attack or {T} for mana. After Player2's next
// turn begins, it can do both.
func TestControlMagicStolenCreatureIsSummoningSick(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player2)
	engine := choicesProgramEngine(t)
	elves := addCombatPermanent(g, game.Player1, registryCardDef(t, "Llanowar Elves"))
	if permanentSummoningSick(g, elves) {
		t.Fatal("fixture: Player1's established Elves is summoning sick")
	}

	castAndResolveRegistered(t, g, engine, game.Player2, "Control Magic", mana.U, 4, elves)
	if effectiveController(g, elves) != game.Player2 {
		t.Fatal("Control Magic did not give Player2 control of the Elves")
	}
	if creatureCanAttack(t, g, engine, game.Player2, elves) {
		t.Fatal("the creature stolen this turn attacked without haste")
	}
	if elvesPayForElves(t, g, engine, game.Player2) {
		t.Fatal("the creature stolen this turn paid a {T} mana ability without haste")
	}
	view := NewObservation(g, game.Player2).Battlefield()
	if i := slices.IndexFunc(view, func(p PermanentView) bool { return p.ObjectID == elves.ObjectID }); i < 0 || !view[i].SummoningSick {
		t.Fatal("the observation did not report the stolen creature as summoning sick")
	}

	beginSicknessTurn(g, engine, game.Player1)
	if creatureCanAttack(t, g, engine, game.Player1, elves) {
		t.Fatal("the previous controller attacked with a creature it no longer controls")
	}
	beginSicknessTurn(g, engine, game.Player2)
	if !creatureCanAttack(t, g, engine, game.Player2, elves) {
		t.Fatal("the creature controlled since Player2's turn began cannot attack")
	}
	if !elvesPayForElves(t, g, engine, game.Player2) {
		t.Fatal("the creature controlled since Player2's turn began cannot tap for mana")
	}
}

// TestWordOfSeizingHasteAllowsStolenCreature proves haste still lifts the
// restriction: the registered Word of Seizing steals Player1's Elves and gives
// it haste, so Player2 may attack with it and tap it for mana this turn.
func TestWordOfSeizingHasteAllowsStolenCreature(t *testing.T) {
	for _, use := range []string{"attack", "mana"} {
		t.Run(use, func(t *testing.T) {
			g := sicknessMainPhaseGame(game.Player2)
			engine := choicesProgramEngine(t)
			elves := addCombatPermanent(g, game.Player1, registryCardDef(t, "Llanowar Elves"))
			elves.Tapped = true
			castAndResolveWordOfSeizing(t, g, engine, game.Player2, elves)
			if !permanentSummoningSick(g, elves) {
				t.Fatal("the stolen creature is not summoning sick")
			}
			if use == "attack" && !creatureCanAttack(t, g, engine, game.Player2, elves) {
				t.Fatal("the stolen creature with haste cannot attack")
			}
			if use == "mana" && !elvesPayForElves(t, g, engine, game.Player2) {
				t.Fatal("the stolen creature with haste cannot tap for mana")
			}
		})
	}
}

// TestControlChangingAwayAndBackRestartsSummoningSickness has Player2 take
// Player1's Elves with Control Magic, then Player1 destroy the Aura with the
// registered Disenchant during Player1's own turn. The Elves is back under
// Player1's control but not continuously since that turn began, so it can't
// attack or tap until Player1's next turn.
func TestControlChangingAwayAndBackRestartsSummoningSickness(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player2)
	engine := choicesProgramEngine(t)
	elves := addCombatPermanent(g, game.Player1, registryCardDef(t, "Llanowar Elves"))
	castAndResolveRegistered(t, g, engine, game.Player2, "Control Magic", mana.U, 4, elves)
	aura := g.Battlefield[len(g.Battlefield)-1]

	beginSicknessTurn(g, engine, game.Player1)
	if effectiveController(g, elves) != game.Player2 {
		t.Fatal("fixture: Control Magic no longer controls the Elves")
	}
	castAndResolveRegistered(t, g, engine, game.Player1, "Disenchant", mana.W, 2, aura)
	if effectiveController(g, elves) != game.Player1 {
		t.Fatal("the Elves did not return to Player1 when Control Magic was destroyed")
	}
	if creatureCanAttack(t, g, engine, game.Player1, elves) {
		t.Fatal("the creature returned this turn attacked without haste")
	}
	if elvesPayForElves(t, g, engine, game.Player1) {
		t.Fatal("the creature returned this turn tapped for mana without haste")
	}

	beginSicknessTurn(g, engine, game.Player2)
	beginSicknessTurn(g, engine, game.Player1)
	if !creatureCanAttack(t, g, engine, game.Player1, elves) {
		t.Fatal("the creature controlled since Player1's turn began cannot attack")
	}
}

// TestControlReturningBeforeTurnBeginsIsNotSummoningSick ends Word of Seizing's
// until-end-of-turn control in Player2's cleanup step. The Elves is back under
// Player1's control before Player1's turn begins, so it may attack and tap that
// turn.
func TestControlReturningBeforeTurnBeginsIsNotSummoningSick(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player2)
	engine := choicesProgramEngine(t)
	elves := addCombatPermanent(g, game.Player1, registryCardDef(t, "Llanowar Elves"))
	castAndResolveWordOfSeizing(t, g, engine, game.Player2, elves)
	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if effectiveController(g, elves) != game.Player1 {
		t.Fatal("Word of Seizing's control did not end in cleanup")
	}
	if !permanentSummoningSick(g, elves) {
		t.Fatal("the creature returned during cleanup is not sick before its owner's turn")
	}

	beginSicknessTurn(g, engine, game.Player1)
	if !creatureCanAttack(t, g, engine, game.Player1, elves) {
		t.Fatal("the creature returned before Player1's turn began cannot attack")
	}
	if !elvesPayForElves(t, g, engine, game.Player1) {
		t.Fatal("the creature returned before Player1's turn began cannot tap for mana")
	}
}

// TestControlMagicStolenUntapAbilityIsSummoningSick applies CR 302.6 to {Q}:
// Player2 steals Player1's tapped Patrol Signaler with Control Magic and can't
// activate its {1}{W}, {Q} ability that turn, but can once Player2's next turn
// has begun.
func TestControlMagicStolenUntapAbilityIsSummoningSick(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player2)
	engine := choicesProgramEngine(t)
	signaler := addCombatPermanent(g, game.Player1, registryCardDef(t, "Patrol Signaler"))
	signaler.Tapped = true
	castAndResolveRegistered(t, g, engine, game.Player2, "Control Magic", mana.U, 4, signaler)
	activate := action.ActivateAbility(signaler.ObjectID, 0, nil, 0)
	g.Players[game.Player2].ManaPool.Add(mana.W, 2)
	if containsAction(engine.legalActions(g, game.Player2), activate) {
		t.Fatal("the creature stolen this turn activated its {Q} ability without haste")
	}

	beginSicknessTurn(g, engine, game.Player1)
	beginSicknessTurn(g, engine, game.Player2)
	signaler.Tapped = true
	g.Players[game.Player2].ManaPool.Add(mana.W, 2)
	if !containsAction(engine.legalActions(g, game.Player2), activate) {
		t.Fatal("the creature controlled since Player2's turn began cannot activate its {Q} ability")
	}
}

// TestPhasingIsNotAControlChange keeps a stolen creature's control history
// across phasing (CR 702.26d, 702.26g): a creature Player2 has controlled
// through Control Magic since its turn began phases out with the Aura and both
// phase back in, so it can still attack.
func TestPhasingIsNotAControlChange(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player2)
	engine := choicesProgramEngine(t)
	elves := addCombatPermanent(g, game.Player1, registryCardDef(t, "Llanowar Elves"))
	castAndResolveRegistered(t, g, engine, game.Player2, "Control Magic", mana.U, 4, elves)
	aura := g.Battlefield[len(g.Battlefield)-1]
	beginSicknessTurn(g, engine, game.Player1)
	beginSicknessTurn(g, engine, game.Player2)

	elves.PhasedOut, aura.PhasedOut = true, true
	settleControlChanges(g)
	elves.PhasedOut, aura.PhasedOut = false, false
	settleControlChanges(g)
	if !creatureCanAttack(t, g, engine, game.Player2, elves) {
		t.Fatal("phasing out and in made the creature summoning sick")
	}
}
