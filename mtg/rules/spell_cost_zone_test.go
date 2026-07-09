package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// graveyardSpellReductionEffect models the active rule effect cardgen emits for
// "Spells you cast from your graveyard cost {N} less to cast.": a controller's
// spell cost modifier scoped to spells cast from the graveyard.
func graveyardSpellReductionEffect(g *game.Game, reduction int) game.RuleEffect {
	return game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCostModifier,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CostModifier: game.CostModifier{
			Kind:             game.CostModifierSpell,
			SourceZone:       opt.Val(zone.Graveyard),
			GenericReduction: reduction,
		},
	}
}

func spellGenericReductionFromZone(g *game.Game, card *game.CardDef, sourceZone zone.Type) int {
	state := &rulesPaymentState{g: g}
	total := 0
	for _, modifier := range state.CostModifiersForSpell(game.Player1, card, 0, sourceZone, nil) {
		total += modifier.GenericReduction
	}
	return total
}

func TestSpellCostModifierGraveyardZoneAppliesOnlyFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, graveyardSpellReductionEffect(g, 1))
	card := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}}

	if got := spellGenericReductionFromZone(g, card, zone.Graveyard); got != 1 {
		t.Fatalf("reduction casting from graveyard = %d, want 1", got)
	}
	if got := spellGenericReductionFromZone(g, card, zone.Hand); got != 0 {
		t.Fatalf("reduction casting from hand = %d, want 0", got)
	}
}

func TestSpellCostModifierMatchesZone(t *testing.T) {
	scoped := game.CostModifier{Kind: game.CostModifierSpell, SourceZone: opt.Val(zone.Graveyard)}
	if !spellCostModifierMatchesZone(scoped, zone.Graveyard) {
		t.Fatal("graveyard-scoped modifier rejected a graveyard cast")
	}
	if spellCostModifierMatchesZone(scoped, zone.Hand) {
		t.Fatal("graveyard-scoped modifier matched a hand cast")
	}
	unscoped := game.CostModifier{Kind: game.CostModifierSpell}
	if !spellCostModifierMatchesZone(unscoped, zone.Hand) || !spellCostModifierMatchesZone(unscoped, zone.Graveyard) {
		t.Fatal("unscoped modifier should match any zone")
	}
	nonHand := game.CostModifier{
		Kind:        game.CostModifierSpell,
		SourceZones: []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command},
	}
	for _, z := range []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command} {
		if !spellCostModifierMatchesZone(nonHand, z) {
			t.Fatalf("non-hand-scoped modifier rejected a %v cast", z)
		}
	}
	if spellCostModifierMatchesZone(nonHand, zone.Hand) {
		t.Fatal("non-hand-scoped modifier matched a hand cast")
	}
}

// nonHandSpellReductionEffect models the active rule effect cardgen emits for
// "Spells you cast from anywhere other than your hand cost {N} less to cast."
// (Sage of the Beyond): a controller spell cost modifier scoped to the non-hand
// cast zones.
func nonHandSpellReductionEffect(g *game.Game, reduction int) game.RuleEffect {
	return game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCostModifier,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CostModifier: game.CostModifier{
			Kind:             game.CostModifierSpell,
			SourceZones:      []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command},
			GenericReduction: reduction,
		},
	}
}

func TestSpellCostModifierNonHandZoneAppliesOnlyOutsideHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, nonHandSpellReductionEffect(g, 2))
	card := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
	}}

	for _, z := range []zone.Type{zone.Exile, zone.Graveyard, zone.Library, zone.Command} {
		if got := spellGenericReductionFromZone(g, card, z); got != 2 {
			t.Fatalf("reduction casting from %v = %d, want 2", z, got)
		}
	}
	if got := spellGenericReductionFromZone(g, card, zone.Hand); got != 0 {
		t.Fatalf("reduction casting from hand = %d, want 0", got)
	}
}
