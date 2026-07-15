package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// leylineLightningPlaneswalker adds a loyalty-3 planeswalker permanent for the
// controller, used to prove Leyline of Lightning may target a planeswalker.
func leylineLightningPlaneswalker(g *game.Game, controller game.PlayerID) *game.Permanent {
	planeswalker := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:    "Test Walker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
	}})
	planeswalker.Counters.Add(counter.Loyalty, 3)
	return planeswalker
}

// leylineLightningTargetAgent selects the target option whose permanent equals
// wantPermanent (falling back to option 0) and accepts every other choice, such
// as the resolution "Pay {1}?" prompt.
type leylineLightningTargetAgent struct {
	wantPermanent id.ID
}

func (*leylineLightningTargetAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *leylineLightningTargetAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		for _, option := range request.Options {
			if len(option.Targets) == 1 && option.Targets[0].PermanentID == a.wantPermanent {
				return []int{option.Index}
			}
		}
		return []int{0}
	}
	return []int{1}
}

// TestLeylineOfLightningTargetsPlayersAndPlaneswalkersOnly proves the cast
// trigger's targets are exactly players and planeswalkers: every player and an
// on-battlefield planeswalker are offered, and a creature is not.
func TestLeylineOfLightningTargetsPlayersAndPlaneswalkersOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfLightning())
	planeswalker := leylineLightningPlaneswalker(g, game.Player2)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	agent := &recordingChoiceAgent{answer: []int{0}}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agentsAll(agent), &TurnLog{}) {
		t.Fatal("cast trigger did not fire for the controller's spell")
	}

	var targetRequest *game.ChoiceRequest
	for i := range agent.requests {
		if agent.requests[i].Kind == game.ChoiceTarget {
			targetRequest = &agent.requests[i]
			break
		}
	}
	if targetRequest == nil {
		t.Fatal("no target choice was requested for the cast trigger")
	}

	sawPlayers := map[game.PlayerID]bool{}
	sawPlaneswalker, sawCreature := false, false
	for _, option := range targetRequest.Options {
		if len(option.Targets) != 1 {
			t.Fatalf("target option %d has %d targets, want 1", option.Index, len(option.Targets))
		}
		target := option.Targets[0]
		switch target.Kind {
		case game.TargetPlayer:
			sawPlayers[target.PlayerID] = true
		case game.TargetPermanent:
			if target.PermanentID == planeswalker.ObjectID {
				sawPlaneswalker = true
			}
			if target.PermanentID == creature.ObjectID {
				sawCreature = true
			}
		default:
		}
	}
	for seat := game.PlayerID(0); int(seat) < game.NumPlayers; seat++ {
		if !sawPlayers[seat] {
			t.Errorf("player %v was not an offered target", seat)
		}
	}
	if !sawPlaneswalker {
		t.Error("planeswalker was not an offered target")
	}
	if sawCreature {
		t.Error("a nonplaneswalker creature was offered as a target")
	}
}

// TestLeylineOfLightningPaysOneToDealDamage proves that when the controller pays
// {1}, the enchantment deals 1 damage to the chosen target player.
func TestLeylineOfLightningPaysOneToDealDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfLightning())
	land := addBasicLandPermanent(g, game.Player1, types.Mountain)

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	agent := &choiceOnlyAgent{choices: [][]int{{0}, {1}}}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agentsAll(agent), &TurnLog{}) {
		t.Fatal("cast trigger did not fire")
	}
	top, ok := g.Stack.Peek()
	if !ok || len(top.Targets) != 1 || top.Targets[0].Kind != game.TargetPlayer {
		t.Fatalf("stack target = %+v, want one player target", top.Targets)
	}
	targetPlayer := top.Targets[0].PlayerID
	before := g.Players[targetPlayer].Life

	engine.resolveTopOfStackWithChoices(g, agentsAll(agent), &TurnLog{})

	if got := before - g.Players[targetPlayer].Life; got != 1 {
		t.Fatalf("damage to target player = %d, want 1", got)
	}
	if !land.Tapped {
		t.Fatal("land was not tapped to pay {1}")
	}
}

// TestLeylineOfLightningDeclinedPaymentDealsNoDamage proves the "you may pay {1}"
// is optional: declining deals no damage and pays nothing.
func TestLeylineOfLightningDeclinedPaymentDealsNoDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfLightning())
	land := addBasicLandPermanent(g, game.Player1, types.Mountain)

	before := [game.NumPlayers]int{}
	for seat := game.PlayerID(0); int(seat) < game.NumPlayers; seat++ {
		before[seat] = g.Players[seat].Life
	}

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	agent := &choiceOnlyAgent{choices: [][]int{{0}, {0}}}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agentsAll(agent), &TurnLog{}) {
		t.Fatal("cast trigger did not fire")
	}
	engine.resolveTopOfStackWithChoices(g, agentsAll(agent), &TurnLog{})

	for seat := game.PlayerID(0); int(seat) < game.NumPlayers; seat++ {
		if got := g.Players[seat].Life; got != before[seat] {
			t.Fatalf("player %v life = %d, want %d (no damage after declining)", seat, got, before[seat])
		}
	}
	if land.Tapped {
		t.Fatal("land was tapped despite declining the payment")
	}
}

// TestLeylineOfLightningDoesNotTriggerForOpponentSpell proves the cast trigger is
// controller-scoped: an opponent casting a spell does not fire it.
func TestLeylineOfLightningDoesNotTriggerForOpponentSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfLightning())

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player2})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cast trigger fired for an opponent's spell")
	}
}

// TestLeylineOfLightningDamagesTargetedPlaneswalker proves the damage lands on a
// chosen planeswalker, removing one loyalty counter.
func TestLeylineOfLightningDamagesTargetedPlaneswalker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.LeylineOfLightning())
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	planeswalker := leylineLightningPlaneswalker(g, game.Player2)

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	agent := &leylineLightningTargetAgent{wantPermanent: planeswalker.ObjectID}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agentsAll(agent), &TurnLog{}) {
		t.Fatal("cast trigger did not fire")
	}
	top, ok := g.Stack.Peek()
	if !ok || len(top.Targets) != 1 || top.Targets[0].PermanentID != planeswalker.ObjectID {
		t.Fatalf("stack target = %+v, want the planeswalker", top.Targets)
	}

	engine.resolveTopOfStackWithChoices(g, agentsAll(agent), &TurnLog{})

	if got := planeswalker.Counters.Get(counter.Loyalty); got != 2 {
		t.Fatalf("planeswalker loyalty = %d, want 2 (1 damage removed one loyalty)", got)
	}
}
