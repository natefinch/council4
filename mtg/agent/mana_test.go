package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func landCardDef(name string, produces mana.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(produces)},
	}}
}

func coloredCreatureDef(name string, power, toughness int, c color.Color) *game.CardDef {
	def := creatureCardDef(name, power, toughness)
	def.Colors = []color.Color{c}
	return def
}

// creatureWithCost is a creature with a generic mana cost so its mana value, and
// thus the mana the agent would spend casting it, is non-zero.
func creatureWithCost(name string, power, toughness, manaValue int) *game.CardDef {
	def := creatureCardDef(name, power, toughness)
	def.ManaCost = opt.Val(genericCost(manaValue))
	return def
}

func instantDef(name string, manaValue int, c color.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{c},
		ManaCost: opt.Val(genericCost(manaValue)),
	}}
}

// genericCost builds a mana cost of the given mana value as generic mana, enough
// for ManaValue() to report the cost without pinning a colour.
func genericCost(amount int) cost.Mana {
	if amount <= 0 {
		return cost.Mana{}
	}
	return cost.Mana{cost.O(amount)}
}

// TestColorScrewAvoidancePrefersMissingColorLand checks the agent plays the land
// that fixes a colour its hand needs but cannot yet produce, rather than adding
// a source of a colour it already has.
func TestColorScrewAvoidancePrefersMissingColorLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Hand wants green and blue; the battlefield already makes blue.
	addObservedHandCard(g, game.Player1, coloredCreatureDef("Green Beast", 3, 3, color.Green))
	addObservedHandCard(g, game.Player1, instantDef("Blue Trick", 1, color.Blue))
	addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))

	forestID := addObservedHandCard(g, game.Player1, landCardDef("Forest", mana.G))
	islandID := addObservedHandCard(g, game.Player1, landCardDef("Island", mana.U))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	forest := strategy.ScoreAction(obs, action.PlayLand(forestID))
	island := strategy.ScoreAction(obs, action.PlayLand(islandID))
	if forest <= island {
		t.Fatalf("forest score %v should beat already-covered island %v", forest, island)
	}
}

// TestColorFixPrefersDualOverSingleNeed checks a land fixing two missing needed
// colours outranks a land fixing only one.
func TestColorFixPrefersDualOverSingleNeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, coloredCreatureDef("Green Beast", 3, 3, color.Green))
	addObservedHandCard(g, game.Player1, instantDef("Blue Trick", 1, color.Blue))

	forestID := addObservedHandCard(g, game.Player1, landCardDef("Forest", mana.G))
	dual := &game.CardDef{CardFace: game.CardFace{
		Name:  "Tropical Island",
		Types: []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{
			game.TapManaAbility(mana.G),
			game.TapManaAbility(mana.U),
		},
	}}
	dualID := addObservedHandCard(g, game.Player1, dual)
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	if dualScore, forestScore := strategy.ScoreAction(obs, action.PlayLand(dualID)), strategy.ScoreAction(obs, action.PlayLand(forestID)); dualScore <= forestScore {
		t.Fatalf("dual score %v should beat single-need forest %v", dualScore, forestScore)
	}
}

// TestHoldUpKeepsManaForCounter checks that, holding a cheap instant, the agent
// prefers a creature that leaves enough mana open for the instant over a more
// expensive creature that would tap it too low.
func TestHoldUpKeepsManaForCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, instantDef("Counterspell", 2, color.Blue))
	cheapID := addObservedHandCard(g, game.Player1, creatureWithCost("Cheap Bear", 2, 2, 2))
	expensiveID := addObservedHandCard(g, game.Player1, creatureWithCost("Big Beast", 4, 4, 4))
	for range 5 {
		addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))
	}
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	cheap := strategy.ScoreAction(obs, action.CastSpell(cheapID, nil, 0, nil))
	expensive := strategy.ScoreAction(obs, action.CastSpell(expensiveID, nil, 0, nil))
	if cheap <= expensive {
		t.Fatalf("holding a counter, cheap creature %v should beat tap-out creature %v", cheap, expensive)
	}
}

// TestNoHoldUpWithoutInstant checks the hold-up penalty applies only when the
// agent actually holds a reactive instant: without one it deploys its biggest
// threat.
func TestNoHoldUpWithoutInstant(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cheapID := addObservedHandCard(g, game.Player1, creatureWithCost("Cheap Bear", 2, 2, 2))
	expensiveID := addObservedHandCard(g, game.Player1, creatureWithCost("Big Beast", 4, 4, 4))
	for range 5 {
		addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))
	}
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	cheap := strategy.ScoreAction(obs, action.CastSpell(cheapID, nil, 0, nil))
	expensive := strategy.ScoreAction(obs, action.CastSpell(expensiveID, nil, 0, nil))
	if expensive <= cheap {
		t.Fatalf("without a counter, big beast %v should beat cheap bear %v", expensive, cheap)
	}
}

// TestHoldUpSkippedWhenManaTooLow checks that when the agent cannot keep the
// instant up regardless of its play, it does not penalise developing.
func TestHoldUpSkippedWhenManaTooLow(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, instantDef("Counterspell", 2, color.Blue))
	creatureID := addObservedHandCard(g, game.Player1, creatureWithCost("Bear", 2, 2, 2))
	// Only one source: the agent can never hold up the 2-mana counter, so it
	// should just develop without a hold-up penalty.
	addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	withInstant := strategy.ScoreAction(obs, action.CastSpell(creatureID, nil, 0, nil))
	wantNoPenalty := scoreCastBase + 2*scoreCastPerMana + scoreCreature
	if withInstant != wantNoPenalty {
		t.Fatalf("score %v should equal un-penalised %v when hold-up is impossible", withInstant, wantNoPenalty)
	}
}
