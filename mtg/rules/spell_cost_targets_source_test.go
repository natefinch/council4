package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// targetsSourceTaxPermanent gives playerID a battlefield permanent whose static
// ability taxes opponents' spells that target it ("Spells your opponents cast
// that target this creature cost {N} more to cast.", Boreal Elemental).
func targetsSourceTaxPermanent(g *game.Game, playerID game.PlayerID, increase int) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Boreal Elemental",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedPlayer: game.PlayerOpponent,
				CostModifier: game.CostModifier{
					Kind:            game.CostModifierSpell,
					TargetsSource:   true,
					GenericIncrease: increase,
				},
			}},
		}},
	}})
}

func spellGenericIncreaseForCaster(g *game.Game, caster game.PlayerID, card *game.CardDef, targets []game.Target) int {
	total := 0
	for _, modifier := range staticCostModifiersForContext(g, caster, card, zone.Hand, targets) {
		total += modifier.GenericIncrease
	}
	return total
}

func TestSpellCostModifierTargetsSourceTaxesOpponentSpellsTargetingSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := targetsSourceTaxPermanent(g, game.Player1, 2)
	card := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Bolt",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	}}

	targetingSource := []game.Target{game.PermanentTarget(source.ObjectID)}
	if got := spellGenericIncreaseForCaster(g, game.Player2, card, targetingSource); got != 2 {
		t.Fatalf("opponent spell targeting source increase = %d, want 2", got)
	}

	other := []game.Target{game.PermanentTarget(g.IDGen.Next())}
	if got := spellGenericIncreaseForCaster(g, game.Player2, card, other); got != 0 {
		t.Fatalf("opponent spell targeting another permanent increase = %d, want 0", got)
	}

	if got := spellGenericIncreaseForCaster(g, game.Player2, card, nil); got != 0 {
		t.Fatalf("opponent spell with no targets increase = %d, want 0", got)
	}

	// The controller is not an opponent, so the tax never applies to their own
	// spells even when they target the source.
	if got := spellGenericIncreaseForCaster(g, game.Player1, card, targetingSource); got != 0 {
		t.Fatalf("controller spell targeting source increase = %d, want 0", got)
	}
}

func TestSpellCostModifierMatchesTargets(t *testing.T) {
	sourceID := id.ID(7)
	modifier := game.CostModifier{Kind: game.CostModifierSpell, TargetsSource: true}

	if !spellCostModifierMatchesTargets(modifier, sourceID, []game.Target{game.PermanentTarget(sourceID)}) {
		t.Fatal("targets-source modifier rejected a spell targeting the source")
	}
	if spellCostModifierMatchesTargets(modifier, sourceID, []game.Target{game.PermanentTarget(sourceID + 1)}) {
		t.Fatal("targets-source modifier matched a spell targeting another permanent")
	}
	if spellCostModifierMatchesTargets(modifier, sourceID, nil) {
		t.Fatal("targets-source modifier matched a spell with no targets")
	}
	if spellCostModifierMatchesTargets(modifier, 0, []game.Target{game.PermanentTarget(sourceID)}) {
		t.Fatal("targets-source modifier matched with no source object id")
	}

	unfiltered := game.CostModifier{Kind: game.CostModifierSpell}
	if !spellCostModifierMatchesTargets(unfiltered, 0, nil) {
		t.Fatal("modifier without targets-source filter rejected a spell")
	}
}
