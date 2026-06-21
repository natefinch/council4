package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func treasureTokenCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Treasure",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Treasure},
	}}
}

func treasureAddendReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Xorn",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.TokenCreationReplacementFiltered(
				"If you would create one or more Treasure tokens, instead create those tokens plus an additional Treasure token.",
				&game.TokenCreationReplacementSpec{
					Multiplier: 1,
					Addend:     1,
					Subtypes:   []types.Sub{types.Treasure},
					Filter:     game.TriggerControllerYou,
				},
			),
		},
	}}
}

func anyControllerTokenDoublingCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Primal Vigor",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.TokenCreationReplacementFiltered(
				"If one or more tokens would be created, twice that many of those tokens are created instead.",
				&game.TokenCreationReplacementSpec{
					Multiplier: 2,
					Filter:     game.TriggerControllerAny,
				},
			),
		},
	}}
}

func TestTokenAddendCreatesExtraSameTypeToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, treasureAddendReplacementCardDef())

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, treasureTokenCardDef(), 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Treasure"); got != 2 {
		t.Fatalf("created Treasure tokens = %d, want 2", got)
	}
}

func TestTokenAddendIgnoresOtherSubtypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, treasureAddendReplacementCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 1 {
		t.Fatalf("created non-Treasure tokens = %d, want 1", got)
	}
}

func TestTokenCreationAnyControllerScopeAffectsOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyControllerTokenDoublingCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Saproling", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player2, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices(Player2) = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Saproling"); got != 2 {
		t.Fatalf("opponent-created tokens under any-player doubler = %d, want 2", got)
	}
}

func anyCreatureCounterDoublingCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Primal Vigor",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.CounterPlacementReplacement(
				"If one or more +1/+1 counters would be put on a creature, twice that many +1/+1 counters are put on that creature instead.",
				2,
				0,
				counter.PlusOnePlusOne,
				game.TriggerControllerAny,
			),
		},
	}}
}

func TestCounterPlacementAnyCreatureScopeDoublesOpponentCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCreatureCounterDoublingCardDef())
	creature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanentControlledBy(g, game.Player2, creature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanentControlledBy(opponent) = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("opponent creature +1/+1 counters under any-creature doubler = %d, want 2", got)
	}
}

func crossTypeAddendCardDef() *game.CardDef {
	food := &game.CardDef{CardFace: game.CardFace{
		Name:     "Food",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Food},
	}}
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Tippy-Toe, Terrific Partner",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.TokenCreationReplacementFiltered(
				"If you would create one or more tokens, instead create those tokens plus an additional Food token.",
				&game.TokenCreationReplacementSpec{
					Multiplier: 1,
					Addend:     1,
					Filter:     game.TriggerControllerYou,
					AddendDef:  food,
				},
			),
		},
	}}
}

func TestTokenAddendCreatesDifferentPredefinedTokenType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, crossTypeAddendCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Squirrel", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Squirrel"); got != 1 {
		t.Fatalf("created Squirrel tokens = %d, want 1 (addend must not duplicate the matched token)", got)
	}
	if got := countTokenPermanentsNamed(g, "Food"); got != 1 {
		t.Fatalf("created Food tokens = %d, want 1 (cross-type addend)", got)
	}
}
