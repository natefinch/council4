package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// legacyCostModifierMatchesCard is the per-field card matcher that
// spellCostModifierBaseMatchesCard used before cost modifiers routed through the
// shared card-subject Selection matcher. The parity test pins the new Selection
// path to this exact reference behavior across a spread of cards and filters.
func legacyCostModifierMatchesCard(modifier game.CostModifier, card *game.CardDef) bool {
	if modifier.MatchCardType && (card == nil || !card.HasType(modifier.CardType)) {
		return false
	}
	if modifier.MatchExcludedCardType && (card == nil || card.HasType(modifier.ExcludedCardType)) {
		return false
	}
	if modifier.MatchColor {
		if card == nil {
			return false
		}
		if modifier.Color == "" {
			if len(card.Colors) != 0 {
				return false
			}
		} else if !slices.Contains(card.Colors, modifier.Color) {
			return false
		}
	}
	if len(modifier.MatchColors) != 0 {
		if card == nil {
			return false
		}
		matched := false
		for _, c := range modifier.MatchColors {
			if slices.Contains(card.Colors, c) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(modifier.MatchSubtypes) != 0 {
		if card == nil {
			return false
		}
		if !slices.ContainsFunc(modifier.MatchSubtypes, card.HasSubtype) {
			return false
		}
	}
	if modifier.MinPower.Exists {
		if card == nil {
			return false
		}
		power := card.Power
		if !power.Exists || power.Val.IsStar || power.Val.Value < modifier.MinPower.Val {
			return false
		}
	}
	return true
}

func cardDef(face game.CardFace) *game.CardDef {
	return &game.CardDef{CardFace: face}
}

func TestSpellCostModifierSelectionMatchesLegacy(t *testing.T) {
	t.Parallel()

	modifiers := map[string]game.CostModifier{
		"no filter":          {},
		"creature type":      {MatchCardType: true, CardType: types.Creature},
		"artifact type":      {MatchCardType: true, CardType: types.Artifact},
		"excluded creature":  {MatchExcludedCardType: true, ExcludedCardType: types.Creature},
		"single red":         {MatchColor: true, Color: color.Red},
		"colorless sentinel": {MatchColor: true, Color: ""},
		"red or green":       {MatchColors: []color.Color{color.Red, color.Green}},
		"aura or equipment":  {MatchSubtypes: []types.Sub{types.Aura, types.Equipment}},
		"min power 4":        {MinPower: opt.Val(4)},
		"black creature":     {MatchCardType: true, CardType: types.Creature, MatchColor: true, Color: color.Black},
		"red goblin power 2": {MatchColor: true, Color: color.Red, MatchSubtypes: []types.Sub{types.Goblin}, MinPower: opt.Val(2)},
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
			want := legacyCostModifierMatchesCard(modifier, card)
			got := spellCostModifierBaseMatchesCard(g, modifier, card)
			if got != want {
				t.Errorf("modifier %q vs card %q: selection match = %v, legacy = %v", modName, cardName, got, want)
			}
		}
	}
}

func TestCostModifierCardSubjectSelectionPrefersExplicit(t *testing.T) {
	t.Parallel()

	explicit := game.Selection{RequiredTypes: []types.Card{types.Artifact}}
	modifier := game.CostModifier{
		MatchCardType: true,
		CardType:      types.Creature,
		CardSelection: explicit,
	}
	creature := cardDef(game.CardFace{Types: []types.Card{types.Creature}})
	artifact := cardDef(game.CardFace{Types: []types.Card{types.Artifact}})

	g := &game.Game{}
	if spellCostModifierBaseMatchesCard(g, modifier, creature) {
		t.Error("explicit artifact CardSelection should not match a creature")
	}
	if !spellCostModifierBaseMatchesCard(g, modifier, artifact) {
		t.Error("explicit artifact CardSelection should match an artifact")
	}
}
