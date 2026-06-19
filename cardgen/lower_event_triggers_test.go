package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerGenericModalActivatedAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Console",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{1}, Discard a card: Choose one —\n• Draw a card.\n• You gain 3 life.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if !ability.ManaCost.Exists || len(ability.ManaCost.Val) != 1 {
		t.Fatalf("mana cost = %#v, want {1}", ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalDiscard {
		t.Fatalf("additional costs = %#v, want discard", ability.AdditionalCosts)
	}
	if !ability.Content.IsModal() || ability.Content.MinModes != 1 || ability.Content.MaxModes != 1 || len(ability.Content.Modes) != 2 {
		t.Fatalf("content = %#v, want choose-one modal content", ability.Content)
	}
	if _, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("first mode primitive = %T, want game.Draw", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	if _, ok := ability.Content.Modes[1].Sequence[0].Primitive.(game.GainLife); !ok {
		t.Fatalf("second mode primitive = %T, want game.GainLife", ability.Content.Modes[1].Sequence[0].Primitive)
	}
}

func TestPrepareModalActivationCondition(t *testing.T) {
	t.Parallel()
	ability := compiler.CompiledAbility{
		Content: compiler.AbilityContent{
			Modes: []compiler.CompiledMode{{Content: compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{{
					Kind: compiler.EffectDraw,
					Span: shared.Span{
						Start: shared.Position{Offset: 10},
						End:   shared.Position{Offset: 20},
					},
				}},
			}}},
			Conditions: []compiler.CompiledCondition{{
				Kind:      compiler.ConditionOnlyIf,
				Text:      "only if you have no cards in hand",
				Predicate: compiler.ConditionPredicateControllerHandEmpty,
				Span: shared.Span{
					Start: shared.Position{Offset: 30},
					End:   shared.Position{Offset: 40},
				},
			}},
		},
	}
	syntax := parser.Ability{}
	condition, ok := prepareActivationCondition(&ability, &syntax)
	if !ok || !condition.Exists || !condition.Val.ControllerHandEmpty {
		t.Fatalf("condition = %#v, ok = %v, want modal activation condition", condition, ok)
	}
	if len(ability.Content.Conditions) != 0 {
		t.Fatalf("remaining conditions = %#v, want none", ability.Content.Conditions)
	}
}

func TestLowerActivatedAbilityEventHistoryCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Raider",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{1}: Draw a card. Activate only if you attacked this turn.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.ActivatedAbilities[0]
	if !ability.ActivationCondition.Exists {
		t.Fatalf("activation condition = %#v, want present", ability.ActivationCondition)
	}
	history := ability.ActivationCondition.Val.EventHistory
	if !history.Exists || history.Val.Window != game.EventHistoryCurrentTurn {
		t.Fatalf("event history = %#v, want current-turn pattern", history)
	}
}

func TestLowerActivatedAbilityGraveyardEventHistoryConditionFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Revenant",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{1}{B}: Return this card from your graveyard to the battlefield. Activate only if you attacked this turn.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(diagnostics, func(diagnostic shared.Diagnostic) bool {
		return diagnostic.Summary == "unsupported activation condition"
	}) {
		t.Fatalf("diagnostics = %#v, want unsupported activation condition for graveyard event-history ability", diagnostics)
	}
}

func TestLowerActivatedAbilityComposesCostTimingAndCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Console",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{1}, {T}, Pay 2 life: Draw a card. Activate only if you control an artifact. Activate only as a sorcery.",
	})
	ability := face.ActivatedAbilities[0]
	if ability.Timing != game.SorceryOnly || !ability.ActivationCondition.Exists {
		t.Fatalf("timing/condition = %v/%#v, want sorcery and condition", ability.Timing, ability.ActivationCondition)
	}
	if len(ability.AdditionalCosts) != 2 ||
		ability.AdditionalCosts[0].Kind != cost.AdditionalTap ||
		ability.AdditionalCosts[1].Kind != cost.AdditionalPayLife {
		t.Fatalf("additional costs = %#v, want printed tap then pay-life order", ability.AdditionalCosts)
	}
}

func TestActivatedAbilityCapabilityDiagnostics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		summary    string
	}{
		{name: "cost", oracleText: "Exile a card: Draw a card.", summary: "unsupported activation cost"},
		{name: "timing", oracleText: "{1}: Draw a card. Activate only during your end step.", summary: "unsupported activation timing"},
		{name: "unrecognized timing grammar", oracleText: "{1}: Draw a card. Activate only before combat.", summary: "unsupported activation timing"},
		{name: "opponent turn timing", oracleText: "{1}: Draw a card. Activate only during an opponent's turn.", summary: "unsupported activation timing"},
		{name: "condition", oracleText: "{1}: Draw a card. Activate only if you have one or fewer cards in hand.", summary: "unsupported activation condition"},
		{name: "references", oracleText: "{1}: It deals 1 damage to any target.", summary: "unsupported activation references"},
		{name: "ambiguous cost references", oracleText: "Put a +1/+1 counter on them: Draw a card.", summary: "unsupported activation references"},
		{name: "cost reference to prior object", oracleText: "Tap an untapped creature you control, Remove a +1/+1 counter from it: Draw a card.", summary: "unsupported activation references"},
		{name: "cost reference after source and prior object", oracleText: "Remove a charge counter from this artifact, Tap an untapped creature you control, Remove a +1/+1 counter from it: Draw a card.", summary: "unsupported activation references"},
		{name: "modes", oracleText: "{1}: Choose any number —\n• Draw a card.\n• You gain 3 life.", summary: "unsupported activation modes"},
		{name: "partially understood mode", oracleText: "{1}: Choose one —\n• Gain control of target creature until end of turn. The Ring tempts you.\n• You gain 3 life.", summary: "unsupported gain-control spell"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Console",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if !slices.ContainsFunc(diagnostics, func(diagnostic shared.Diagnostic) bool {
				return diagnostic.Summary == test.summary
			}) {
				t.Fatalf("diagnostics = %#v, want %q", diagnostics, test.summary)
			}
		})
	}
}

func TestActivatedAbilityZoneDiagnostic(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileTestOracle("{1}: Draw a card.", parser.Context{}, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	ability.ActivationZone = zone.Hand
	_, diagnostic := lowerActivationShell("", ability, &compilation.Syntax.Abilities[0])
	if diagnostic == nil || diagnostic.Summary != "unsupported activation zone" {
		t.Fatalf("diagnostic = %#v, want unsupported activation zone", diagnostic)
	}
}

func TestSemanticManaAbilityRequiresNoTargets(t *testing.T) {
	t.Parallel()
	untargeted, diagnostics := compileTestOracle("{T}: Add {G}.", parser.Context{}, compiler.Context{})
	if len(diagnostics) != 0 || !isSemanticManaAbility(untargeted.Abilities[0]) {
		t.Fatalf("untargeted add-mana ability classification = false, diagnostics %#v", diagnostics)
	}
	targeted, diagnostics := compileTestOracle("{T}: Target player adds {G}.", parser.Context{}, compiler.Context{})
	if len(diagnostics) != 0 || isSemanticManaAbility(targeted.Abilities[0]) {
		t.Fatalf("targeted add-mana ability classification = true, diagnostics %#v", diagnostics)
	}
}

func TestLowerAddManaThroughSharedAbilityContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		content    func(loweredFaceAbilities) game.AbilityContent
	}{
		{
			name:       "spell",
			typeLine:   "Instant",
			oracleText: "Add {B}{B}{B}.",
			content:    func(face loweredFaceAbilities) game.AbilityContent { return face.SpellAbility.Val },
		},
		{
			name:       "trigger",
			typeLine:   "Creature — Goblin",
			oracleText: "When this creature enters, add {R}.",
			content:    func(face loweredFaceAbilities) game.AbilityContent { return face.TriggeredAbilities[0].Content },
		},
		{
			name:       "mana ability",
			typeLine:   "Land",
			oracleText: "{T}: Add {G}.",
			content:    func(face loweredFaceAbilities) game.AbilityContent { return face.ManaAbilities[0].Content },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{Name: "Test Card", Layout: "normal", TypeLine: test.typeLine, OracleText: test.oracleText})
			content := test.content(face)
			if len(content.Modes) != 1 || len(content.Modes[0].Sequence) == 0 {
				t.Fatalf("content = %#v, want add-mana sequence", content)
			}
			for _, instruction := range content.Modes[0].Sequence {
				if _, ok := instruction.Primitive.(game.AddMana); !ok {
					t.Fatalf("primitive = %T, want game.AddMana", instruction.Primitive)
				}
			}
		})
	}
}

func TestLowerEventPlayerDrawInTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		oracle  string
		wantAmt int
	}{
		{"they draw a card", "Whenever an opponent draws a card, they draw a card.", 1},
		{"they draw 2 cards", "Whenever a player draws a card, they draw 2 cards.", 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			draw, ok := mode.Sequence[0].Primitive.(game.Draw)
			if !ok {
				t.Fatalf("primitive = %T, want game.Draw", mode.Sequence[0].Primitive)
			}
			if draw.Amount.Value() != tc.wantAmt {
				t.Errorf("amount = %v, want %d", draw.Amount, tc.wantAmt)
			}
			if draw.Player != game.EventPlayerReference() {
				t.Errorf("player = %v, want EventPlayerReference", draw.Player)
			}
		})
	}
}

func TestLowerEventPlayerDiscardInTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		oracle  string
		wantAmt int
	}{
		{"they discard a card", "Whenever a player discards a card, they discard a card.", 1},
		{"they discard 2 cards", "Whenever a player draws a card, they discard 2 cards.", 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			discard, ok := mode.Sequence[0].Primitive.(game.Discard)
			if !ok {
				t.Fatalf("primitive = %T, want game.Discard", mode.Sequence[0].Primitive)
			}
			if discard.Amount.Value() != tc.wantAmt {
				t.Errorf("amount = %v, want %d", discard.Amount, tc.wantAmt)
			}
			if discard.Player != game.EventPlayerReference() {
				t.Errorf("player = %v, want EventPlayerReference", discard.Player)
			}
		})
	}
}

func TestLowerEventPlayerMillInTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Watcher",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a player draws a card, they mill 2 cards.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	mill, ok := mode.Sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("primitive = %T, want game.Mill", mode.Sequence[0].Primitive)
	}
	if mill.Amount.Value() != 2 {
		t.Errorf("amount = %v, want 2", mill.Amount)
	}
	if mill.Player != game.EventPlayerReference() {
		t.Errorf("player = %v, want EventPlayerReference", mill.Player)
	}
}

func TestLowerEventPlayerDrawFailsClosedOnUnsupportedForm(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		// Non-EventPlayer reference must still fail closed.
		"Whenever a creature attacks, they draw a card.",
		// Unsupported dynamic amount.
		"Whenever an opponent draws a card, they draw X cards.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: oracleText,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected diagnostic for %q but none produced", oracleText)
			}
		})
	}
}

func TestLowerEventPermanentDestroyItInTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{"dies trigger", "Whenever a creature dies, destroy it."},
		{"attack trigger", "Whenever a creature attacks, destroy it."},
		{"tapped trigger", "Whenever a creature becomes tapped, destroy it."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			destroy, ok := mode.Sequence[0].Primitive.(game.Destroy)
			if !ok {
				t.Fatalf("primitive = %T, want game.Destroy", mode.Sequence[0].Primitive)
			}
			if destroy.Object != game.EventPermanentReference() {
				t.Errorf("object = %v, want EventPermanentReference", destroy.Object)
			}
		})
	}
}

func TestLowerEventPermanentExileItInTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Watcher",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature becomes tapped, exile it.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok {
		t.Fatalf("primitive = %T, want game.Exile", mode.Sequence[0].Primitive)
	}
	if exile.Object != game.EventPermanentReference() {
		t.Errorf("object = %v, want EventPermanentReference", exile.Object)
	}
}

func TestLowerEventPermanentTapItInTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Watcher",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature becomes tapped, tap it.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	tap, ok := mode.Sequence[0].Primitive.(game.Tap)
	if !ok {
		t.Fatalf("primitive = %T, want game.Tap", mode.Sequence[0].Primitive)
	}
	if tap.Object != game.EventPermanentReference() {
		t.Errorf("object = %v, want EventPermanentReference", tap.Object)
	}
}

func TestLowerEventPermanentUntapItInTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Watcher",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature enters, untap it.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	untap, ok := mode.Sequence[0].Primitive.(game.Untap)
	if !ok {
		t.Fatalf("primitive = %T, want game.Untap", mode.Sequence[0].Primitive)
	}
	if untap.Object != game.EventPermanentReference() {
		t.Errorf("object = %v, want EventPermanentReference", untap.Object)
	}
}

func TestLowerEventPermanentReturnToHandInTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{"attack trigger bounce", "Whenever a creature attacks, return it to its owner's hand."},
		{"tapped trigger bounce", "Whenever a creature becomes tapped, return it to its owner's hand."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			bounce, ok := mode.Sequence[0].Primitive.(game.Bounce)
			if !ok {
				t.Fatalf("primitive = %T, want game.Bounce", mode.Sequence[0].Primitive)
			}
			if bounce.Object != game.EventPermanentReference() {
				t.Errorf("object = %v, want EventPermanentReference", bounce.Object)
			}
		})
	}
}

// TestLowerDiesEventCardReturnStillUsesMoveCard verifies that the dies-trigger
// "return it to its owner's hand" path still routes through lowerDiesEventCardEffect
// (MoveCard from graveyard to hand, not Bounce) after the shared pronoun path
// is in place. Dies-trigger LKI requires MoveCard, not Bounce.
func TestLowerDiesEventCardReturnStillUsesMoveCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, return it to its owner's hand.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard for dies trigger return", mode.Sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceEvent ||
		move.FromZone != zone.Graveyard ||
		move.Destination != zone.Hand {
		t.Fatalf("move = %+v, want event card from graveyard to hand", move)
	}
}

func TestLowerEventPermanentSacrificeItInTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{"dies trigger sacrifice", "Whenever a creature dies, sacrifice it."},
		{"attack trigger sacrifice", "Whenever a creature attacks, sacrifice it."},
		{"ETB trigger sacrifice", "Whenever a creature enters, sacrifice it."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			sac, ok := mode.Sequence[0].Primitive.(game.Sacrifice)
			if !ok {
				t.Fatalf("primitive = %T, want game.Sacrifice", mode.Sequence[0].Primitive)
			}
			if sac.Object != game.EventPermanentReference() {
				t.Errorf("object = %v, want EventPermanentReference", sac.Object)
			}
		})
	}
}

func TestLowerSourceBoundSacrificeItInSelfTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{"ETB self sacrifice", "When this creature enters, sacrifice it."},
		{"attack self sacrifice", "Whenever this creature attacks, sacrifice it."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Doomed Soldier",
				Layout:     "normal",
				TypeLine:   "Creature — Human Soldier",
				OracleText: tc.oracle,
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			sac, ok := mode.Sequence[0].Primitive.(game.Sacrifice)
			if !ok {
				t.Fatalf("primitive = %T, want game.Sacrifice", mode.Sequence[0].Primitive)
			}
			if sac.Object.Kind() == game.ObjectReferenceNone {
				t.Error("sacrifice object is zero, want source or event reference")
			}
		})
	}
}

func TestLowerEventPermanentPronounEffectFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		// Text mismatch: "that creature" instead of "it".
		"Whenever a creature attacks, destroy that creature.",
		// Negated form is unsupported.
		"Whenever a creature attacks, don't destroy it.",
		// Wrong text form for return.
		"Whenever a creature attacks, return it to the battlefield.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: oracleText,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected diagnostic for unsupported pronoun body %q", oracleText)
			}
		})
	}
}
