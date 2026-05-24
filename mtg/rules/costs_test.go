package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestCanPayCostWithUntappedForest(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if !canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = false, want true")
	}
}

func TestCanPayCostRejectsWrongBasicLandColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Mountain")
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = true with wrong basic land color, want false")
	}
}

func TestCanPayGenericCostWithAnyBasicLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Mountain")
	cost := mana.Cost{mana.GenericMana(1)}

	if !canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = false, want true")
	}
}

func TestPayCostTapsLandUsed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if !payCost(g, game.Player1, &cost) {
		t.Fatal("payCost() = false, want true")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped")
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("mana pool total = %d, want 0", g.Players[game.Player1].ManaPool.Total())
	}
}

func TestTappedLandCannotPayAgain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if !payCost(g, game.Player1, &cost) {
		t.Fatal("first payCost() = false, want true")
	}
	if canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = true after land tapped, want false")
	}
}

func TestPayCostFailureDoesNotMutate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.ColoredMana(mana.Green), mana.ColoredMana(mana.Green)}

	if payCost(g, game.Player1, &cost) {
		t.Fatal("payCost() = true with insufficient mana, want false")
	}
	if forest.Tapped {
		t.Fatal("forest was tapped by failed payment")
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("mana pool total = %d, want 0", g.Players[game.Player1].ManaPool.Total())
	}
}

func TestPayCostUsesPoolBeforeTappingLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Players[game.Player1].ManaPool.Add(mana.Green, 1)
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if !payCost(g, game.Player1, &cost) {
		t.Fatal("payCost() = false, want true")
	}
	if forest.Tapped {
		t.Fatal("forest was tapped even though pool could pay")
	}
}

func TestManaPoolsEmptyAfterMainPhase(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.Add(mana.Green, 1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	engine := NewEngine(nil)

	engine.runMainPhase(g, [game.NumPlayers]PlayerAgent{}, game.PhasePrecombatMain, &TurnLog{})

	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("mana pool total = %d, want 0", g.Players[game.Player1].ManaPool.Total())
	}
}

func TestUnsupportedCostCannotBePaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.VariableMana()}

	if canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = true for unsupported X cost, want false")
	}
}

func addBasicLandPermanent(g *game.Game, controller game.PlayerID, subtype string) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{
			Name:     subtype,
			Types:    []game.CardType{game.TypeLand},
			Subtypes: []string{subtype},
		},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       id.ID(g.IDGen.Next()),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}
