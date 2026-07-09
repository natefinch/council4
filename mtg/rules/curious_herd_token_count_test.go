package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCuriousHerdCreatesTokenPerTargetArtifact resolves Curious Herd's "You
// create X 3/3 green Beast creature tokens, where X is the number of artifacts
// that player controls" as a CreateToken whose dynamic count is a
// player-controlled artifact group anchored to the single player target. The
// caster must create exactly one token per artifact the targeted opponent
// controls — the caster's own artifacts do not contribute — and every token
// enters under the caster's control, not the target's.
func TestCuriousHerdCreatesTokenPerTargetArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addCombatPermanent(g, game.Player2, artifactPermanentDef())
	addCombatPermanent(g, game.Player2, artifactPermanentDef())
	addCombatPermanent(g, game.Player2, artifactPermanentDef())
	// A creature the target controls and an artifact the caster controls must
	// both be excluded: only the target's artifacts are counted.
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Target Creature",
		Types: []types.Card{types.Creature},
	}})
	addCombatPermanent(g, game.Player1, artifactPermanentDef())

	beastToken := &game.CardDef{CardFace: game.CardFace{
		Name:      "Beast",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Beast},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}}
	addEffectSpellToStack(g, game.Player1, game.CreateToken{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountCountSelector,
			Multiplier: 1,
			Group: game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{
				RequiredTypes: []types.Card{types.Artifact},
			}),
		}),
		Source: game.TokenDef(beastToken),
	}, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStack(g, &TurnLog{})

	casterTokens := 0
	targetTokens := 0
	for _, permanent := range g.Battlefield {
		if !permanent.Token {
			continue
		}
		switch permanent.Controller {
		case game.Player1:
			casterTokens++
		case game.Player2:
			targetTokens++
		default:
		}
	}
	if casterTokens != 3 {
		t.Fatalf("caster Beast tokens = %d, want 3 (one per target artifact)", casterTokens)
	}
	if targetTokens != 0 {
		t.Fatalf("target tokens = %d, want 0 (tokens enter under the caster)", targetTokens)
	}
}
