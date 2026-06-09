package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestStormCreatesCopiesForPriorSpellsThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
	stormID := addCardToHand(g, game.Player1, stormGainLifeInstant())
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil)) {
		t.Fatal("first spell cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.applyAction(g, game.Player1, action.CastSpell(stormID, nil, 0, nil)) {
		t.Fatal("storm spell cast failed")
	}

	if g.Stack.Size() != 2 {
		t.Fatalf("stack size = %d, want storm original plus one copy", g.Stack.Size())
	}
	top, _ := g.Stack.Peek()
	if !top.Copy || top.SourceID != stormID {
		t.Fatalf("top stack object = %+v, want storm copy", top)
	}
	if spellCastEventCount(g) != 2 {
		t.Fatalf("spell cast events = %d, want copies not to emit cast events", spellCastEventCount(g))
	}
}

func TestStormCopyResolvesWithoutMovingSourceCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
	stormID := addCardToHand(g, game.Player1, stormGainLifeInstant())
	g.Turn.PriorityPlayer = game.Player1

	engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil))
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyAction(g, game.Player1, action.CastSpell(stormID, nil, 0, nil))
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(stormID) {
		t.Fatal("storm copy moved source card to graveyard")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want original storm spell remaining", g.Stack.Size())
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(stormID) {
		t.Fatal("storm original did not move to graveyard")
	}
}

func TestStormCopyCounteredByRulesDoesNotMoveSourceCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
	stormID := addCardToHand(g, game.Player1, stormTargetCreatureInstant())
	target := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.PriorityPlayer = game.Player1

	engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil))
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyAction(g, game.Player1, action.CastSpell(stormID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil))
	movePermanentToZone(g, target, zone.Graveyard)
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(stormID) {
		t.Fatal("countered storm copy moved source card to graveyard")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want original storm spell remaining", g.Stack.Size())
	}
}

func TestCounteringStormCopyDoesNotMoveSourceCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	stormID := addCardToHand(g, game.Player1, stormGainLifeInstant())
	g.Players[game.Player1].Hand.Remove(stormID)
	copyObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   stormID,
		Controller: game.Player1,
		Copy:       true,
	}
	g.Stack.Push(copyObj)

	if !counterStackObject(g, copyObj.ID) {
		t.Fatal("counterStackObject(copy) = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(stormID) {
		t.Fatal("countered storm copy moved source card to graveyard")
	}
}

func simpleGainLifeInstant(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name,
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			},
		}.Ability())},
	}
}

func stormGainLifeInstant() *game.CardDef {
	card := simpleGainLifeInstant("Storm Spell")
	card.StaticAbilities = append([]game.StaticAbility{game.StormStaticBody}, card.StaticAbilities...)
	return card
}

func stormTargetCreatureInstant() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Targeted Storm Spell",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{game.StormStaticBody}},
	}
}

func spellCastEventCount(g *game.Game) int {
	count := 0
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast {
			count++
		}
	}
	return count
}
