package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceDiesTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, draw two cards.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventPermanentDied",
		"game.TriggerSourceSelf",
		"Primitive: game.Draw",
		"game.Fixed(2)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourcePermanentZoneChangeTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wants      []string
	}{
		{
			name:       "self leaves battlefield",
			oracleText: "When this creature leaves the battlefield, draw a card.",
			wants:      []string{"game.EventZoneChanged", "game.TriggerSourceSelf", "MatchFromZone: true", "zone.Battlefield"},
		},
		{
			name:       "batched controlled creatures go to graveyard",
			oracleText: "Whenever one or more creatures you control are put into a graveyard from the battlefield, draw a card.",
			wants:      []string{"game.EventZoneChanged", "OneOrMore:", "zone.Graveyard", "game.TriggerControllerYou"},
		},
		{
			name:       "qualified Human dies",
			oracleText: "Whenever another legendary green Human you control dies, draw a card.",
			wants:      []string{"game.EventPermanentDied", "SubtypesAny: []types.Sub{types.Sub(\"Human\")}", "Supertypes: []types.Super{types.Legendary}", "ColorsAny: []color.Color{color.Green}", "ExcludeSelf:"},
		},
		{
			name:       "self enters or put into graveyard union",
			oracleText: "When this artifact enters or is put into a graveyard from the battlefield, draw a card.",
			wants:      []string{"game.EventPermanentEnteredBattlefield", "game.TriggerSourceSelf", "UnionEvent: game.EventPermanentDied"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.wants {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceSelfDiesDamageTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		OracleText: "When this creature dies, it deals 3 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventPermanentDied",
		"Primitive: game.Damage",
		"DamageSource: opt.Val(game.EventPermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfDiesCounterAbsence(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Undying Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, if it had no +1/+1 counters on it, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`InterveningIf: "if it had no +1/+1 counters on it"`,
		"InterveningIfEventPermanentHadNoCounterKind: opt.Val(counter.PlusOnePlusOne)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfDiesEventCardReturn(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Phoenix",
		Layout:     "normal",
		TypeLine:   "Creature — Phoenix",
		OracleText: "When this creature dies, return it to its owner's hand.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.MoveCard",
		"game.CardReference{Kind: game.CardReferenceEvent}",
		"zone.Graveyard",
		"zone.Hand",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceSelfDiesAdventurePermission(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:   "Test Dreadknight // Test Whispers",
		Layout: "adventure",
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Test Dreadknight",
				ManaCost:   "{1}{G}",
				TypeLine:   "Creature — Human Knight",
				OracleText: "When Test Dreadknight dies, you may cast it from your graveyard as an Adventure until the end of your next turn.",
				Power:      new("2"),
				Toughness:  new("1"),
			},
			{
				Name:       "Test Whispers",
				ManaCost:   "{1}{B}",
				TypeLine:   "Sorcery — Adventure",
				OracleText: "Draw a card.",
			},
		},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Optional: true",
		"Primitive: game.GrantCastPermission",
		"game.CardReference{Kind: game.CardReferenceEvent}",
		"zone.Graveyard",
		"Face:     game.FaceAlternate",
		"Duration: game.DurationUntilEndOfYourNextTurn",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDynamicSelfDiesDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		OracleText: "When this creature dies, it deals damage equal to its power to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DynamicAmountObjectPower",
		"Object:     game.EventPermanentReference()",
		"DamageSource: opt.Val(game.EventPermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDynamicCountDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Swarm",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Test Swarm deals damage equal to twice the number of creatures you control to any target.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DynamicAmountCountSelector",
		"Multiplier: 2",
		"game.BattlefieldGroup",
		"types.Creature",
		"Controller: game.ControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "/counter\"") {
		t.Fatalf("source has unused counter import:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceDynamicSourcePowerCounters(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Druid",
		Layout:     "normal",
		TypeLine:   "Creature — Druid",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "{T}: Put X +1/+1 counters on target creature, where X is Test Druid's power.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DynamicAmountObjectPower",
		"Object:     game.SourcePermanentReference()",
		"CounterKind: counter.PlusOnePlusOne",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceDynamicSourcePowerDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "{T}: Test Devil deals damage equal to Test Devil's power to any target.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DynamicAmountObjectPower",
		"Object:     game.SourcePermanentReference()",
		"DamageSource: opt.Val(game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceRejectsAmbiguousDynamicAmount(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Test Swarm deals damage equal to creatures you control to any target.",
		"You gain X life, where X is opponent.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			}

			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v, want rejection", source, diagnostics)
			}
		})
	}
}

func TestGenerateExecutableCardSourceRejectsDynamicAmountNumberDisagreement(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Draw a card for each creatures you control.",
		"Test Swarm deals damage equal to the number of creature you control to any target.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			}

			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v, want rejection", source, diagnostics)
			}
		})
	}
}

func TestGenerateExecutableCardSourceRejectsAmbiguousDynamicPowerReference(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Druid",
		Layout:     "normal",
		TypeLine:   "Creature — Druid",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "{T}: Put X +1/+1 counters on target creature, where X is its power.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v, want rejection", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceDiesMultipleEffectTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Generous Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, draw a card. You gain 2 life.",
		Power:      new("2"),
		Toughness:  new("2"),
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	draw := strings.Index(source, "Primitive: game.Draw")
	gain := strings.Index(source, "Primitive: game.GainLife")
	if draw < 0 || gain < 0 || draw >= gain {
		t.Fatalf("trigger sequence is not draw then gain life:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceDiesOrExileReanimationAura covers a
// reanimation Aura (Kaya's Ghostform) whose recursion trigger fires on the
// enchanted permanent's "dies or is put into exile" verb disjunction. The
// parser splits the disjunction into independent death and exile triggers, each
// returning the leaving card to the battlefield.
func TestGenerateExecutableCardSourceDiesOrExileReanimationAura(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Test Ghostform",
		Layout:   "normal",
		ManaCost: "{B}",
		TypeLine: "Enchantment — Aura",
		OracleText: "Enchant creature or planeswalker you control\n" +
			"When enchanted permanent dies or is put into exile, return that card to the battlefield under your control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EnchantStaticAbility(&game.TargetSpec{",
		"RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}",
		"Controller: game.ControllerYou",
		"game.EventPermanentDied",
		"game.EventZoneChanged",
		"game.TriggerSourceAttachedPermanent",
		"ToZone:        zone.Exile",
		"Primitive: game.PutOnBattlefield",
		"game.CardReferenceEvent",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	died := strings.Index(source, "game.EventPermanentDied")
	exiled := strings.Index(source, "game.EventZoneChanged")
	if died < 0 || exiled < 0 || died >= exiled {
		t.Fatalf("expected death trigger before exile trigger:\n%s", source)
	}
}
