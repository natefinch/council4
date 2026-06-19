package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// addManaLandPermanent adds a battlefield land controlled by controller whose
// only ability is "{T}: Add {color}.", so its producible color is exactly color.
func addManaLandPermanent(g *game.Game, controller game.PlayerID, name string, color mana.Color) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(color)},
	}})
}

// TestLandsProduceManaUnionsOpponentLandColors verifies that the opponent-scoped
// "any color that a land an opponent controls could produce" choice (Exotic
// Orchard, Fellwar Stone) offers exactly the colors the opponent's lands could
// produce, in WUBRG order, and offers nothing when the opponent controls no
// producing land.
func TestLandsProduceManaUnionsOpponentLandColors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	choice := &game.ResolutionChoice{
		Kind:           game.ResolutionChoiceMana,
		ColorSource:    game.ResolutionChoiceColorSourceLandsProduce,
		PlayerRelation: game.PlayerOpponent,
	}

	// No opponent lands: the choice is empty.
	if got := landsProduceMana(g, game.Player1, choice); len(got) != 0 {
		t.Fatalf("lands-produce with no opponent lands = %v, want empty", got)
	}

	// The controller's own lands must not feed an opponent-scoped choice.
	addManaLandPermanent(g, game.Player1, "Mountain", mana.R)
	if got := landsProduceMana(g, game.Player1, choice); len(got) != 0 {
		t.Fatalf("opponent-scoped choice counted controller lands: %v", got)
	}

	// Opponent controls Forest + Island: the choice offers G and U (WUBRG order).
	addManaLandPermanent(g, game.Player2, "Forest", mana.G)
	addManaLandPermanent(g, game.Player2, "Island", mana.U)
	got := landsProduceMana(g, game.Player1, choice)
	if want := []mana.Color{mana.U, mana.G}; !slices.Equal(got, want) {
		t.Fatalf("lands-produce colors = %v, want %v", got, want)
	}
}

// TestLandsProduceManaYouScopeReadsOwnLands verifies the you-scoped wording
// (Reflecting Pool, Harvester Druid) reads the controller's own lands.
func TestLandsProduceManaYouScopeReadsOwnLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	choice := &game.ResolutionChoice{
		Kind:           game.ResolutionChoiceMana,
		ColorSource:    game.ResolutionChoiceColorSourceLandsProduce,
		PlayerRelation: game.PlayerYou,
	}
	addManaLandPermanent(g, game.Player1, "Plains", mana.W)
	addManaLandPermanent(g, game.Player2, "Swamp", mana.B)
	got := landsProduceMana(g, game.Player1, choice)
	if want := []mana.Color{mana.W}; !slices.Equal(got, want) {
		t.Fatalf("you-scoped lands-produce colors = %v, want %v", got, want)
	}
}

// TestLandsProduceManaAnyTypeIncludesColorless verifies the "any type" wording
// (Reflecting Pool) additionally offers colorless when a matching land could
// produce it, while the "any color" wording never does.
func TestLandsProduceManaAnyTypeIncludesColorless(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addManaLandPermanent(g, game.Player1, "Forest", mana.G)
	addManaLandPermanent(g, game.Player1, "Wastes", mana.C)

	anyColor := &game.ResolutionChoice{
		Kind:           game.ResolutionChoiceMana,
		ColorSource:    game.ResolutionChoiceColorSourceLandsProduce,
		PlayerRelation: game.PlayerYou,
	}
	if got := landsProduceMana(g, game.Player1, anyColor); !slices.Equal(got, []mana.Color{mana.G}) {
		t.Fatalf("any-color lands-produce = %v, want [G] (no colorless)", got)
	}

	anyType := &game.ResolutionChoice{
		Kind:             game.ResolutionChoiceMana,
		ColorSource:      game.ResolutionChoiceColorSourceLandsProduce,
		PlayerRelation:   game.PlayerYou,
		IncludeColorless: true,
	}
	if got := landsProduceMana(g, game.Player1, anyType); !slices.Equal(got, []mana.Color{mana.G, mana.C}) {
		t.Fatalf("any-type lands-produce = %v, want [G C]", got)
	}
}

// TestLandsProduceManaAbilityActivationGating verifies the Exotic Orchard ability
// is unactivatable when the opponent controls no producing land and activatable
// once the opponent does.
func TestLandsProduceManaAbilityActivationGating(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	orchard := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:          "Exotic Orchard",
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaLandsProduceAbility(game.PlayerOpponent, false)},
	}})
	card, ok := permanentCardDef(g, orchard)
	if !ok {
		t.Fatal("permanent card definition not found")
	}

	if canActivateManaAbility(g, game.Player1, orchard, &card.ManaAbilities[0], 0) {
		t.Fatal("canActivateManaAbility() = true, want false without opponent lands")
	}

	addManaLandPermanent(g, game.Player2, "Island", mana.U)
	if !canActivateManaAbility(g, game.Player1, orchard, &card.ManaAbilities[0], 0) {
		t.Fatal("canActivateManaAbility() = false, want true with an opponent land")
	}
}

// TestLandsProduceManaAbilityActivationAddsChosenColor exercises the full Exotic
// Orchard activation: with the opponent controlling a Forest and an Island, the
// choice offers G and U; choosing the second option adds one blue mana.
func TestLandsProduceManaAbilityActivationAddsChosenColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setSorcerySpeedTurn(g, game.Player1)
	orchard := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:          "Exotic Orchard",
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaLandsProduceAbility(game.PlayerOpponent, false)},
	}})
	addManaLandPermanent(g, game.Player2, "Forest", mana.G)
	addManaLandPermanent(g, game.Player2, "Island", mana.U)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(orchard.ObjectID, 0, nil, 0), agents, &log) {
		t.Fatal("applyActionWithChoices(Exotic Orchard) = false, want true")
	}
	if !orchard.Tapped {
		t.Fatal("Exotic Orchard was not tapped")
	}
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want one mana choice", log.Choices)
	}
	options := log.Choices[0].Request.Options
	if len(options) != 2 || options[0].Label != "U" || options[1].Label != "G" {
		t.Fatalf("choice options = %+v, want [U G]", options)
	}
	// The agent selected option index 1 ("G"), so one green mana is added.
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana = %d, want 1", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 0 {
		t.Fatalf("blue mana = %d, want 0 (chose green)", got)
	}
}

// TestLandsProduceManaAvoidsLoopBetweenTwoOrchards verifies the loop-avoidance
// ruling: when both players control only an Exotic Orchard and no other land,
// neither orchard can produce mana, so the choice is empty.
func TestLandsProduceManaAvoidsLoopBetweenTwoOrchards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	orchardDef := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name:          "Exotic Orchard",
			Types:         []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{game.TapManaLandsProduceAbility(game.PlayerOpponent, false)},
		}}
	}
	addCombatPermanent(g, game.Player1, orchardDef())
	addCombatPermanent(g, game.Player2, orchardDef())
	choice := &game.ResolutionChoice{
		Kind:           game.ResolutionChoiceMana,
		ColorSource:    game.ResolutionChoiceColorSourceLandsProduce,
		PlayerRelation: game.PlayerOpponent,
	}
	if got := landsProduceMana(g, game.Player1, choice); len(got) != 0 {
		t.Fatalf("two opposing orchards offered %v, want empty (loop avoidance)", got)
	}
}
