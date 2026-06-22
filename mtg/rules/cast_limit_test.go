package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// castLimitPermanent gives controller a battlefield permanent whose continuous
// static ability caps how many spells the affected players may cast each turn
// (Rule of Law, Eidolon of Rhetoric; Moderation).
func castLimitPermanent(g *game.Game, controller game.PlayerID, affected game.PlayerRelation, limit int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Castcap",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectCastLimitPerTurn,
				AffectedPlayer:   affected,
				CastLimitPerTurn: limit,
			}},
		}},
	}})
}

// addFreeInstant gives playerID a no-cost instant that can be cast at any time.
func addFreeInstant(g *game.Game, playerID game.PlayerID, name string) id.ID {
	return addCardToHand(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Instant},
	}})
}

func TestCastLimitForbidsSecondSpellEachPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	castLimitPermanent(g, game.Player1, game.PlayerAny, 1)

	first := addFreeInstant(g, game.Player1, "First")
	second := addFreeInstant(g, game.Player1, "Second")

	if !engine.canCastSpell(g, game.Player1, first, nil, 0, nil) {
		t.Fatal("first spell should be castable before any spell is cast this turn")
	}
	castSpellTargeting(g, game.Player1)
	if engine.canCastSpell(g, game.Player1, second, nil, 0, nil) {
		t.Fatal("second spell should be forbidden once the per-turn cast limit is reached")
	}
}

func TestCastLimitOpponentScopeSparesController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// An opponent-scoped limit controlled by Player1 restricts Player2, not Player1.
	castLimitPermanent(g, game.Player1, game.PlayerOpponent, 1)

	controllerSpell := addFreeInstant(g, game.Player1, "Mine")

	castSpellTargeting(g, game.Player1)
	if !engine.canCastSpell(g, game.Player1, controllerSpell, nil, 0, nil) {
		t.Fatal("an opponent-only cast limit must never restrict the controller")
	}

	opponentSpell := &game.CardDef{CardFace: game.CardFace{Name: "X", Types: []types.Card{types.Instant}}}
	if spellCastLimitReached(g, game.Player2, opponentSpell) {
		t.Fatal("the opponent has cast nothing yet, so the limit is not reached")
	}
	castSpellTargeting(g, game.Player2)
	if !spellCastLimitReached(g, game.Player2, opponentSpell) {
		t.Fatal("an opponent-scoped cast limit should restrict the opponent's second spell")
	}
}

func TestCastLimitControllerScopeRestrictsControllerOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	castLimitPermanent(g, game.Player1, game.PlayerYou, 1)

	controllerSpell := addFreeInstant(g, game.Player1, "Mine")

	if !engine.canCastSpell(g, game.Player1, controllerSpell, nil, 0, nil) {
		t.Fatal("first controller spell should be castable")
	}
	castSpellTargeting(g, game.Player1)
	if engine.canCastSpell(g, game.Player1, controllerSpell, nil, 0, nil) {
		t.Fatal("second controller spell should be forbidden by a You-scoped cast limit")
	}

	opponentSpell := &game.CardDef{CardFace: game.CardFace{Name: "Theirs", Types: []types.Card{types.Instant}}}
	castSpellTargeting(g, game.Player2)
	if spellCastLimitReached(g, game.Player2, opponentSpell) {
		t.Fatal("a You-scoped cast limit must not restrict the opponent")
	}
}
