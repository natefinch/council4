package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const underworldBreachOracle = "Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three other cards from your graveyard. (You may cast cards from your graveyard for their escape cost.)\n" +
	"At the beginning of the end step, sacrifice this enchantment."

func underworldBreachCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Underworld Breach",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Enchantment",
		OracleText: underworldBreachOracle,
	}
}

// TestGenerateExecutableCardSourceUnderworldBreach proves the full
// parser→compiler→lowering→rendering pipeline turns Underworld Breach's text
// into a graveyard-escape keyword grant carrying the computed escape cost, plus
// the separate end-step self-sacrifice, with no diagnostics.
func TestGenerateExecutableCardSourceUnderworldBreach(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(underworldBreachCard(), "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Kind:           game.RuleEffectGrantGraveyardCardKeyword",
		"AffectedPlayer: game.PlayerYou",
		"GrantedKeyword: game.Escape",
		"CardSelection:  game.Selection{ExcludedTypes: []types.Card{types.Land}}",
		"GraveyardCastCost: game.GraveyardCastGrantCost{",
		"UseCardManaCost: true,",
		"Kind:          cost.AdditionalExile,",
		"Amount:        3,",
		"Source:        zone.Graveyard,",
		"ExcludeSource: true,",
		"game.Sacrifice{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerUnderworldBreachGrantsGraveyardEscape checks the lowered rule effect:
// escape granted to nonland graveyard cards with the computed cost (the card's
// own mana cost plus exiling three other graveyard cards, excluding the spell).
func TestLowerUnderworldBreachGrantsGraveyardEscape(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, underworldBreachCard())
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	grant := face.StaticAbilities[0].Body
	if len(grant.RuleEffects) != 1 {
		t.Fatalf("rule effects = %d, want 1", len(grant.RuleEffects))
	}
	effect := grant.RuleEffects[0]
	if effect.Kind != game.RuleEffectGrantGraveyardCardKeyword {
		t.Fatalf("rule effect kind = %v", effect.Kind)
	}
	if effect.GrantedKeyword != game.Escape {
		t.Fatalf("granted keyword = %v, want Escape", effect.GrantedKeyword)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if effect.RestrictedDuringControllerTurn {
		t.Fatal("escape grant is not restricted to the controller's turn")
	}
	if len(effect.CardSelection.RequiredTypesAny) != 0 {
		t.Fatalf("required types any = %#v, want none (any nonland card)", effect.CardSelection.RequiredTypesAny)
	}
	if len(effect.CardSelection.ExcludedTypes) != 1 || effect.CardSelection.ExcludedTypes[0] != types.Land {
		t.Fatalf("excluded types = %#v, want [Land]", effect.CardSelection.ExcludedTypes)
	}
	castCost := effect.GraveyardCastCost
	if !castCost.UseCardManaCost {
		t.Fatal("escape cost must use the card's own mana cost")
	}
	if len(castCost.AdditionalCosts) != 1 {
		t.Fatalf("additional costs = %d, want 1", len(castCost.AdditionalCosts))
	}
	exile := castCost.AdditionalCosts[0]
	if exile.Kind != cost.AdditionalExile || exile.Amount != 3 || exile.Source != zone.Graveyard || !exile.ExcludeSource {
		t.Fatalf("exile additional cost = %+v, want exile 3 others from graveyard excluding source", exile)
	}
}

// TestUnderworldBreachEscapeGrantFailsClosed proves unsupported escape-grant
// wordings never lower to a silent partial ability: each must report a
// diagnostic. The parser owns the wording, so a deviation in the cost clause,
// the exiled-card count phrase, or a missing computed cost fails closed.
func TestUnderworldBreachEscapeGrantFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		// Escape granted with no computed cost clause at all.
		"Each nonland card in your graveyard has escape.",
		// Cost clause pays life instead of exiling other graveyard cards.
		"Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus pay 3 life.",
		// Cost clause exiles from the battlefield rather than the graveyard.
		"Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three other cards from the battlefield.",
		// Cost clause omits the "other" exclusion of the escaping card.
		"Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three cards from your graveyard.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Near Breach",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: oracle,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", oracle)
			}
			for _, face := range faces {
				for _, ability := range face.StaticAbilities {
					for _, effect := range ability.Body.RuleEffects {
						if effect.Kind == game.RuleEffectGrantGraveyardCardKeyword && effect.GrantedKeyword == game.Escape {
							t.Fatalf("unsupported escape grant lowered to a partial ability for %q", oracle)
						}
					}
				}
			}
		})
	}
}
