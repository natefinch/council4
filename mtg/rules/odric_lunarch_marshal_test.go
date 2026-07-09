package rules

import (
	"testing"

	cardo "github.com/natefinch/council4/mtg/cards/o"
	"github.com/natefinch/council4/mtg/game"
)

// stageOdricCombat puts the real Odric, Lunarch Marshal onto Player1's
// battlefield and stages Player1's beginning-of-combat step so a test can fire
// Odric's real beginning-of-combat keyword-sharing trigger.
func stageOdricCombat(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	odric := addCombatPermanent(g, game.Player1, cardo.OdricLunarchMarshal())
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{}
	return g, engine, odric
}

// fireOdricBeginningOfCombat drives Odric's real beginning-of-combat trigger:
// it emits the turn-based event, puts the trigger on the stack, and resolves it.
func fireOdricBeginningOfCombat(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	agents := [game.NumPlayers]PlayerAgent{}
	for p := range agents {
		agents[p] = defaultChoiceAgent{}
	}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Odric beginning-of-combat trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
}

// TestOdricSharesFirstStrikeToAllYourCreatures proves the core mechanic: when a
// creature you control has first strike, at the beginning of combat every
// creature you control (including Odric) gains first strike until end of turn.
func TestOdricSharesFirstStrikeToAllYourCreatures(t *testing.T) {
	g, engine, odric := stageOdricCombat(t)
	seed := addCombatCreaturePermanent(g, game.Player1, game.FirstStrike)
	vanilla := addCombatCreaturePermanent(g, game.Player1)

	fireOdricBeginningOfCombat(t, g, engine)

	for name, p := range map[string]*game.Permanent{"Odric": odric, "seed": seed, "vanilla": vanilla} {
		if !hasKeyword(g, p, game.FirstStrike) {
			t.Fatalf("%s did not gain first strike from Odric", name)
		}
	}
}

// TestOdricSharesAcrossTheFullKeywordList proves the "the same is true for ..."
// clause: every keyword in Odric's list is shared independently. A creature you
// control that has vigilance grants vigilance to all your creatures, while a
// keyword no creature you control has (flying) is not granted to anyone.
func TestOdricSharesAcrossTheFullKeywordList(t *testing.T) {
	g, engine, odric := stageOdricCombat(t)
	// One creature seeds vigilance, another seeds trample; nobody has flying.
	addCombatCreaturePermanent(g, game.Player1, game.Vigilance)
	addCombatCreaturePermanent(g, game.Player1, game.Trample)
	vanilla := addCombatCreaturePermanent(g, game.Player1)

	fireOdricBeginningOfCombat(t, g, engine)

	if !hasKeyword(g, vanilla, game.Vigilance) {
		t.Fatal("vanilla creature did not gain shared vigilance")
	}
	if !hasKeyword(g, vanilla, game.Trample) {
		t.Fatal("vanilla creature did not gain shared trample")
	}
	if !hasKeyword(g, odric, game.Vigilance) || !hasKeyword(g, odric, game.Trample) {
		t.Fatal("Odric did not gain the shared keywords")
	}
	if hasKeyword(g, vanilla, game.Flying) {
		t.Fatal("flying was granted even though no creature you control has it")
	}
	if hasKeyword(g, odric, game.Flying) {
		t.Fatal("Odric gained flying even though no creature you control has it")
	}
}

// TestOdricDoesNotShareToOpponents proves the group grant is scoped to
// "creatures you control": a creature an opponent controls neither satisfies the
// gate nor receives the shared keyword.
func TestOdricDoesNotShareToOpponents(t *testing.T) {
	g, engine, _ := stageOdricCombat(t)
	yours := addCombatCreaturePermanent(g, game.Player1, game.FirstStrike)
	theirs := addCombatCreaturePermanent(g, game.Player2)

	fireOdricBeginningOfCombat(t, g, engine)

	if !hasKeyword(g, yours, game.FirstStrike) {
		t.Fatal("your creature did not gain first strike")
	}
	if hasKeyword(g, theirs, game.FirstStrike) {
		t.Fatal("an opponent's creature gained first strike from your Odric")
	}
}

// TestOdricGateIsControllerScoped proves the "if a creature you control has it"
// gate is scoped to Odric's controller: an opponent's first striker does not
// satisfy the gate, so your creatures gain nothing.
func TestOdricGateIsControllerScoped(t *testing.T) {
	g, engine, odric := stageOdricCombat(t)
	yourVanilla := addCombatCreaturePermanent(g, game.Player1)
	addCombatCreaturePermanent(g, game.Player2, game.FirstStrike)

	fireOdricBeginningOfCombat(t, g, engine)

	if hasKeyword(g, yourVanilla, game.FirstStrike) {
		t.Fatal("your creature gained first strike from an opponent's first striker")
	}
	if hasKeyword(g, odric, game.FirstStrike) {
		t.Fatal("Odric gained first strike from an opponent's first striker")
	}
}
