package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/k"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestKutzilStaticStopsOpponentSpellsDuringControllerTurn proves Kutzil's static
// ("Your opponents can't cast spells during your turn.") restricts opponents only
// while its controller is the active player, and never restricts the controller.
func TestKutzilStaticStopsOpponentSpellsDuringControllerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, k.KutzilMalametExemplar())
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Test Bolt", Types: []types.Card{types.Instant}}}

	g.Turn.ActivePlayer = game.Player1
	if !spellCastProhibited(g, game.Player2, spell) {
		t.Fatal("opponent should not be able to cast a spell during Kutzil's controller's turn")
	}
	if spellCastProhibited(g, game.Player1, spell) {
		t.Fatal("Kutzil never restricts its own controller from casting spells")
	}

	g.Turn.ActivePlayer = game.Player2
	if spellCastProhibited(g, game.Player2, spell) {
		t.Fatal("the turn-scoped restriction must lift on the opponent's own turn")
	}
}

// TestKutzilTriggerDrawsWhenBuffedCreatureDealsCombatDamage proves Kutzil's
// batched combat-damage trigger fires when a creature its controller controls
// whose current power exceeds its base power (here, raised by a +1/+1 counter)
// deals combat damage to a player, drawing one card. An identical creature with
// no such raise does not satisfy the PowerAboveBase filter and draws nothing.
func TestKutzilTriggerDrawsWhenBuffedCreatureDealsCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, k.KutzilMalametExemplar())
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	buffed := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	buffed.Counters.Add(counter.PlusOnePlusOne, 1)

	dealPlayerDamage(g, buffed.CardInstanceID, buffed.ObjectID, game.Player1, game.Player2, 3, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Kutzil's combat-damage trigger was not put on the stack for a buffed creature")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want Kutzil to draw one card for the buffed creature", got)
	}
}

// TestKutzilTriggerIgnoresUnbuffedCreatureCombatDamage proves the PowerAboveBase
// filter fails closed: a creature whose current power equals its base power does
// not satisfy the trigger, so no card is drawn.
func TestKutzilTriggerIgnoresUnbuffedCreatureCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, k.KutzilMalametExemplar())
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	unbuffed := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	dealPlayerDamage(g, unbuffed.CardInstanceID, unbuffed.ObjectID, game.Player1, game.Player2, 2, true)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Kutzil's trigger fired for a creature whose power does not exceed its base power")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want no draw when the creature is not buffed above its base power", got)
	}
}

// TestKutzilTriggerFiresWhenBuffedCreatureDiesDealingCombatDamage proves the
// PowerAboveBase filter uses last-known information: a buffed creature that deals
// combat damage to a player and leaves the battlefield in the same step (a
// trampler dying to its blocker while connecting) still satisfies the trigger, so
// Kutzil draws using the creature's last-known power.
func TestKutzilTriggerFiresWhenBuffedCreatureDiesDealingCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, k.KutzilMalametExemplar())
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	buffed := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	buffed.Counters.Add(counter.PlusOnePlusOne, 1)

	dealPlayerDamage(g, buffed.CardInstanceID, buffed.ObjectID, game.Player1, game.Player2, 3, true)

	// The creature dies in the same combat-damage step; its last-known snapshot
	// records current power 3 above base power 2.
	snapshot := snapshotPermanent(g, buffed, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	if _, ok := removePermanentFromBattlefield(g, buffed.ObjectID); !ok {
		t.Fatal("failed to remove the buffed creature from the battlefield")
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Kutzil's trigger did not fire for a buffed creature that died dealing combat damage")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want Kutzil to draw using the dead creature's last-known power", got)
	}
}
