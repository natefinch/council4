package cardgen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerFixedCounterSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Put two +1/+1 counters on target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("targets = %+v, want one creature target", mode.Targets)
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok ||
		add.Amount.Value() != 2 ||
		add.CounterKind != counter.PlusOnePlusOne ||
		add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want two +1/+1 counters on target 0", mode.Sequence[0].Primitive)
	}
}

func TestLowerNamedCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind counter.Kind
	}{
		{"Put a charge counter on target artifact.", counter.Charge},
		{"Put two shield counters on target creature you control.", counter.Shield},
		{"Put a first strike counter on target creature.", counter.FirstStrike},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.text,
			})
			add, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok || add.CounterKind != test.kind {
				t.Fatalf("primitive = %+v", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
			}
		})
	}
}

func TestLowerKeywordNamedCounterPlacementAbilityShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		card    *ScryfallCard
		content func(loweredFaceAbilities) (game.AbilityContent, bool)
		kind    counter.Kind
	}{
		{
			name: "activated",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: "{T}: Put a flying counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.ActivatedAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.ActivatedAbilities[0].Content, true
			},
			kind: counter.Flying,
		},
		{
			name: "loyalty",
			card: &ScryfallCard{
				Name:       "Test Walker",
				Layout:     "normal",
				TypeLine:   "Legendary Planeswalker — Test",
				OracleText: "+1: Put a lifelink counter on target creature.",
				Loyalty:    func() *string { loyalty := "3"; return &loyalty }(),
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.LoyaltyAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.LoyaltyAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
		{
			name: "triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Creature — Test",
				OracleText: "When this creature enters, put a first strike counter on target creature.",
				Power:      new("2"),
				Toughness:  new("2"),
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.FirstStrike,
		},
		{
			name: "phase triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "At the beginning of your upkeep, put a flying counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.Flying,
		},
		{
			name: "non-self enter triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "Whenever another creature enters, put a lifelink counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
		{
			name: "ordered effects",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Put a flying counter on target creature. Draw a card.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if !face.SpellAbility.Exists {
					return game.AbilityContent{}, false
				}
				return face.SpellAbility.Val, true
			},
			kind: counter.Flying,
		},
		{
			name: "Saga chapter",
			card: &ScryfallCard{
				Name:       "Test Saga",
				Layout:     "saga",
				TypeLine:   "Enchantment — Saga",
				OracleText: "I — Put a lifelink counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.ChapterAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.ChapterAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, test.card)
			content, ok := test.content(face)
			if !ok ||
				len(content.Modes) != 1 ||
				len(content.Modes[0].Sequence) == 0 {
				t.Fatalf("face = %+v, want lowered counter placement", face)
			}
			add, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok || add.CounterKind != test.kind {
				t.Fatalf("primitive = %+v, want %s counter placement", content.Modes[0].Sequence[0].Primitive, test.kind)
			}
		})
	}
}

func TestLowerPlayerCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind counter.Kind
	}{
		{"Put an energy counter on target player.", counter.Energy},
		{"Put two experience counters on target player.", counter.Experience},
		{"Put three poison counters on target opponent.", counter.Poison},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.text,
			})
			mode := face.SpellAbility.Val.Modes[0]
			add, ok := mode.Sequence[0].Primitive.(game.AddPlayerCounter)
			if !ok ||
				add.CounterKind != test.kind ||
				add.Player != game.TargetPlayerReference(0) ||
				mode.Targets[0].Allow != game.TargetAllowPlayer {
				t.Fatalf("mode = %+v", mode)
			}
		})
	}
}

func TestLowerEveryRecognizedCounterKindOnItsValidTarget(t *testing.T) {
	t.Parallel()
	for kind := counter.PlusOnePlusOne; kind <= counter.Experience; kind++ {
		if kind == counter.Stun || kind == counter.Finality {
			continue
		}
		if !compiler.CounterKindPlacementSupported(kind) {
			t.Fatalf("%s unexpectedly excluded from named placement", kind)
		}
		name := kind.String()
		article := "a"
		if strings.ContainsRune("aeiou", rune(name[0])) {
			article = "an"
		}
		target := "target permanent"
		if kind.PlayerOnly() {
			target = "target player"
		}
		oracleText := fmt.Sprintf("Put %s %s counter on %s.", article, name, target)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			})
			primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
			if kind.PlayerOnly() {
				add, ok := primitive.(game.AddPlayerCounter)
				if !ok || add.CounterKind != kind {
					t.Fatalf("primitive = %+v", primitive)
				}
				return
			}
			add, ok := primitive.(game.AddCounter)
			if !ok || add.CounterKind != kind {
				t.Fatalf("primitive = %+v", primitive)
			}
		})
	}
}

func TestLowerCounterPlacementRejectsMissingRuntimeMechanics(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		kind counter.Kind
	}{
		{"stun", counter.Stun},
		{"finality", counter.Finality},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := contentCtx{
				text: "Put a " + test.name + " counter on target creature.",
				content: compiler.AbilityContent{
					Targets: []compiler.CompiledTarget{{
						Text:        "target creature",
						Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
						Selector:    compiler.CompiledSelector{Kind: compiler.SelectorCreature},
					}},
					Effects: []compiler.CompiledEffect{{
						Kind:             compiler.EffectPut,
						Amount:           compiler.CompiledAmount{Value: 1, Known: true},
						CounterKind:      test.kind,
						CounterKindKnown: true,
					}},
				},
			}
			if _, diagnostic := lowerCounterPlacementSpell(ctx); diagnostic == nil {
				t.Fatal("lowering accepted counter kind without runtime mechanics")
			}
		})
	}
}

func TestLowerDynamicNamedCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind game.DynamicAmountKind
	}{
		{"Put X charge counters on target artifact.", game.DynamicAmountX},
		{"Put X poison counters on target player, where X is the number of lands you control.", game.DynamicAmountCountSelector},
		{"Put X energy counters on target player, where X is Test Counter's power.", game.DynamicAmountObjectPower},
	}
	for _, test := range tests {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Counter",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: test.text,
		})
		primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
		amount, ok := counterPlacementAmount(primitive)
		if !ok {
			t.Fatalf("%q primitive = %T", test.text, primitive)
		}
		dynamic := amount.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != test.kind {
			t.Fatalf("%q amount = %+v", test.text, dynamic)
		}
		if test.kind == game.DynamicAmountObjectPower &&
			dynamic.Val.Object != game.SourcePermanentReference() {
			t.Fatalf("%q source reference = %+v", test.text, dynamic.Val.Object)
		}
	}

}

func TestRebaseAddPlayerCounterTargetReference(t *testing.T) {
	t.Parallel()
	primitive, ok := rebaseTargetedPrimitive(game.AddPlayerCounter{
		Amount:      game.Fixed(1),
		Player:      game.TargetPlayerReference(0),
		CounterKind: counter.Poison,
	}, 2)
	if !ok {
		t.Fatal("AddPlayerCounter target was not rebased")
	}
	add, ok := primitive.(game.AddPlayerCounter)
	if !ok || add.Player != game.TargetPlayerReference(2) {
		t.Fatalf("rebased primitive = %+v", primitive)
	}
}

func TestLowerCounterPlacementRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Put a quest counter on target permanent.",
		"Put an energy counter on target creature.",
		"Put a charge counter on target player.",
		"Put a charge counter on any target.",
		"Put a +1/+1 counter on each creature you control.",
		"Put a charge and time counter on target artifact.",
		"Put 0 charge counters on target artifact.",
		"Put -1 charge counters on target artifact.",
		"Put a charge counter on target artifact for each land you control.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Counter",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: oracleText,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without diagnostics", oracleText)
		}
	}
}

func TestLowerRegenerateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Regenerate",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Regenerate target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	regenerate, ok := mode.Sequence[0].Primitive.(game.Regenerate)
	if !ok || regenerate.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want regenerate target permanent", mode.Sequence[0].Primitive)
	}
}

func TestLowerFightSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fight",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control fights target creature you don't control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %+v, want two creatures", mode.Targets)
	}
	fight, ok := mode.Sequence[0].Primitive.(game.Fight)
	if !ok ||
		fight.Object != game.TargetPermanentReference(0) ||
		fight.RelatedObject != game.TargetPermanentReference(1) {
		t.Fatalf("primitive = %+v, want targets 0 and 1 fight", mode.Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityPositiveCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "+1: Draw a card.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != 1 {
		t.Fatalf("LoyaltyCost = %d, want 1", la.LoyaltyCost)
	}
	if la.Content.IsModal() || len(la.Content.Modes) != 1 {
		t.Fatalf("content = %+v, want single non-modal mode", la.Content)
	}
	draw, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("primitive = %+v, want controller draws one", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityNegativeCost(t *testing.T) {
	t.Parallel()
	loyaltyText := "\u22122: Target player mills three cards."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: loyaltyText,
		Loyalty:    func() *string { s := "4"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != -2 {
		t.Fatalf("LoyaltyCost = %d, want -2", la.LoyaltyCost)
	}
	mill, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Mill)
	if !ok || mill.Amount.Value() != 3 {
		t.Fatalf("primitive = %+v, want mills three", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityZeroCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "0: Scry 2.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != 0 {
		t.Fatalf("LoyaltyCost = %d, want 0", la.LoyaltyCost)
	}
	scry, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Scry)
	if !ok || scry.Amount.Value() != 2 {
		t.Fatalf("primitive = %+v, want scry two", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityMultiple(t *testing.T) {
	t.Parallel()
	oracleText := "+1: Draw a card.\n\u22122: You lose 3 life."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: oracleText,
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 2 {
		t.Fatalf("got %d loyalty abilities, want 2", len(face.LoyaltyAbilities))
	}
	if face.LoyaltyAbilities[0].LoyaltyCost != 1 {
		t.Fatalf("first LoyaltyCost = %d, want 1", face.LoyaltyAbilities[0].LoyaltyCost)
	}
	if face.LoyaltyAbilities[1].LoyaltyCost != -2 {
		t.Fatalf("second LoyaltyCost = %d, want -2", face.LoyaltyAbilities[1].LoyaltyCost)
	}
}

func TestLowerLoyaltyAbilityVariableCostRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "\u2212X: Target player mills X cards.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for variable loyalty cost, got none")
	}
}

func TestLowerModalChooseOneSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability")
	}
	content := face.SpellAbility.Val
	if !content.IsModal() {
		t.Fatal("spell ability is not modal")
	}
	if len(content.Modes) != 2 {
		t.Fatalf("got %d modes, want 2", len(content.Modes))
	}
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("MinModes=%d MaxModes=%d, want both 1", content.MinModes, content.MaxModes)
	}
	draw, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 {
		t.Fatalf("mode 0 primitive = %+v, want draw one", content.Modes[0].Sequence[0].Primitive)
	}
	gain, ok := content.Modes[1].Sequence[0].Primitive.(game.GainLife)
	if !ok || gain.Amount.Value() != 3 {
		t.Fatalf("mode 1 primitive = %+v, want gain 3 life", content.Modes[1].Sequence[0].Primitive)
	}
}

func TestLowerModalChooseOneWithTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Destroy target artifact.\n\u2022 Draw a card.",
	})
	content := face.SpellAbility.Val
	if !content.IsModal() || len(content.Modes) != 2 {
		t.Fatalf("content = %+v, want modal with 2 modes", content)
	}
	if len(content.Modes[0].Targets) != 1 {
		t.Fatalf("mode 0 targets = %+v, want one target", content.Modes[0].Targets)
	}
	if _, ok := content.Modes[0].Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("mode 0 primitive = %T, want game.Destroy", content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerModalChooseTwoSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Command",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Choose two \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.\n\u2022 Proliferate.",
	})
	content := face.SpellAbility.Val
	if content.MinModes != 2 || content.MaxModes != 2 {
		t.Fatalf("MinModes=%d MaxModes=%d, want both 2", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 3 {
		t.Fatalf("got %d modes, want 3", len(content.Modes))
	}
}

func TestLowerModalChooseOneOrBoth(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one or both \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	content := face.SpellAbility.Val
	if content.MinModes != 1 || content.MaxModes != 2 {
		t.Fatalf("MinModes=%d MaxModes=%d, want 1 and 2", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 2 {
		t.Fatalf("got %d modes, want 2", len(content.Modes))
	}
}

func TestLowerModalChoiceCountExceedsModesRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Command",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Choose three \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics when choice count exceeds modes, got none")
	}
}

func TestLowerModalUnsupportedModeBodyRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Draw a card.\n\u2022 Search your library for a card.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported mode body, got none")
	}
}
