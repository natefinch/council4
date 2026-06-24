package cardgen

import (
	"strings"
	"testing"
)

func generatedSourceContains(t *testing.T, card *ScryfallCard, wants []string) {
	t.Helper()
	source, diagnostics, err := GenerateExecutableCardSource(card, "cards")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range wants {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// Glint-Horn Buccaneer: an activated draw gated by a source-object combat
// activation condition ("Activate only if this creature is attacking"). The
// condition reference must not leak into the resolving draw body.
func TestGenerateGlintHornBuccaneer(t *testing.T) {
	t.Parallel()
	power, toughness := "2", "4"
	card := &ScryfallCard{
		Name:      "Glint-Horn Buccaneer",
		Layout:    "normal",
		ManaCost:  "{1}{R}{R}",
		TypeLine:  "Creature — Minotaur Pirate",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Haste\n" +
			"Whenever you discard a card, this creature deals 1 damage to each opponent.\n" +
			"{1}{R}, Discard a card: Draw a card. Activate only if this creature is attacking.",
	}
	generatedSourceContains(t, card, []string{
		"game.HasteStaticBody",
		"ActivationCondition: opt.Val(game.Condition{",
		"Object:        opt.Val(game.SourcePermanentReference())",
		"CombatState: game.CombatStateAttacking",
		"Primitive: game.Draw{",
		"Amount: game.Fixed(1)",
		"game.EventCardDiscarded",
	})
}

// Decaying Time Loop: the whole-hand discard-then-draw-that-many sequence,
// published as a result the draw reads back, plus Retrace.
func TestGenerateDecayingTimeLoop(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Decaying Time Loop",
		Layout:   "normal",
		ManaCost: "{3}{R}",
		TypeLine: "Instant",
		OracleText: "Discard all the cards in your hand, then draw that many cards.\n" +
			"Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)",
	}
	generatedSourceContains(t, card, []string{
		"game.RetraceStaticBody",
		"Primitive: game.Discard{",
		"EntireHand: true",
		"PublishResult: game.ResultKey(\"discarded-this-way\")",
		"Kind:      game.DynamicAmountPreviousEffectResult",
		"ResultKey: game.ResultKey(\"discarded-this-way\")",
	})
}

// Kitsa, Otterball Elite: copy a controlled instant or sorcery spell with the
// optional new-targets rider, gated by a named-self power activation condition.
func TestGenerateKitsaOtterballElite(t *testing.T) {
	t.Parallel()
	power, toughness := "1", "3"
	card := &ScryfallCard{
		Name:      "Kitsa, Otterball Elite",
		Layout:    "normal",
		ManaCost:  "{1}{U}",
		TypeLine:  "Legendary Creature — Otter Wizard",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Vigilance\n" +
			"Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)\n" +
			"{T}: Draw a card, then discard a card.\n" +
			"{2}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy. Activate only if Kitsa's power is 3 or greater.",
	}
	generatedSourceContains(t, card, []string{
		"SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery}",
		"StackObjectKinds:  []game.StackObjectKind{game.StackSpell}",
		"Controller:        game.ControllerYou",
		"Primitive: game.CopyStackObject{",
		"Object:              game.TargetStackObjectReference(0)",
		"MayChooseNewTargets: true",
		"Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})",
	})
}

// Vnwxt, Verbose Host: the Max-speed-gated draw-doubling replacement.
func TestGenerateVnwxtVerboseHost(t *testing.T) {
	t.Parallel()
	power, toughness := "0", "4"
	card := &ScryfallCard{
		Name:      "Vnwxt, Verbose Host",
		Layout:    "normal",
		ManaCost:  "{1}{U}",
		TypeLine:  "Legendary Creature — Homunculus",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)\n" +
			"You have no maximum hand size.\n" +
			"Max speed — If you would draw a card, draw two cards instead.",
	}
	generatedSourceContains(t, card, []string{
		"game.NoMaximumHandSizeStaticBody",
		"game.MaxSpeedDrawCardMultiplierReplacement(",
	})
}
