package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// referenceCostSelectionMatches is an independent reimplementation of the
// card-subject filter a spell cost modifier applies, reading the canonical
// CardSelection's printed-characteristic fields directly. The parity test pins
// the production Selection matcher (spellCostModifierBaseMatchesCard, which
// routes through cardDefMatchesCostSelection) to this reference across a spread
// of representative cards and filters so the set of cards a modifier matches —
// and therefore the set whose cost it changes — stays identical.
func referenceCostSelectionMatches(sel game.Selection, card *game.CardDef) bool {
	if sel.Empty() {
		return true
	}
	if card == nil {
		return false
	}
	for _, t := range sel.RequiredTypes {
		if !card.HasType(t) {
			return false
		}
	}
	if slices.ContainsFunc(sel.ExcludedTypes, card.HasType) {
		return false
	}
	if sel.Colorless && len(card.Colors) != 0 {
		return false
	}
	if len(sel.ColorsAny) != 0 {
		if !slices.ContainsFunc(sel.ColorsAny, func(c color.Color) bool {
			return slices.Contains(card.Colors, c)
		}) {
			return false
		}
	}
	if len(sel.SubtypesAny) != 0 {
		if !slices.ContainsFunc(sel.SubtypesAny, card.HasSubtype) {
			return false
		}
	}
	if sel.Power.Exists {
		power := card.Power
		if !power.Exists || power.Val.IsStar || !sel.Power.Val.Matches(power.Val.Value) {
			return false
		}
	}
	return true
}

func cardDef(face game.CardFace) *game.CardDef {
	return &game.CardDef{CardFace: face}
}

// TestSpellCostModifierSelectionMatchesReference proves that the canonical
// CardSelection card filter matches the same set of cards the reference
// interpretation does for the representative cost reducers Stage 1b cares about:
// an artifact-cost reducer (Ruby Medallion), a creature-spell reducer, an
// excluded-card-type reducer (Elspeth Conquers Death), single-color and
// colorless reducers, a color-disjunction reducer, a subtype reducer, a
// minimum-power reducer (Goreclaw), and combined color+subtype+power filters.
func TestSpellCostModifierSelectionMatchesReference(t *testing.T) {
	t.Parallel()

	modifiers := map[string]game.CostModifier{
		"no filter":          {Kind: game.CostModifierSpell, GenericReduction: 1},
		"artifact reducer":   {Kind: game.CostModifierSpell, GenericReduction: 1, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}}},
		"creature reducer":   {Kind: game.CostModifierSpell, GenericReduction: 1, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}}},
		"excluded creature":  {Kind: game.CostModifierSpell, GenericIncrease: 2, CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}}},
		"single red":         {Kind: game.CostModifierSpell, GenericReduction: 1, CardSelection: game.Selection{ColorsAny: []color.Color{color.Red}}},
		"colorless sentinel": {Kind: game.CostModifierSpell, GenericReduction: 1, CardSelection: game.Selection{Colorless: true}},
		"red or green":       {Kind: game.CostModifierSpell, GenericReduction: 1, CardSelection: game.Selection{ColorsAny: []color.Color{color.Red, color.Green}}},
		"aura or equipment":  {Kind: game.CostModifierSpell, GenericReduction: 1, CardSelection: game.Selection{SubtypesAny: []types.Sub{types.Aura, types.Equipment}}},
		"min power 4":        {Kind: game.CostModifierSpell, GenericReduction: 2, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}},
		"red goblin power 2": {Kind: game.CostModifierSpell, GenericReduction: 1, CardSelection: game.Selection{ColorsAny: []color.Color{color.Red}, SubtypesAny: []types.Sub{types.Goblin}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 2})}},
	}

	cards := map[string]*game.CardDef{
		"nil": nil,
		"red creature power 5": cardDef(game.CardFace{
			Types:  []types.Card{types.Creature},
			Colors: []color.Color{color.Red},
			Power:  opt.Val(game.PT{Value: 5}),
		}),
		"colorless artifact": cardDef(game.CardFace{
			Types: []types.Card{types.Artifact},
		}),
		"green goblin power 1": cardDef(game.CardFace{
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Goblin},
			Colors:   []color.Color{color.Green},
			Power:    opt.Val(game.PT{Value: 1}),
		}),
		"red goblin power 4": cardDef(game.CardFace{
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Goblin},
			Colors:   []color.Color{color.Red},
			Power:    opt.Val(game.PT{Value: 4}),
		}),
		"white aura": cardDef(game.CardFace{
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			Colors:   []color.Color{color.White},
		}),
		"blue equipment": cardDef(game.CardFace{
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			Colors:   []color.Color{color.Blue},
		}),
		"star power creature": cardDef(game.CardFace{
			Types:  []types.Card{types.Creature},
			Colors: []color.Color{color.Red},
			Power:  opt.Val(game.PT{IsStar: true}),
		}),
		"powerless instant": cardDef(game.CardFace{
			Types:  []types.Card{types.Instant},
			Colors: []color.Color{color.Red},
		}),
		"multicolor creature": cardDef(game.CardFace{
			Types:  []types.Card{types.Creature},
			Colors: []color.Color{color.Red, color.White},
			Power:  opt.Val(game.PT{Value: 3}),
		}),
	}

	g := &game.Game{}
	for modName, modifier := range modifiers {
		for cardName, card := range cards {
			want := referenceCostSelectionMatches(modifier.CardSelection, card)
			got := spellCostModifierBaseMatchesCard(g, modifier, card)
			if got != want {
				t.Errorf("modifier %q vs card %q: selection match = %v, reference = %v", modName, cardName, got, want)
			}
		}
	}
}
