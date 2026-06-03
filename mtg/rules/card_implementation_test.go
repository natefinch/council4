package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

type fakeDrawBurnImplementation struct{}

func (fakeDrawBurnImplementation) ResolveSpell(ctx *CardContext, obj *game.StackObject, card *game.CardInstance) {
	ctx.DrawCards(obj.Controller, 2)
	if target, ok := ctx.TargetPlayer(obj, 0); ok {
		ctx.DealPlayerDamageFromStack(obj, target, 3)
	}
}

type fakePermanentDamageImplementation struct{}

func (fakePermanentDamageImplementation) ResolveSpell(ctx *CardContext, obj *game.StackObject, card *game.CardInstance) {
	if target, ok := ctx.TargetPermanentID(obj, 0); ok {
		ctx.DealPermanentDamageFromStack(obj, target, 3)
	}
}

func TestCardImplementationHandlesSpellResolutionThroughContext(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	engine.RegisterCardImplementation("test/draw-burn", fakeDrawBurnImplementation{})
	firstDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	secondDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	sourceID := addImplementationSpellToStack(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Hand-Written Draw Burn",
		Types:            []types.Card{types.Sorcery},
		ImplementationID: "test/draw-burn",
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "target player"},
				},
			},
		}},
	}, []game.Target{game.PlayerTarget(game.Player2)})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if !g.Players[game.Player1].Hand.Contains(firstDraw) || !g.Players[game.Player1].Hand.Contains(secondDraw) {
		t.Fatal("custom implementation did not draw both cards")
	}
	if len(log.Draws) != 2 {
		t.Fatalf("draw logs = %d, want 2", len(log.Draws))
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("player 2 life = %d, want 37", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("custom-implemented spell did not move to graveyard after resolution")
	}
	assertEvent(t, g.Events, game.EventCardDrawn, func(event game.GameEvent) bool {
		return event.Player == game.Player1 && event.CardID == firstDraw
	})
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.GameEvent) bool {
		return event.SourceID == sourceID &&
			event.Player == game.Player2 &&
			event.Amount == 3 &&
			event.DamageRecipient == game.DamageRecipientPlayer
	})
}

func TestCardImplementationUsesNormalDamagePreventionHelpers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	engine.RegisterCardImplementation("test/permanent-damage", fakePermanentDamageImplementation{})
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)
	sourceID := addImplementationSpellToStack(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Hand-Written Permanent Burn",
		Types:            []types.Card{types.Sorcery},
		ImplementationID: "test/permanent-damage",
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "target creature"},
				},
			},
		}},
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", target.MarkedDamage)
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.GameEvent) bool {
		return event.SourceID == sourceID &&
			event.PermanentID == target.ObjectID &&
			event.Amount == 3
	})
	assertNoEvent(t, g.Events, game.EventDamageDealt, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestUnregisteredCardImplementationPanics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addImplementationSpellToStack(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Missing Implementation",
		Types:            []types.Card{types.Sorcery},
		ImplementationID: "test/missing",
		Abilities:        []game.AbilityDef{{Kind: game.SpellAbility}}},
	}, nil)

	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("resolveTopOfStack did not panic for an unregistered implementation")
		}
	}()

	engine.resolveTopOfStack(g, &TurnLog{})
}

func TestDuplicateCardImplementationRegistrationPanics(t *testing.T) {
	engine := NewEngine(nil)
	engine.RegisterCardImplementation("test/duplicate", fakeDrawBurnImplementation{})

	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("duplicate RegisterCardImplementation did not panic")
		}
	}()

	engine.RegisterCardImplementation("test/duplicate", fakeDrawBurnImplementation{})
}

func addImplementationSpellToStack(g *game.Game, controller game.PlayerID, def *game.CardDef, targets []game.Target) id.ID {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID:    sourceID,
		Def:   def,
		Owner: controller,
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: controller,
		Targets:    targets,
	})
	return sourceID
}
