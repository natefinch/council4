package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func beastTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Beast",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Beast},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}}
}

func countTokens(g *game.Game) int {
	tokens := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens++
		}
	}
	return tokens
}

// TestCreateTokenDynamicCountZeroCreatesNoTokens resolves Curious Herd's "create
// X 3/3 Beasts, where X is the number of artifacts that player controls" against
// a target that controls no artifacts. The dynamic count resolves to 0, so the
// spell must create zero tokens — the resolved-to-0 dynamic amount is honored
// rather than floored up to a single "create a token" default. Before the fix
// this created one Beast; this case fails without it.
func TestCreateTokenDynamicCountZeroCreatesNoTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// The caster controls an artifact; the target controls none. The count is
	// scoped to the target, so it resolves to 0 regardless of the caster's board.
	addCombatPermanent(g, game.Player1, artifactPermanentDef())

	addEffectSpellToStack(g, game.Player1, game.CreateToken{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountCountSelector,
			Multiplier: 1,
			Group: game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{
				RequiredTypes: []types.Card{types.Artifact},
			}),
		}),
		Source: game.TokenDef(beastTokenDef()),
	}, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countTokens(g); got != 0 {
		t.Fatalf("tokens created = %d, want 0 (dynamic count resolved to 0)", got)
	}
}

// TestCreateTokenUnsetAmountCreatesOne resolves a plain "create a token" whose
// Amount is the zero/unset Quantity. That unset default must still create a
// single token.
func TestCreateTokenUnsetAmountCreatesOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addEffectSpellToStack(g, game.Player1, game.CreateToken{
		Source: game.TokenDef(beastTokenDef()),
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countTokens(g); got != 1 {
		t.Fatalf("tokens created = %d, want 1 (unset amount defaults to one)", got)
	}
}

// TestCreateTokenFixedAmountCreatesThatMany resolves an explicit fixed count so
// the honored amount is neither floored nor defaulted.
func TestCreateTokenFixedAmountCreatesThatMany(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addEffectSpellToStack(g, game.Player1, game.CreateToken{
		Amount: game.Fixed(2),
		Source: game.TokenDef(beastTokenDef()),
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countTokens(g); got != 2 {
		t.Fatalf("tokens created = %d, want 2 (explicit fixed amount)", got)
	}
}
