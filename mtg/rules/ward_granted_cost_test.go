package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addPayLifeWardGrantSource adds a permanent that grants Ward—Pay 2 life to the
// other creatures its controller controls, mirroring Hexing Squelcher's
// "Other creatures you control have \"Ward—Pay 2 life.\"" static grant. The
// source itself is excluded from the grant and has no native Ward.
func addPayLifeWardGrantSource(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Ward Grant Source",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroupExcluding(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
					game.SourcePermanentReference(),
				),
				AddAbilities: []game.Ability{
					new(game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
						{Kind: cost.AdditionalPayLife, Text: "Pay 2 life", Amount: 2},
					})),
				},
			}},
		}},
	}})
}

// addVanillaCreature adds a plain creature with no abilities of its own.
func addVanillaCreature(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Beneficiary",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
}

// A creature that gains Ward—Pay N life from a controlled static grant triggers
// the ward when an opponent targets it; the opponent pays the granted life and
// the targeting spell stays on the stack.
func TestGrantedWardPayLifeTriggersAndPays(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPayLifeWardGrantSource(g, game.Player2)
	beneficiary := addVanillaCreature(g, game.Player2)
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1
	startLife := g.Players[game.Player1].Life

	if !hasKeyword(g, beneficiary, game.Ward) {
		t.Fatal("granted Ward keyword not present on controlled creature")
	}
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(beneficiary.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("granted ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != startLife-2 {
		t.Fatalf("life = %d, want %d (granted ward pay-2-life)", got, startLife-2)
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want targeting spell still on stack", g.Stack.Size())
	}
}

// The "other creatures you control" grant excludes the granting source itself,
// and the grant is removed when the source leaves the battlefield.
func TestGrantedWardExcludesSourceAndFollowsSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addPayLifeWardGrantSource(g, game.Player2)
	beneficiary := addVanillaCreature(g, game.Player2)

	if hasKeyword(g, source, game.Ward) {
		t.Fatal("granting source should be excluded from its own Ward grant")
	}
	if !hasKeyword(g, beneficiary, game.Ward) {
		t.Fatal("controlled creature should receive the Ward grant")
	}

	if _, ok := removePermanentFromBattlefield(g, source.ObjectID); !ok {
		t.Fatal("failed to remove grant source from battlefield")
	}
	if hasKeyword(g, beneficiary, game.Ward) {
		t.Fatal("Ward grant should end when the granting source leaves")
	}
}

// A creature warded only through a controlled grant does not trigger when its
// own controller targets it: Ward triggers only for opponents.
func TestGrantedWardDoesNotTriggerForControllersOwnSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPayLifeWardGrantSource(g, game.Player2)
	beneficiary := addVanillaCreature(g, game.Player2)
	spellID := addCardToHand(g, game.Player2, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player2

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, []game.Target{game.PermanentTarget(beneficiary.ObjectID)}, 0, nil)) {
		t.Fatal("self-targeting spell cast failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("granted ward triggered for controller's own spell")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want only targeting spell", g.Stack.Size())
	}
}

// Ward—Pay N life is a payment, not a life loss: a player who cannot pay the
// full life cost (CR 119.4) does not pay, and the targeting spell is countered.
func TestWardPayLifeInsufficientLifeCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addCompositeWardPermanent(g, game.Player2, nil, []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 3}})
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Players[game.Player1].Life = 2
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 2 {
		t.Fatalf("life = %d, want 2 (insufficient life is never partially paid)", got)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want ward to counter targeting spell", g.Stack.Size())
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("countered spell did not move to graveyard")
	}
}
