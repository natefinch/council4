package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestLowerFilterLandManaAbility verifies that the filter-land cycle lowers both
// its colorless ability and its hybrid-cost two-color filter ability. The filter
// ability pays one hybrid {W/U} mana plus a tap and adds two mana, each
// independently chosen from the {W, U} pair.
func TestLowerFilterLandManaAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Mystic Gate",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{W/U}, {T}: Add {W}{W}, {W}{U}, or {U}{U}.",
	})
	if len(face.ManaAbilities) != 2 {
		t.Fatalf("mana abilities = %d, want 2", len(face.ManaAbilities))
	}
	filter := face.ManaAbilities[1]
	if !filter.ManaCost.Exists ||
		len(filter.ManaCost.Val) != 1 ||
		filter.ManaCost.Val[0] != cost.HybridMana(mana.W, mana.U) {
		t.Fatalf("filter mana cost = %#v, want {W/U}", filter.ManaCost)
	}
	if len(filter.AdditionalCosts) != 1 || filter.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("filter additional costs = %#v, want tap", filter.AdditionalCosts)
	}
	sequence := filter.Content.Modes[0].Sequence
	if len(sequence) != 4 {
		t.Fatalf("filter sequence = %#v, want choose+add+choose+add", sequence)
	}
	for _, index := range []int{0, 2} {
		choose, ok := sequence[index].Primitive.(game.Choose)
		if !ok || choose.Choice.Kind != game.ResolutionChoiceMana {
			t.Fatalf("sequence[%d] = %#v, want mana Choose", index, sequence[index].Primitive)
		}
		if len(choose.Choice.Colors) != 2 ||
			choose.Choice.Colors[0] != mana.W ||
			choose.Choice.Colors[1] != mana.U {
			t.Fatalf("sequence[%d] colors = %#v, want [W U]", index, choose.Choice.Colors)
		}
	}
	firstChoose, ok := sequence[0].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want Choose", sequence[0].Primitive)
	}
	secondChoose, ok := sequence[2].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("sequence[2] = %#v, want Choose", sequence[2].Primitive)
	}
	if firstChoose.PublishChoice == secondChoose.PublishChoice {
		t.Fatal("the two filter choices must publish under distinct keys")
	}
	for _, index := range []int{1, 3} {
		add, ok := sequence[index].Primitive.(game.AddMana)
		if !ok || add.Amount != game.Fixed(1) {
			t.Fatalf("sequence[%d] = %#v, want AddMana of 1", index, sequence[index].Primitive)
		}
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("filter ability instruction sequence invalid: %v", err)
	}
}

// TestLowerFilterLandManaAbilityFailsClosed verifies that near-miss filter-land
// bodies do not lower a mana ability and fail closed with a diagnostic.
func TestLowerFilterLandManaAbilityFailsClosed(t *testing.T) {
	t.Parallel()
	for name, oracle := range map[string]string{
		"truncated final group": "{T}: Add {C}.\n{W/U}, {T}: Add {W}{W}, {W}{U}, or {U}.",
		"colors disagree":       "{T}: Add {C}.\n{W/U}, {T}: Add {W}{W}, {W}{B}, or {U}{U}.",
		"single color":          "{T}: Add {C}.\n{W/W}, {T}: Add {W}{W}, {W}{W}, or {W}{W}.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Filter Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: oracle,
			})
			for i := range face.ManaAbilities {
				if content := face.ManaAbilities[i].Content; len(content.Modes) == 1 && len(content.Modes[0].Sequence) == 4 {
					t.Fatalf("%s unexpectedly lowered a filter mana ability", name)
				}
			}
		})
	}
}

// TestTwoColorFilterManaAbilityTemplate exercises the runtime template directly,
// confirming it builds a valid, faithful instruction sequence and rejects
// non-pair color inputs.
func TestTwoColorFilterManaAbilityTemplate(t *testing.T) {
	t.Parallel()
	ability := game.TwoColorFilterManaAbility(mana.B, mana.R)
	if ability.Text != "{B/R}, {T}: Add {B}{B}, {B}{R}, or {R}{R}." {
		t.Fatalf("text = %q", ability.Text)
	}
	if !ability.ManaCost.Exists || ability.ManaCost.Val[0] != cost.HybridMana(mana.B, mana.R) {
		t.Fatalf("mana cost = %#v", ability.ManaCost)
	}
	if err := game.ValidateInstructionSequence(ability.Content.Modes[0].Sequence); err != nil {
		t.Fatalf("template sequence invalid: %v", err)
	}
	for _, bad := range [][2]mana.Color{{mana.W, mana.W}, {mana.C, mana.U}} {
		func(first, second mana.Color) {
			defer func() {
				if recover() == nil {
					t.Fatalf("TwoColorFilterManaAbility(%q, %q) did not panic", first, second)
				}
			}()
			_ = game.TwoColorFilterManaAbility(first, second)
		}(bad[0], bad[1])
	}
}
