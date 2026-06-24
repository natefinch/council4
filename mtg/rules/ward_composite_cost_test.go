package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func addCompositeWardPermanent(g *game.Game, controller game.PlayerID, manaCost cost.Mana, additional []cost.Additional) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Composite Ward Creature",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.WardStaticAbilityWithCosts(manaCost, additional)},
	}})
}

// Ward—Pay 2 life.: the controller of the targeting spell pays the life and the
// spell stays on the stack.
func TestWardPayLifePaidLeavesSpellOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addCompositeWardPermanent(g, game.Player2, nil, []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}})
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1
	startLife := g.Players[game.Player1].Life

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != startLife-2 {
		t.Fatalf("life = %d, want %d (ward pay-2-life)", got, startLife-2)
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want targeting spell still on stack", g.Stack.Size())
	}
	if g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("spell moved to graveyard after ward was paid")
	}
}

// Ward—{2}, Pay 2 life.: the composite mana-plus-life cost is paid together and
// the spell stays on the stack (Captain Howler, Sea Scourge).
func TestWardCompositeManaAndLifePaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addCompositeWardPermanent(g, game.Player2, cost.Mana{cost.O(2)}, []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}})
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	island1 := addBasicLandPermanent(g, game.Player1, types.Island)
	island2 := addBasicLandPermanent(g, game.Player1, types.Island)
	g.Turn.PriorityPlayer = game.Player1
	startLife := g.Players[game.Player1].Life

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != startLife-2 {
		t.Fatalf("life = %d, want %d (ward pay-2-life)", got, startLife-2)
	}
	if !island1.Tapped || !island2.Tapped {
		t.Fatal("ward mana component did not tap both Islands")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want targeting spell still on stack", g.Stack.Size())
	}
}

// When the warded permanent's controller cannot or will not pay the composite
// cost, the targeting spell is countered.
func TestWardCompositeUnpaidCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// A {20} mana component is unpayable, so the ward counters the spell.
	warded := addCompositeWardPermanent(g, game.Player2, cost.Mana{cost.O(20)}, []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}})
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want ward to counter targeting spell", g.Stack.Size())
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("countered spell did not move to graveyard")
	}
}
