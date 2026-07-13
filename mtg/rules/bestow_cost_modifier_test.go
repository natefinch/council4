package rules

import (
	"testing"

	cardsb "github.com/natefinch/council4/mtg/cards/b"
	cardsg "github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// bestowSpellGenericReduction sums the generic cost reductions the pre-payment
// modifier path (CR 601.2f) resolves for casting card from the player's hand,
// with the bestowed flag selecting the normal-creature vs Aura cast branch. It
// mirrors sourceSpellGenericReduction but threads the branch that the payment
// planner now passes into CostModifiersForSpell.
func bestowSpellGenericReduction(g *game.Game, playerID game.PlayerID, card *game.CardDef, bestowed bool) int {
	state := &rulesPaymentState{g: g}
	total := 0
	for _, modifier := range state.CostModifiersForSpell(playerID, card, 0, zone.Hand, nil, false, bestowed) {
		total += modifier.GenericReduction
	}
	return total
}

// TestGoreclawReducesNormalCastButNotBestowedBoonSatyr drives the real Goreclaw,
// Terror of Qal Sisma ("Creature spells you cast with power 4 or greater cost
// {2} less") against the real Boon Satyr (power-4 Enchantment Creature with
// Bestow). A normal creature cast qualifies and is reduced by {2}, but a bestowed
// cast is an Aura spell and not a creature spell (CR 702.103b), so during 601.2f
// cost determination Goreclaw's creature filter must not match — the pre-payment
// bug that wrongly cheapened bestowed creatures.
func TestGoreclawReducesNormalCastButNotBestowedBoonSatyr(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cardsg.GoreclawTerrorOfQalSisma())

	boonSatyr := cardsb.BoonSatyr()

	if got := bestowSpellGenericReduction(g, game.Player1, boonSatyr, false); got != 2 {
		t.Fatalf("normal creature cast reduction = %d, want 2", got)
	}
	if got := bestowSpellGenericReduction(g, game.Player1, boonSatyr, true); got != 0 {
		t.Fatalf("bestowed cast reduction = %d, want 0 (Aura spell, not creature)", got)
	}
}

// TestAuraCostModifierAppliesToBestowNotNormalCast is the converse of the
// Goreclaw case: an "Aura spells you cast cost {2} less" modifier must reduce a
// bestowed cast (which is an Aura spell) but leave a normal creature cast of the
// same card untouched. It confirms castSelectionFace adds the Aura subtype to the
// bestow branch during cost determination.
func TestAuraCostModifierAppliesToBestowNotNormalCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.CostModifiers = append(g.CostModifiers, game.CostModifier{
		Kind:             game.CostModifierSpell,
		CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Aura}},
		GenericReduction: 2,
	})

	boonSatyr := cardsb.BoonSatyr()

	if got := bestowSpellGenericReduction(g, game.Player1, boonSatyr, true); got != 2 {
		t.Fatalf("bestowed Aura cast reduction = %d, want 2", got)
	}
	if got := bestowSpellGenericReduction(g, game.Player1, boonSatyr, false); got != 0 {
		t.Fatalf("normal creature cast reduction = %d, want 0 (not an Aura spell)", got)
	}
}

// TestCreatureCostModifierUnchangedByCastSelectionFace guards the printed/native
// path: with no bestow branch (bestowed=false), a creature-filtered modifier
// still matches a plain Enchantment Creature exactly as before, proving
// castSelectionFace is a no-op for normal casts and does not mutate the CardDef.
func TestCreatureCostModifierUnchangedByCastSelectionFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.CostModifiers = append(g.CostModifiers, game.CostModifier{
		Kind:             game.CostModifierSpell,
		CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Creature}},
		GenericReduction: 1,
	})

	boonSatyr := cardsb.BoonSatyr()
	before := boonSatyr.Types

	if got := bestowSpellGenericReduction(g, game.Player1, boonSatyr, false); got != 1 {
		t.Fatalf("normal Enchantment Creature reduction = %d, want 1", got)
	}
	// castSelectionFace must never mutate the original card's printed types.
	if len(boonSatyr.Types) != len(before) ||
		!boonSatyr.HasType(types.Creature) || !boonSatyr.HasType(types.Enchantment) {
		t.Fatalf("printed types mutated: %v", boonSatyr.Types)
	}
	// A bestowed read of the same card must not have altered the shared def.
	_ = bestowSpellGenericReduction(g, game.Player1, boonSatyr, true)
	if !boonSatyr.HasType(types.Creature) {
		t.Fatal("bestow selection read mutated the shared CardDef")
	}
}

// TestBestowAffordabilityUsesCorrectModifierBranch exercises the end-to-end
// alternative-cost option math: with Goreclaw in play the affordability check for
// a bestowed Boon Satyr must plan the full {3}{G}{G} bestow cost (no creature
// reduction), so a player holding exactly that mana can legally cast it, while a
// player one mana short cannot.
func TestBestowAffordabilityUsesCorrectModifierBranch(t *testing.T) {
	setup := func(green int) (*game.Game, *Engine, id.ID, []game.Target) {
		g, engine := setupBestowMain(t)
		addCombatPermanent(g, game.Player1, cardsg.GoreclawTerrorOfQalSisma())
		target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		spellID := addCardToHand(g, game.Player1, cardsb.BoonSatyr())
		g.Players[game.Player1].ManaPool.Add(mana.G, green)
		return g, engine, spellID, []game.Target{game.PermanentTarget(target.ObjectID)}
	}

	// Bestow cost is {3}{G}{G} = five mana; Goreclaw must not reduce it.
	fullG, fullEngine, fullID, fullTargets := setup(5)
	if !fullEngine.applyAction(fullG, game.Player1, action.CastBestowSpell(fullID, fullTargets, 0, nil)) {
		t.Fatal("bestowed cast with exactly {3}{G}{G} available should be legal")
	}

	shortG, shortEngine, shortID, shortTargets := setup(4)
	if shortEngine.applyAction(shortG, game.Player1, action.CastBestowSpell(shortID, shortTargets, 0, nil)) {
		t.Fatal("bestowed cast with only four mana should be illegal (no creature reduction)")
	}
}
