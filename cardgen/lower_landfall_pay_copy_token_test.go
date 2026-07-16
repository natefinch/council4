package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// springheartNantukoCard is the real Springheart Nantuko card: a fixed-mana
// Bestow enchantment creature whose landfall trigger offers an optional {1}{G}
// payment gated on being attached to a creature its controller controls, arms a
// reflexive trigger that copies the still-attached creature, and otherwise
// creates a fixed 1/1 green Insect.
func springheartNantukoCard() *ScryfallCard {
	return &ScryfallCard{
		Name:      "Springheart Nantuko",
		Layout:    "normal",
		TypeLine:  "Enchantment Creature — Insect Monk",
		ManaCost:  "{1}{G}",
		Power:     new("1"),
		Toughness: new("1"),
		OracleText: "Bestow {1}{G}\n" +
			"Enchanted creature gets +1/+1.\n" +
			"Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature. If you didn't create a token this way, create a 1/1 green Insect creature token.",
	}
}

// TestLowerLandfallPayCopyToken proves Springheart Nantuko's landfall body lowers
// to the composed reflexive-payment shape: an optional gated {1}{G} payment, a
// reflexive trigger armed by paying that copies the still-attached creature or
// falls back to a fixed Insect, plus the declined-payment and unattached Insect
// fallbacks. Every path that makes no copy creates exactly one Insect.
func TestLowerLandfallPayCopyToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, springheartNantukoCard())

	// Bestow and the +1/+1 grant lower as the two static abilities; the landfall
	// body is the sole triggered ability.
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %d, want 2 (Bestow + grant)", len(face.StaticAbilities))
	}
	if _, ok := game.StaticBodyBestow(&face.StaticAbilities[0].Body); !ok {
		t.Fatalf("first static ability is not Bestow: %#v", face.StaticAbilities[0].Body)
	}
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]

	// Landfall is "Whenever a land you control enters" — a land ETB event.
	if ability.Trigger.Type != game.TriggerWhenever {
		t.Fatalf("trigger type = %v, want TriggerWhenever", ability.Trigger.Type)
	}
	if ability.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("trigger event = %v, want EventPermanentEnteredBattlefield", ability.Trigger.Pattern.Event)
	}
	if ability.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger controller = %v, want TriggerControllerYou", ability.Trigger.Pattern.Controller)
	}
	if rt := ability.Trigger.Pattern.SubjectSelection.RequiredTypes; len(rt) != 1 || rt[0] != types.Land {
		t.Fatalf("trigger subject types = %v, want [Land]", rt)
	}

	seq := ability.Content.Modes[0].Sequence
	if len(seq) != 4 {
		t.Fatalf("landfall sequence length = %d, want 4", len(seq))
	}

	// [0] Optional payment, gated on being attached to a creature you control,
	// publishing the paid result.
	pay, ok := seq[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("seq[0] = %T, want game.Pay", seq[0].Primitive)
	}
	if !pay.Payment.ManaCost.Exists {
		t.Fatalf("payment has no mana cost: %#v", pay.Payment)
	}
	if seq[0].PublishResult != game.ResultKey("springheart-landfall-paid") {
		t.Fatalf("seq[0] publish = %q, want springheart-landfall-paid", seq[0].PublishResult)
	}
	assertAttachedToControlledCreature(t, "seq[0].Condition", seq[0].Condition, false)

	// [1] Reflexive trigger, armed only when the payment succeeded.
	reflexive, ok := seq[1].Primitive.(game.CreateReflexiveTrigger)
	if !ok {
		t.Fatalf("seq[1] = %T, want game.CreateReflexiveTrigger", seq[1].Primitive)
	}
	assertResultGate(t, "seq[1].ResultGate", seq[1].ResultGate, "springheart-landfall-paid", game.TriTrue)

	rseq := reflexive.Trigger.Content.Modes[0].Sequence
	if len(rseq) != 2 {
		t.Fatalf("reflexive sequence length = %d, want 2", len(rseq))
	}
	// [1.0] Copy the still-attached creature, publishing the copied result. No
	// condition/gate: an unresolvable attached-permanent reference yields no
	// token and a false result, arming the fallback.
	copyTok, ok := rseq[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("reflexive[0] = %T, want game.CreateToken", rseq[0].Primitive)
	}
	copySpec, ok := copyTok.Source.TokenCopy()
	if !ok || copySpec.Source != game.TokenCopySourceObject {
		t.Fatalf("reflexive[0] is not a token copy of an object: %#v", copyTok.Source)
	}
	if copySpec.Object != game.SourceAttachedPermanentReference() {
		t.Fatalf("reflexive[0] copies %#v, want SourceAttachedPermanentReference", copySpec.Object)
	}
	if rseq[0].PublishResult != game.ResultKey("springheart-landfall-copied") {
		t.Fatalf("reflexive[0] publish = %q, want springheart-landfall-copied", rseq[0].PublishResult)
	}
	// [1.1] Insect fallback when no copy was made (paid but detached before payoff).
	assertInsectToken(t, "reflexive[1]", rseq[1].Primitive)
	assertResultGate(t, "reflexive[1].ResultGate", rseq[1].ResultGate, "springheart-landfall-copied", game.TriFalse)

	// [2] Declined-payment Insect: attached, offered, but not paid.
	assertInsectToken(t, "seq[2]", seq[2].Primitive)
	assertResultGate(t, "seq[2].ResultGate", seq[2].ResultGate, "springheart-landfall-paid", game.TriFalse)

	// [3] Unattached Insect: no payment offered because the source is not attached
	// to a creature you control (negated condition).
	assertInsectToken(t, "seq[3]", seq[3].Primitive)
	assertAttachedToControlledCreature(t, "seq[3].Condition", seq[3].Condition, true)
}

// TestGenerateExecutableCardSourceLandfallPayCopyToken proves the committed Go
// source reconstructs the reflexive payment/copy/Insect composition and the
// fixed 1/1 green Insect token definition deterministically.
func TestGenerateExecutableCardSourceLandfallPayCopyToken(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(springheartNantukoCard(), "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.CreateReflexiveTrigger{",
		"game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceObject,",
		"Object: game.SourceAttachedPermanentReference(),",
		`game.ResultKey("springheart-landfall-paid")`,
		`Key:       "springheart-landfall-copied",`,
		"Subtypes:  []types.Sub{types.Insect},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestLandfallPayCopyTokenNearMissFailsClosed proves each near-miss of the exact
// landfall body is not recognized and fails closed (diagnostics, no partial
// ability), so only Springheart Nantuko's precise wording composes this shape.
func TestLandfallPayCopyTokenNearMissFailsClosed(t *testing.T) {
	t.Parallel()
	const prefix = "Bestow {1}{G}\nEnchanted creature gets +1/+1.\n"
	cases := []struct {
		name string
		body string
	}{
		{
			name: "different payment cost",
			body: "Landfall — Whenever a land you control enters, you may pay {2}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature. If you didn't create a token this way, create a 1/1 green Insect creature token.",
		},
		{
			name: "different token subtype",
			body: "Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature. If you didn't create a token this way, create a 1/1 green Beetle creature token.",
		},
		{
			name: "different token color",
			body: "Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature. If you didn't create a token this way, create a 1/1 white Insect creature token.",
		},
		{
			name: "different token size",
			body: "Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature. If you didn't create a token this way, create a 2/2 green Insect creature token.",
		},
		{
			name: "missing fixed-Insect fallback branch",
			body: "Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that creature.",
		},
		{
			name: "copies a different object",
			body: "Landfall — Whenever a land you control enters, you may pay {1}{G} if this permanent is attached to a creature you control. If you do, create a token that's a copy of that land. If you didn't create a token this way, create a 1/1 green Insect creature token.",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Near Miss Nantuko",
				Layout:     "normal",
				TypeLine:   "Enchantment Creature — Insect Monk",
				ManaCost:   "{1}{G}",
				Power:      new("1"),
				Toughness:  new("1"),
				OracleText: prefix + test.body,
			})
		})
	}
}

func assertInsectToken(t *testing.T, label string, primitive game.Primitive) {
	t.Helper()
	tok, ok := primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("%s = %T, want game.CreateToken", label, primitive)
	}
	def, ok := tok.Source.TokenDefRef()
	if !ok {
		t.Fatalf("%s is not a fixed token definition: %#v", label, tok.Source)
	}
	if def.Name != "Insect" {
		t.Fatalf("%s token name = %q, want Insect", label, def.Name)
	}
	if !def.Power.Exists || def.Power.Val.Value != 1 || !def.Toughness.Exists || def.Toughness.Val.Value != 1 {
		t.Fatalf("%s token is not 1/1: power=%v toughness=%v", label, def.Power, def.Toughness)
	}
	if len(def.Subtypes) != 1 || def.Subtypes[0] != types.Insect {
		t.Fatalf("%s token subtypes = %v, want [Insect]", label, def.Subtypes)
	}
}

func assertResultGate(t *testing.T, label string, gate opt.V[game.InstructionResultGate], key game.ResultKey, want game.TriState) {
	t.Helper()
	if !gate.Exists {
		t.Fatalf("%s missing result gate", label)
	}
	if gate.Val.Key != key {
		t.Fatalf("%s gate key = %q, want %q", label, gate.Val.Key, key)
	}
	if gate.Val.Succeeded != want {
		t.Fatalf("%s gate succeeded = %v, want %v", label, gate.Val.Succeeded, want)
	}
}

func assertAttachedToControlledCreature(t *testing.T, label string, cond opt.V[game.EffectCondition], negated bool) {
	t.Helper()
	if !cond.Exists || !cond.Val.Condition.Exists {
		t.Fatalf("%s missing attachment condition", label)
	}
	c := cond.Val.Condition.Val
	if c.Negate != negated {
		t.Fatalf("%s negate = %v, want %v", label, c.Negate, negated)
	}
	if !c.Object.Exists || c.Object.Val != game.SourceAttachedPermanentReference() {
		t.Fatalf("%s object = %#v, want SourceAttachedPermanentReference", label, c.Object)
	}
	if !c.ObjectMatches.Exists {
		t.Fatalf("%s missing ObjectMatches selection", label)
	}
	sel := c.ObjectMatches.Val
	if sel.Controller != game.ControllerYou {
		t.Fatalf("%s selection controller = %v, want ControllerYou", label, sel.Controller)
	}
	if len(sel.RequiredTypes) != 1 || sel.RequiredTypes[0] != types.Creature {
		t.Fatalf("%s selection required types = %v, want [Creature]", label, sel.RequiredTypes)
	}
}
