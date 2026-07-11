package rules

import (
	"testing"

	cardsd "github.com/natefinch/council4/mtg/cards/d"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestDevastatingSummonsSacrificeXLandsAnnouncesXForTokens proves the
// spell-level "As an additional cost to cast this spell, sacrifice X lands"
// variable cost: the announced X is bound from the sacrifice count, exactly that
// many lands are sacrificed, and the two Elemental tokens the spell creates are
// each X/X (their power and toughness read the same announced X).
//
// The {R} mana comes from a non-land source here so the mana payment and the
// sacrifice-X-lands cost do not compete for the same lands: reconciling
// land-based mana with a same-turn sacrifice of lands is a pre-existing
// payment-planner limitation (a fixed "sacrifice a land" cost paired with land
// mana fails the same way) that is orthogonal to this variable-X mechanic.
func TestDevastatingSummonsSacrificeXLandsAnnouncesXForTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, cardsd.DevastatingSummons())
	addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Red Rock", Types: []types.Card{types.Artifact}}}, mana.R, 1)
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	setSorcerySpeedTurn(g, game.Player1)

	landsBefore := countLands(g, game.Player1)

	// X is bounded by the sacrificeable lands: only three Forests are present.
	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 4, nil)) {
		t.Fatal("X=4 cast was legal with only three sacrificeable lands")
	}

	act := action.CastSpell(spellID, nil, 3, nil)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast Devastating Summons with X=3) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.XValue != 3 || len(obj.AdditionalCostsPaid) != 1 {
		t.Fatalf("stack object = %+v, want X=3 with one additional cost paid", obj)
	}
	if got := landsBefore - countLands(g, game.Player1); got != 3 {
		t.Fatalf("lands sacrificed = %d, want 3", got)
	}

	engine.resolveTopOfStack(g, nil)

	tokens := elementalTokens(g, game.Player1)
	if len(tokens) != 2 {
		t.Fatalf("Elemental tokens created = %d, want 2", len(tokens))
	}
	for _, token := range tokens {
		if got := effectivePower(g, token); got != 3 {
			t.Fatalf("Elemental token power = %d, want 3 (X=3)", got)
		}
		toughness, ok := effectiveToughness(g, token)
		if !ok || toughness != 3 {
			t.Fatalf("Elemental token toughness = %d (ok=%v), want 3 (X=3)", toughness, ok)
		}
	}
}

func countLands(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == controller && permanentHasType(g, permanent, types.Land) {
			count++
		}
	}
	return count
}

func elementalTokens(g *game.Game, controller game.PlayerID) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Controller == controller && permanent.Token &&
			permanentHasSubtype(g, permanent, types.Elemental) {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}
