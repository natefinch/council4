package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
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
	movePermanentToZone(g, target, game.ZoneGraveyard)
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
	return &game.CardDef{
		Name:  name,
		Types: []game.CardType{game.TypeInstant},
		Abilities: []game.AbilityDef{{
			Kind:    game.SpellAbility,
			Effects: []game.Effect{{Type: game.EffectGainLife, Amount: 1, TargetIndex: -1}},
		}},
	}
}

func stormGainLifeInstant() *game.CardDef {
	card := simpleGainLifeInstant("Storm Spell")
	card.Abilities = append([]game.AbilityDef{{Kind: game.StaticAbility, Keywords: []game.Keyword{game.Storm}}}, card.Abilities...)
	return card
}

func stormTargetCreatureInstant() *game.CardDef {
	return &game.CardDef{
		Name:  "Targeted Storm Spell",
		Types: []game.CardType{game.TypeInstant},
		Abilities: []game.AbilityDef{
			{Kind: game.StaticAbility, Keywords: []game.Keyword{game.Storm}},
			{
				Kind:    game.SpellAbility,
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
				Effects: []game.Effect{{Type: game.EffectDamage, Amount: 1, TargetIndex: 0}},
			},
		},
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
