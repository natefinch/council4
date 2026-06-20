package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceRoamingThroneCategory(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Roaming Throne",
		Layout:     "normal",
		ManaCost:   "{4}",
		TypeLine:   "Artifact Creature — Golem",
		OracleText: "Ward {2}\nAs this creature enters, choose a creature type.\nThis creature is the chosen type in addition to its other types.\nIf a triggered ability of another creature you control of the chosen type triggers, it triggers an additional time.",
		Power:      new("4"),
		Toughness:  new("4"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntryTypeChoiceReplacement",
		"AddSubtypeFromEntryChoice: game.EntryTypeChoiceKey",
		"Kind: game.RuleEffectAdditionalTriggerForChosenCreatureType",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}

}

func TestGenerateExecutableCardSourceRoamingThroneNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for name, ability := range map[string]string{
		"chosen subtype": "This creature is a chosen type in addition to its other types.",
		"trigger source": "If a triggered ability of another permanent you control of the chosen type triggers, it triggers an additional time.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Near Miss Throne",
				Layout:     "normal",
				ManaCost:   "{4}",
				TypeLine:   "Artifact Creature — Golem",
				OracleText: "As this creature enters, choose a creature type.\n" + ability,
				Power:      new("4"),
				Toughness:  new("4"),
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "n")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v; want fail closed", source, diagnostics)
			}
		})
	}
}
