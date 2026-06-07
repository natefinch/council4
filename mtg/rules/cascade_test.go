package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCascadeExilesUntilLowerManaNonlandAndCastsIt(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cascadeID := addCardToHand(g, game.Player1, cascadeSpell(5))
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 2))
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island", Types: []types.Card{types.Land}}})
	skippedID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Too Expensive", 7))
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(cascadeID, nil, 0, nil)) {
		t.Fatal("cascade spell cast failed")
	}

	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != hitID {
		t.Fatalf("stack top = %+v, want cascaded spell %v", obj, hitID)
	}
	if g.Players[game.Player1].Exile.Contains(hitID) {
		t.Fatal("cascaded spell remained in exile")
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{skippedID, landID}) {
		t.Fatalf("library = %+v, want skipped cards on bottom in random order containing [%v %v]", got, skippedID, landID)
	}
	if spellCastEventCount(g) != 2 {
		t.Fatalf("spell cast events = %d, want cascade spell and cascaded spell", spellCastEventCount(g))
	}
}

func TestCascadeBottomsCardsWhenNoEligibleCardExists(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cascadeID := addCardToHand(g, game.Player1, cascadeSpell(3))
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island", Types: []types.Card{types.Land}}})
	skippedID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Too Expensive", 7))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(cascadeID, nil, 0, nil)) {
		t.Fatal("cascade spell cast failed")
	}

	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want only original cascade spell", g.Stack.Size())
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{skippedID, landID}) {
		t.Fatalf("library = %+v, want revealed cards returned in random order containing [%v %v]", got, skippedID, landID)
	}
	if spellCastEventCount(g) != 1 {
		t.Fatalf("spell cast events = %d, want only cascade spell", spellCastEventCount(g))
	}
}

func TestCascadeChainsFromCascadedSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cascadeID := addCardToHand(g, game.Player1, cascadeSpell(6))
	secondHitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Second Hit", 1))
	firstHitID := addCardToLibrary(g, game.Player1, cascadeSpell(3))
	g.Players[game.Player1].ManaPool.Add(mana.C, 6)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(cascadeID, nil, 0, nil)) {
		t.Fatal("cascade spell cast failed")
	}

	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != secondHitID {
		t.Fatalf("stack top = %+v, want chained cascade hit %v", top, secondHitID)
	}
	if g.Stack.Size() != 3 {
		t.Fatalf("stack size = %d, want original plus first and second cascade hits", g.Stack.Size())
	}
	if !stackContainsSource(g, firstHitID) {
		t.Fatalf("stack does not contain first cascaded spell %v", firstHitID)
	}
}

func TestDiscoverExilesUntilEligibleCardAndCastsIt(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.DiscoverCards{Amount: game.Fixed(3)}, nil)
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Discover Hit", 3))
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island", Types: []types.Card{types.Land}}})
	skippedID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Too Expensive", 4))

	engine.resolveTopOfStack(g, &TurnLog{})

	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != hitID {
		t.Fatalf("stack top = %+v, want discovered spell %v", obj, hitID)
	}
	if g.Players[game.Player1].Exile.Contains(hitID) {
		t.Fatal("discovered spell remained in exile")
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{skippedID, landID}) {
		t.Fatalf("library = %+v, want skipped cards on bottom in random order containing [%v %v]", got, skippedID, landID)
	}
}

func TestDiscoverDeclinePutsFoundCardIntoHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.DiscoverCards{Amount: game.Fixed(2)}, nil)
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Declined Hit", 2))
	skippedID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Too Expensive", 5))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no discovered spell cast", g.Stack.Size())
	}
	if !g.Players[game.Player1].Hand.Contains(hitID) {
		t.Fatal("declined discovered card was not put into hand")
	}
	if got := g.Players[game.Player1].Library.All(); len(got) != 1 || got[0] != skippedID {
		t.Fatalf("library = %+v, want skipped card on bottom [%v]", got, skippedID)
	}
}

func cascadeSpell(manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Cascade Spell",
		ManaCost:        opt.Val(cost.Mana{cost.O(manaValue)}),
		Types:           []types.Card{types.Instant},
		SpellAbility:    opt.Val(game.ModalAbilityContent{}),
		StaticAbilities: []game.StaticAbilityBody{game.CascadeStaticBody}},
	}
}

func simpleGainLifeInstantWithManaValue(name string, manaValue int) *game.CardDef {
	card := simpleGainLifeInstant(name)
	card.ManaCost = opt.Val(cost.Mana{cost.O(manaValue)})
	return card
}

func stackContainsSource(g *game.Game, sourceID id.ID) bool {
	for _, obj := range g.Stack.Objects() {
		if obj.SourceID == sourceID {
			return true
		}

	}
	return false
}

func sameCardIDs(got, want []id.ID) bool {
	if len(got) != len(want) {
		return false
	}
	counts := make(map[id.ID]int, len(want))
	for _, cardID := range want {
		counts[cardID]++
	}
	for _, cardID := range got {
		if counts[cardID] == 0 {
			return false
		}
		counts[cardID]--
	}
	return true
}
