package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

const flareOfFortitudeOracle = "You may sacrifice a nontoken white creature rather than pay this spell's mana cost.\nUntil end of turn, your life total can't change, and permanents you control gain hexproof and indestructible."

func flareOfFortitudeCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Flare of Fortitude",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{W}{W}",
		OracleText: flareOfFortitudeOracle,
	}
}

func TestLowerFlareOfFortitude(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, flareOfFortitudeCard())

	// Free alternative cost: sacrifice a nontoken white creature rather than pay
	// mana. The alternative carries no mana cost and a single sacrifice payment.
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alt := face.AlternativeCosts[0]
	if alt.ManaCost.Exists {
		t.Fatalf("free alternative should carry no mana cost: %#v", alt)
	}
	if alt.Condition != cost.AlternativeConditionNone {
		t.Fatalf("condition = %v, want unconditional", alt.Condition)
	}
	if len(alt.AdditionalCosts) != 1 {
		t.Fatalf("additional costs = %#v, want a single sacrifice cost", alt.AdditionalCosts)
	}
	sacrifice := alt.AdditionalCosts[0]
	if sacrifice.Kind != cost.AdditionalSacrifice ||
		sacrifice.Amount != 1 ||
		!sacrifice.MatchPermanentType ||
		sacrifice.PermanentType != types.Creature ||
		!sacrifice.MatchCardColor ||
		sacrifice.CardColor != color.White ||
		!sacrifice.RequireNonToken {
		t.Fatalf("sacrifice cost = %#v, want sacrifice a nontoken white creature", sacrifice)
	}

	// Spell body: the ordered life-total-can't-change player rule followed by the
	// group hexproof/indestructible grant, both until end of turn.
	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("spell ability = %+v, want one mode", face.SpellAbility)
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2: %+v", len(sequence), sequence)
	}

	apply, ok := sequence[0].Primitive.(game.ApplyRule)
	if !ok || apply.Duration != game.DurationUntilEndOfTurn ||
		len(apply.RuleEffects) != 1 ||
		apply.RuleEffects[0].Kind != game.RuleEffectLifeTotalCantChange ||
		apply.RuleEffects[0].AffectedPlayer != game.PlayerYou {
		t.Fatalf("sequence[0] = %+v, want until-end-of-turn life-total-can't-change for you", sequence[0])
	}

	cont, ok := sequence[1].Primitive.(game.ApplyContinuous)
	if !ok || cont.Duration != game.DurationUntilEndOfTurn ||
		len(cont.ContinuousEffects) != 1 {
		t.Fatalf("sequence[1] = %+v, want until-end-of-turn continuous grant", sequence[1])
	}
	grant := cont.ContinuousEffects[0]
	if !grant.Group.Valid() {
		t.Fatalf("grant group = %+v, want the permanents you control", grant.Group)
	}
	selection := grant.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 0 {
		t.Fatalf("grant selection = %+v, want all permanents you control", selection)
	}
	if len(grant.AddKeywords) != 2 ||
		grant.AddKeywords[0] != game.Hexproof ||
		grant.AddKeywords[1] != game.Indestructible {
		t.Fatalf("grant keywords = %+v, want hexproof and indestructible", grant.AddKeywords)
	}
}

func TestGenerateFlareOfFortitudeSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(flareOfFortitudeCard(), "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	for _, want := range []string{
		"AlternativeCosts: []cost.Alternative{",
		"Kind:               cost.AdditionalSacrifice,",
		"RequireNonToken:    true,",
		"game.RuleEffectLifeTotalCantChange",
		"game.ApplyContinuous{",
		"Group: game.BattlefieldGroup(",
		"game.Hexproof,",
		"game.Indestructible,",
		"game.DurationUntilEndOfTurn,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestFlareOfFortitudeVariantsFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		// "Your life total can't change" without an until-end-of-turn scope must
		// not compose (Platinum Emperion's static rule stays unsupported).
		"Your life total can't change, and permanents you control gain hexproof and indestructible.",
		// A different life-clause wording is not the fixed recognized clause.
		"Until end of turn, your life total can't increase, and permanents you control gain hexproof and indestructible.",
		// An unrecognized trailing grant fails the whole compound sentence closed.
		"Until end of turn, your life total can't change, and permanents you control get +1/+1.",
	} {
		face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
			Name:       "Flare of Fortitude",
			Layout:     "normal",
			TypeLine:   "Instant",
			ManaCost:   "{2}{W}{W}",
			OracleText: oracleText,
		})
		if face.SpellAbility.Exists {
			t.Fatalf("unsupported variant produced a spell ability: %q", oracleText)
		}
	}
}
