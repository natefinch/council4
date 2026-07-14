package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// applyPlayerRuleEffect resolves a single until-end-of-turn player rule effect
// (e.g. player hexproof or player shroud) controlled by controller and affecting
// the PlayerYou relation, the shape Dawn's Truce's player-scoped grant produces.
func applyPlayerRuleEffect(engine *Engine, g *game.Game, controller game.PlayerID, kind game.RuleEffectKind) {
	obj := &game.StackObject{Controller: controller}
	resolveInstruction(engine, g, obj, game.ApplyRule{
		RuleEffects: []game.RuleEffect{
			{Kind: kind, AffectedPlayer: game.PlayerYou},
		},
		Duration: game.DurationUntilEndOfTurn,
	}, nil)
}

func playerIsCandidate(g *game.Game, sourceController game.PlayerID, player game.PlayerID) bool {
	spec := &game.TargetSpec{MinTargets: 1, MaxTargets: 1, Constraint: "player"}
	for _, candidate := range targetCandidatesForSpec(g, sourceController, nil, 0, game.Event{}, spec) {
		if candidate.Kind == game.TargetPlayer && candidate.PlayerID == player {
			return true
		}
	}
	return false
}

// TestPlayerHexproofBlocksOpponentTargeting proves the player hexproof rule
// effect makes its controller untargetable by an opponent's spells and abilities
// (CR 702.11d) while still allowing that player to target themselves. Dawn's
// Truce grants this to its caster until end of turn.
func TestPlayerHexproofBlocksOpponentTargeting(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	applyPlayerRuleEffect(engine, g, game.Player1, game.RuleEffectPlayerHexproof)

	if playerIsCandidate(g, game.Player2, game.Player1) {
		t.Fatal("opponent could target a player with hexproof")
	}
	if !playerIsCandidate(g, game.Player1, game.Player1) {
		t.Fatal("hexproof must still allow the player to target themselves")
	}
	if !playerIsCandidate(g, game.Player1, game.Player2) {
		t.Fatal("hexproof on Player1 must not protect Player2")
	}
}

// TestPlayerShroudBlocksAllTargeting proves the player shroud rule effect makes
// its controller untargetable by any spell or ability, including their own
// (CR 702.18b), unlike hexproof.
func TestPlayerShroudBlocksAllTargeting(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	applyPlayerRuleEffect(engine, g, game.Player1, game.RuleEffectPlayerShroud)

	if playerIsCandidate(g, game.Player2, game.Player1) {
		t.Fatal("opponent could target a player with shroud")
	}
	if playerIsCandidate(g, game.Player1, game.Player1) {
		t.Fatal("shroud must block the player from targeting themselves")
	}
}

// TestPlayerHexproofExpiresWithDuration proves the until-end-of-turn player
// hexproof no longer blocks targeting once its duration ends and the rule effect
// is cleared.
func TestPlayerHexproofExpiresWithDuration(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	applyPlayerRuleEffect(engine, g, game.Player1, game.RuleEffectPlayerHexproof)
	if playerIsCandidate(g, game.Player2, game.Player1) {
		t.Fatal("opponent could target a player with active hexproof")
	}

	expireRuleEffects(g)
	if !playerIsCandidate(g, game.Player2, game.Player1) {
		t.Fatal("hexproof should not persist past its until-end-of-turn duration")
	}
}

// TestGiftDeliveryDrawsForPromisedRecipient proves Dawn's Truce's gift delivery
// draws a card for the promised opponent (via GiftRecipientReference) and not
// for the caster, resolving before the spell's other effects (CR 702.171).
func TestGiftDeliveryDrawsForPromisedRecipient(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Gifted"}})

	obj := &game.StackObject{
		Kind:          game.StackSpell,
		Controller:    game.Player1,
		GiftPromised:  true,
		GiftRecipient: game.Player2,
	}
	resolveInstruction(engine, g, obj, game.Draw{
		Amount: game.Fixed(1),
		Player: game.GiftRecipientReference(),
	}, nil)

	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("recipient hand size = %d, want 1 (opponent draws the gift)", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("caster hand size = %d, want 0 (the gift goes to the opponent)", got)
	}
}
