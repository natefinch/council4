package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const teferisProtectionOracle = "Until your next turn, your life total can't change and you gain protection from everything. All permanents you control phase out. (While they're phased out, they're treated as though they don't exist. They phase in before you untap during your untap step.)\nExile Teferi's Protection."

func TestLowerTeferisProtection(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Teferi's Protection",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: teferisProtectionOracle,
	})
	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("spell ability = %+v, want one mode", face.SpellAbility)
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 4 {
		t.Fatalf("sequence length = %d, want 4: %+v", len(sequence), sequence)
	}
	for i, want := range []game.RuleEffectKind{
		game.RuleEffectLifeTotalCantChange,
		game.RuleEffectPlayerProtection,
	} {
		apply, ok := sequence[i].Primitive.(game.ApplyRule)
		if !ok || apply.Duration != game.DurationUntilYourNextTurn ||
			len(apply.RuleEffects) != 1 ||
			apply.RuleEffects[0].Kind != want ||
			apply.RuleEffects[0].AffectedPlayer != game.PlayerYou {
			t.Fatalf("sequence[%d] = %+v, want player rule effect %v", i, sequence[i], want)
		}
		if want == game.RuleEffectPlayerProtection && !apply.RuleEffects[0].Protection.Everything {
			t.Fatalf("sequence[%d] = %+v, want protection from everything", i, sequence[i])
		}
	}
	phase, ok := sequence[2].Primitive.(game.PhaseOut)
	if !ok || !phase.Group.Valid() {
		t.Fatalf("sequence[2] = %+v, want group PhaseOut", sequence[2])
	}
	selection := phase.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 0 {
		t.Fatalf("phase-out selection = %+v, want all permanents you control", selection)
	}
	exile, ok := sequence[3].Primitive.(game.Exile)
	if !ok || !exile.SourceSpell {
		t.Fatalf("sequence[3] = %+v, want source-spell exile", sequence[3])
	}
}

func TestGenerateTeferisProtectionSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Teferi's Protection",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: teferisProtectionOracle,
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	for _, want := range []string{
		"game.RuleEffectLifeTotalCantChange",
		"game.RuleEffectPlayerProtection",
		"game.ProtectionKeyword{Everything: true}",
		"game.DurationUntilYourNextTurn",
		"game.PhaseOut{",
		"Group: game.BattlefieldGroup(",
		"game.Exile{",
		"SourceSpell: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestTeferisProtectionVariantsFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Until your next turn, your life total can't increase or decrease and you gain protection from everything. All permanents you control phase out. Exile Teferi's Protection.",
		"Until your next turn, your life total can't change and you gain protection from everything. All nonland permanents you control phase out. Exile Teferi's Protection.",
	} {
		face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
			Name:       "Teferi's Protection",
			Layout:     "normal",
			TypeLine:   "Instant",
			OracleText: oracleText,
		})
		if face.SpellAbility.Exists {
			t.Fatalf("unsupported variant produced a spell ability: %q", oracleText)
		}
	}
}

func TestLowerSourceSpellExileRequiresSpellShell(t *testing.T) {
	t.Parallel()
	for _, typeLine := range []string{"Instant", "Sorcery"} {
		t.Run(typeLine, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Spell",
				Layout:     "normal",
				TypeLine:   typeLine,
				OracleText: "Exile Test Spell.",
			})
			if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) != 1 {
				t.Fatalf("spell ability = %+v, want one mode", face.SpellAbility)
			}
			sequence := face.SpellAbility.Val.Modes[0].Sequence
			if len(sequence) != 1 {
				t.Fatalf("sequence = %+v, want one instruction", sequence)
			}
			exile, ok := sequence[0].Primitive.(game.Exile)
			if !ok || !exile.SourceSpell {
				t.Fatalf("instruction = %+v, want source-spell exile", sequence[0])
			}
		})
	}

	tests := map[string]string{
		"activated":          "{T}: Exile Test Relic.",
		"activated sequence": "{T}: Draw a card. Exile Test Relic.",
		"triggered":          "When Test Relic enters, exile Test Relic.",
		"triggered sequence": "When Test Relic enters, draw a card. Exile Test Relic.",
	}
	for name, oracleText := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Relic",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracleText,
			})
			if len(face.ActivatedAbilities) != 0 ||
				len(face.TriggeredAbilities) != 0 ||
				face.SpellAbility.Exists {
				t.Fatalf("unsupported %s source exile produced an ability: %+v", name, face)
			}
		})
	}
}
