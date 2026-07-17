package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// theMycosynthGardensCard is the Scryfall shape of The Mycosynth Gardens, whose
// become-a-copy activated ability binds an X activation cost to the copied
// artifact's mana value ("{X}, {T}: This land becomes a copy of target nontoken
// artifact you control with mana value X.") alongside two mana abilities.
func theMycosynthGardensCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "The Mycosynth Gardens",
		Layout:   "normal",
		TypeLine: "Land — Sphere",
		OracleText: "{T}: Add {C}.\n" +
			"{1}, {T}: Add one mana of any color.\n" +
			"{X}, {T}: This land becomes a copy of target nontoken artifact you control with mana value X.",
	}
}

// TestLowerMycosynthGardensBecomeCopyManaValueX proves the X-cost become-a-copy
// ability lowers to a BecomeCopy over a single permanent target whose mana value
// is bound to the ability's chosen X via ManaValueEqualsX (not a fixed
// Selection.ManaValue), restricted to a nontoken artifact the controller
// controls, while the two colorless/any-color mana abilities lower alongside it.
func TestLowerMycosynthGardensBecomeCopyManaValueX(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, theMycosynthGardensCard())

	if len(face.ManaAbilities) != 2 {
		t.Fatalf("mana abilities = %d, want 2 (colorless and any-color)", len(face.ManaAbilities))
	}
	// The first ability is the base colorless mana ability, {T}: Add {C}.
	if !reflect.DeepEqual(face.ManaAbilities[0], game.TapManaAbility(mana.C)) {
		t.Fatalf("mana ability[0] = %#v, want TapManaAbility(mana.C)", face.ManaAbilities[0])
	}
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1 (become-a-copy)", len(face.ActivatedAbilities))
	}

	ability := face.ActivatedAbilities[0]
	if !ability.ManaCost.Exists || ability.ManaCost.Val.ManaValue() != 0 {
		t.Fatalf("ability mana cost = %v, want an {X} cost (mana value 0 off the stack)", ability.ManaCost)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	spec := mode.Targets[0]
	if !spec.ManaValueEqualsX {
		t.Fatal("target spec must set ManaValueEqualsX for the exact \"mana value X\" bound")
	}
	if spec.ManaValueAtMostX {
		t.Fatal("target spec must not set ManaValueAtMostX for the exact bound")
	}
	if spec.Allow != game.TargetAllowPermanent {
		t.Fatalf("target allow = %v, want TargetAllowPermanent", spec.Allow)
	}
	if !spec.Selection.Exists {
		t.Fatal("target selection must exist")
	}
	sel := spec.Selection.Val
	if sel.ManaValue.Exists {
		t.Fatalf("selection must carry no fixed mana value bound, got %+v", sel.ManaValue)
	}
	if !sel.NonToken {
		t.Fatal("selection must require a nontoken target")
	}
	if sel.Controller != game.ControllerYou {
		t.Fatalf("selection controller = %v, want ControllerYou", sel.Controller)
	}
	if len(sel.RequiredTypesAny) != 1 || sel.RequiredTypesAny[0] != types.Artifact {
		t.Fatalf("selection types = %v, want [Artifact]", sel.RequiredTypesAny)
	}
	become, ok := mode.Sequence[0].Primitive.(game.BecomeCopy)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.BecomeCopy", mode.Sequence[0].Primitive)
	}
	if become.Object != game.TargetPermanentReference(0) {
		t.Fatalf("BecomeCopy.Object = %v, want TargetPermanentReference(0)", become.Object)
	}
	if become.UntilEndOfTurn {
		t.Fatal("BecomeCopy must be permanent (no until-end-of-turn duration)")
	}
}

// TestGenerateMycosynthGardensSource proves The Mycosynth Gardens generates
// executable source with no diagnostics and carries the ManaValueEqualsX flag,
// so the whole card is supported rather than failing closed.
func TestGenerateMycosynthGardensSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(theMycosynthGardensCard(), "m")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if !strings.Contains(source, "ManaValueEqualsX: true") {
		t.Fatalf("source missing ManaValueEqualsX flag:\n%s", source)
	}
	// The base colorless mana ability ({T}: Add {C}) and the any-color ability
	// ({1}, {T}: Add one mana of any color) must survive alongside the novel
	// become-a-copy ability.
	if !strings.Contains(source, "game.TapManaAbility(mana.C)") {
		t.Fatalf("source missing base colorless mana ability game.TapManaAbility(mana.C):\n%s", source)
	}
	if !strings.Contains(source, "game.BecomeCopy{") {
		t.Fatalf("source missing BecomeCopy effect:\n%s", source)
	}
}
