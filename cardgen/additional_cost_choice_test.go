package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// redirectLightningText is the authoritative Oracle text for Redirect Lightning:
// a printed choice among alternative additional costs ("pay 5 life or pay {2}")
// followed by the change-target effect.
const redirectLightningText = "As an additional cost to cast this spell, pay 5 life or pay {2}.\n" +
	"Change the target of target spell or ability with a single target."

// TestLowerRedirectLightningAdditionalCostChoice proves Redirect Lightning's
// "pay 5 life or pay {2}" lowers to a single cost.AdditionalChoice with exactly
// two branches — a non-mana pay-5-life branch and an additive {2} mana branch —
// and never leaks into the mandatory cost.Additional list. The change-target
// spell ability is retained with a single stack-object target.
func TestLowerRedirectLightningAdditionalCostChoice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Redirect Lightning",
		Layout:     "normal",
		TypeLine:   "Instant — Lesson",
		ManaCost:   "{R}",
		OracleText: redirectLightningText,
	})
	if len(face.AdditionalCosts) != 0 {
		t.Fatalf("AdditionalCosts = %#v, want none (a choice routes to AdditionalCostChoices)", face.AdditionalCosts)
	}
	if len(face.AdditionalCostChoices) != 1 {
		t.Fatalf("AdditionalCostChoices = %#v, want exactly one choice", face.AdditionalCostChoices)
	}
	options := face.AdditionalCostChoices[0].Options
	if len(options) != 2 {
		t.Fatalf("choice options = %#v, want two branches", options)
	}

	life := options[0]
	if life.Label != "Pay 5 life" {
		t.Fatalf("life branch label = %q, want \"Pay 5 life\"", life.Label)
	}
	if len(life.Mana) != 0 {
		t.Fatalf("life branch mana = %#v, want none (life is not mana)", life.Mana)
	}
	if len(life.Costs) != 1 || life.Costs[0].Kind != cost.AdditionalPayLife || life.Costs[0].Amount != 5 {
		t.Fatalf("life branch costs = %#v, want a single pay-5-life", life.Costs)
	}
	if life.Costs[0].ChoiceGroup != 0 {
		t.Fatalf("life branch cost carries ChoiceGroup %d, want 0 (the branch itself is the choice)", life.Costs[0].ChoiceGroup)
	}

	pay := options[1]
	if pay.Label != "Pay {2}" {
		t.Fatalf("mana branch label = %q, want \"Pay {2}\"", pay.Label)
	}
	if len(pay.Costs) != 0 {
		t.Fatalf("mana branch costs = %#v, want none", pay.Costs)
	}
	if !slices.Equal(pay.Mana, cost.Mana{cost.O(2)}) {
		t.Fatalf("mana branch mana = %#v, want {2}", pay.Mana)
	}

	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].MinTargets != 1 || mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("targets = %#v, want a single stack-object target", mode.Targets)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.ChooseNewTargets); !ok {
		t.Fatalf("primitive = %#v, want ChooseNewTargets", mode.Sequence[0].Primitive)
	}
}

// TestLowerBayouGroffSacrificeOrManaChoice proves a beneficiary whose choice
// mixes a sacrifice branch with a mana branch lowers to one cost.AdditionalChoice
// carrying the sacrifice as a non-mana branch and the {3} as an additive mana
// branch.
func TestLowerBayouGroffSacrificeOrManaChoice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Bayou Groff",
		Layout:     "normal",
		TypeLine:   "Creature — Plant Dog",
		ManaCost:   "{1}{G}",
		OracleText: "As an additional cost to cast this spell, sacrifice a creature or pay {3}.",
	})
	if len(face.AdditionalCostChoices) != 1 {
		t.Fatalf("AdditionalCostChoices = %#v, want exactly one choice", face.AdditionalCostChoices)
	}
	options := face.AdditionalCostChoices[0].Options
	if len(options) != 2 {
		t.Fatalf("choice options = %#v, want two branches", options)
	}

	sac := options[0]
	if len(sac.Mana) != 0 {
		t.Fatalf("sacrifice branch mana = %#v, want none", sac.Mana)
	}
	if len(sac.Costs) != 1 || sac.Costs[0].Kind != cost.AdditionalSacrifice ||
		!sac.Costs[0].MatchPermanentType || sac.Costs[0].PermanentType != types.Creature {
		t.Fatalf("sacrifice branch costs = %#v, want a single sacrifice-a-creature", sac.Costs)
	}

	pay := options[1]
	if len(pay.Costs) != 0 {
		t.Fatalf("mana branch costs = %#v, want none", pay.Costs)
	}
	if !slices.Equal(pay.Mana, cost.Mana{cost.O(3)}) {
		t.Fatalf("mana branch mana = %#v, want {3}", pay.Mana)
	}
}

// TestLowerPureNonManaChoiceKeepsChoiceGroupPath proves a pure non-mana additional
// cost choice ("sacrifice a creature or discard a card") is unaffected by the
// mana-choice feature: it stays on the comparable cost.Additional ChoiceGroup path
// and never routes to AdditionalCostChoices. This keeps every existing pure
// non-mana choice card byte-identical after the change.
func TestLowerPureNonManaChoiceKeepsChoiceGroupPath(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Bone Shards",
		Layout:   "normal",
		TypeLine: "Sorcery",
		ManaCost: "{B}",
		OracleText: "As an additional cost to cast this spell, sacrifice a creature or discard a card.\n" +
			"Destroy target creature or planeswalker.",
	})
	if len(face.AdditionalCostChoices) != 0 {
		t.Fatalf("AdditionalCostChoices = %#v, want none (pure non-mana stays on the old path)", face.AdditionalCostChoices)
	}
	if len(face.AdditionalCosts) != 2 {
		t.Fatalf("AdditionalCosts = %#v, want two ChoiceGroup members", face.AdditionalCosts)
	}
	if face.AdditionalCosts[0].ChoiceGroup == 0 ||
		face.AdditionalCosts[0].ChoiceGroup != face.AdditionalCosts[1].ChoiceGroup {
		t.Fatalf("choice groups = %d,%d; want equal and nonzero",
			face.AdditionalCosts[0].ChoiceGroup, face.AdditionalCosts[1].ChoiceGroup)
	}
}

// TestLowerBoltBendUnaffectedByChoiceFeature proves Bolt Bend — the ferocious
// change-target spell already curated — still lowers cleanly, carries no
// additional-cost choice, and keeps its self cost-modifier static ability and
// change-target spell ability.
func TestLowerBoltBendUnaffectedByChoiceFeature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Bolt Bend",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{3}{R}",
		OracleText: "This spell costs {3} less to cast if you control a creature with power 4 or greater.\n" +
			"Change the target of target spell or ability with a single target.",
	})
	if len(face.AdditionalCostChoices) != 0 {
		t.Fatalf("AdditionalCostChoices = %#v, want none", face.AdditionalCostChoices)
	}
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("StaticAbilities = %#v, want one self cost-modifier", face.StaticAbilities)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if _, ok := mode.Sequence[0].Primitive.(game.ChooseNewTargets); !ok {
		t.Fatalf("primitive = %#v, want ChooseNewTargets", mode.Sequence[0].Primitive)
	}
}

// TestLowerLoneManaAdditionalCostFailsClosed proves a mandatory (non-choice)
// "pay {2}" additional cost fails closed with a diagnostic rather than being
// silently dropped. Mana cannot be represented as a comparable cost.Additional,
// and only a "<cost> or <cost>" choice with a mana branch routes to the additive
// AdditionalChoice path, so a lone mana additional cost is unsupported.
func TestLowerLoneManaAdditionalCostFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Lone Mana Tax",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		OracleText: "As an additional cost to cast this spell, pay {2}.\nDraw a card.",
	})
	if len(face.AdditionalCostChoices) != 0 {
		t.Fatalf("AdditionalCostChoices = %#v, want none on fail-closed", face.AdditionalCostChoices)
	}
	if len(face.AdditionalCosts) != 0 {
		t.Fatalf("AdditionalCosts = %#v, want none on fail-closed", face.AdditionalCosts)
	}
}

// TestGenerateRedirectLightningSource proves the full parser→compiler→lower→render
// pipeline emits Redirect Lightning's additive additional-cost choice as
// executable Go: a cost.AdditionalChoice with a pay-5-life branch and an additive
// {2} mana branch, and a ChooseNewTargets spell ability. The generated source
// compiles as part of ./..., so this locks the rendered shape the curated card
// relies on.
func TestGenerateRedirectLightningSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Redirect Lightning",
		Layout:     "normal",
		TypeLine:   "Instant — Lesson",
		ManaCost:   "{R}",
		OracleText: redirectLightningText,
	}, "redirectLightning")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"AdditionalCostChoices: []cost.AdditionalChoice{",
		"Options: []cost.AdditionalChoiceOption{",
		`Label: "Pay 5 life"`,
		"Kind:   cost.AdditionalPayLife,",
		"Amount: 5,",
		`Label: "Pay {2}"`,
		"Mana:  cost.Mana{cost.O(2)}",
		"game.ChooseNewTargets{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
