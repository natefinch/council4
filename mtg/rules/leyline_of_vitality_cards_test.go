package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// leylineVitalityEnterCreature makes a plain creature enter the battlefield
// under controller from hand, emitting the enters-the-battlefield event that
// Leyline of Vitality's trigger watches.
func leylineVitalityEnterCreature(t *testing.T, g *game.Game, controller game.PlayerID) *game.Permanent {
	t.Helper()
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Owner: controller,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:      "Grizzly Bears",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		}},
	}
	permanent, ok := createCardPermanent(g, g.CardInstances[cardID], controller, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent() = false, want true")
	}
	return permanent
}

// TestLeylineOfVitalityAnthemGivesPlusZeroPlusOne proves the real card's
// "Creatures you control get +0/+1." raises only toughness for the controller's
// creatures.
func TestLeylineOfVitalityAnthemGivesPlusZeroPlusOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cards.LeylineOfVitality())
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("creature power = %d, want 2 (+0)", got)
	}
	toughness, ok := effectiveToughness(g, creature)
	if !ok || toughness != 3 {
		t.Fatalf("creature toughness = %d (ok=%v), want 3 (+1)", toughness, ok)
	}
}

// TestLeylineOfVitalityAnthemIsControllerScoped proves an opponent's creature is
// not buffed by the controller's Leyline of Vitality.
func TestLeylineOfVitalityAnthemIsControllerScoped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cards.LeylineOfVitality())
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	toughness, ok := effectiveToughness(g, opponentCreature)
	if !ok || toughness != 2 {
		t.Fatalf("opponent creature toughness = %d (ok=%v), want 2 (unbuffed)", toughness, ok)
	}
}

// TestLeylineOfVitalityGainsLifeWhenControllerCreatureEnters proves the "Whenever
// a creature you control enters, you may gain 1 life." trigger fires for the
// controller's entering creature and, when accepted, gains exactly one life.
func TestLeylineOfVitalityGainsLifeWhenControllerCreatureEnters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfVitality())
	agents := agentsAll(&recordingChoiceAgent{answer: []int{1}})
	before := g.Players[game.Player1].Life

	leylineVitalityEnterCreature(t, g, game.Player1)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) {
		t.Fatal("creature-entered trigger did not fire for the controller's creature")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life - before; got != 1 {
		t.Fatalf("life gained = %d, want 1", got)
	}
}

// TestLeylineOfVitalityLifeGainIsOptional proves the "you may" is honored: a
// declining controller gains no life.
func TestLeylineOfVitalityLifeGainIsOptional(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfVitality())
	agents := agentsAll(&recordingChoiceAgent{answer: []int{0}})
	before := g.Players[game.Player1].Life

	leylineVitalityEnterCreature(t, g, game.Player1)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life - before; got != 0 {
		t.Fatalf("life gained after declining = %d, want 0", got)
	}
}

// TestLeylineOfVitalityDoesNotTriggerForOpponentCreature proves the trigger is
// controller-scoped: an opponent's creature entering does not fire it.
func TestLeylineOfVitalityDoesNotTriggerForOpponentCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfVitality())

	leylineVitalityEnterCreature(t, g, game.Player2)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature-entered trigger fired for an opponent-controlled creature")
	}
}
