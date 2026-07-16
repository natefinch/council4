package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

const aberrantOracleText = "Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)\n" +
	"Trample\n" +
	"Heavy Power Hammer — Whenever this creature deals combat damage to a player, destroy target artifact or enchantment that player controls."

func aberrantCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Aberrant",
		Layout:     "normal",
		ManaCost:   "{X}{1}{G}",
		TypeLine:   "Creature — Tyranid Mutant",
		Power:      new("0"),
		Toughness:  new("0"),
		OracleText: aberrantOracleText,
	}
}

func TestLowerAberrantReusableMechanics(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, aberrantCard())
	if len(face.ReplacementAbilities) != 1 ||
		face.ReplacementAbilities[0].Replacement.EntersWithCounters[0].Kind != counter.PlusOnePlusOne ||
		!face.ReplacementAbilities[0].Replacement.EntersWithCounters[0].AmountFromX {
		t.Fatalf("replacement abilities = %#v; want Ravenous X counters", face.ReplacementAbilities)
	}
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d; want Ravenous draw and combat-damage destroy", len(face.TriggeredAbilities))
	}
	if !game.BodyHasKeyword(&face.TriggeredAbilities[0], game.Ravenous) {
		t.Fatalf("first trigger = %#v; want Ravenous keyword", face.TriggeredAbilities[0])
	}
	destroy := face.TriggeredAbilities[1]
	if destroy.Trigger.Pattern.Event != game.EventDamageDealt ||
		!destroy.Trigger.Pattern.RequireCombatDamage ||
		destroy.Trigger.Pattern.DamageRecipient != game.DamageRecipientPlayer {
		t.Fatalf("destroy trigger pattern = %#v", destroy.Trigger.Pattern)
	}
	mode := destroy.Content.Modes[0]
	if len(mode.Targets) != 1 || !mode.Targets[0].Selection.Exists {
		t.Fatalf("destroy targets = %#v; want one selected target", mode.Targets)
	}
	selection := mode.Targets[0].Selection.Val
	if !selection.ControlledByEventPlayer ||
		!slices.Equal(selection.RequiredTypesAny, []types.Card{types.Artifact, types.Enchantment}) {
		t.Fatalf("destroy selection = %#v; want damaged player's artifact or enchantment", selection)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("destroy primitive = %#v", mode.Sequence[0].Primitive)
	}
}

func TestGenerateExecutableAberrantSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(aberrantCard(), "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.RavenousEntersWithCountersReplacement()",
		"game.RavenousDrawTriggeredAbility()",
		"ControlledByEventPlayer: true",
		"RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}",
		"Primitive: game.Destroy{",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestRavenousNearMissFailsClosed(t *testing.T) {
	t.Parallel()
	card := aberrantCard()
	card.Name = "Malformed Ravenous"
	card.OracleText = "Ravenous 2 (This creature enters with X +1/+1 counters on it.)"
	face := lowerSingleFaceExpectingUnsupported(t, card)
	if len(face.ReplacementAbilities) != 0 || len(face.TriggeredAbilities) != 0 {
		t.Fatalf("near-miss Ravenous produced partial abilities: replacements=%#v triggers=%#v",
			face.ReplacementAbilities, face.TriggeredAbilities)
	}
}
