package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerCounterSpellTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		oracleText        string
		wantSpellTypes    []types.Card
		wantExcludedTypes []types.Card
		wantKinds         []game.StackObjectKind
	}{
		{
			name:       "any spell",
			oracleText: "Counter target spell.",
			wantKinds:  []game.StackObjectKind{game.StackSpell},
		},
		{
			name:           "creature spell",
			oracleText:     "Counter target creature spell.",
			wantSpellTypes: []types.Card{types.Creature},
			wantKinds:      []game.StackObjectKind{game.StackSpell},
		},
		{
			name:           "artifact spell",
			oracleText:     "Counter target artifact spell.",
			wantSpellTypes: []types.Card{types.Artifact},
			wantKinds:      []game.StackObjectKind{game.StackSpell},
		},
		{
			name:           "instant spell",
			oracleText:     "Counter target instant spell.",
			wantSpellTypes: []types.Card{types.Instant},
			wantKinds:      []game.StackObjectKind{game.StackSpell},
		},
		{
			name:           "sorcery spell",
			oracleText:     "Counter target sorcery spell.",
			wantSpellTypes: []types.Card{types.Sorcery},
			wantKinds:      []game.StackObjectKind{game.StackSpell},
		},
		{
			name:              "noncreature spell",
			oracleText:        "Counter target noncreature spell.",
			wantExcludedTypes: []types.Card{types.Creature},
			wantKinds:         []game.StackObjectKind{game.StackSpell},
		},
		{
			name:       "activated ability",
			oracleText: "Counter target activated ability.",
			wantKinds:  []game.StackObjectKind{game.StackActivatedAbility},
		},
		{
			name:       "triggered ability",
			oracleText: "Counter target triggered ability.",
			wantKinds:  []game.StackObjectKind{game.StackTriggeredAbility},
		},
		{
			name:       "activated or triggered ability",
			oracleText: "Counter target activated or triggered ability.",
			wantKinds:  []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility},
		},
		{
			name:       "spell activated or triggered ability",
			oracleText: "Counter target spell, activated ability, or triggered ability.",
			wantKinds:  []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			ability := face.SpellAbility.Val
			if len(ability.Modes) != 1 {
				t.Fatalf("modes = %d, want 1", len(ability.Modes))
			}
			mode := ability.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			target := mode.Targets[0]
			if target.Allow != game.TargetAllowStackObject {
				t.Fatalf("target allow = %v, want stack object", target.Allow)
			}
			if !slices.Equal(target.Predicate.SpellCardTypes, test.wantSpellTypes) {
				t.Fatalf("spell types = %+v, want %+v", target.Predicate.SpellCardTypes, test.wantSpellTypes)
			}
			if !slices.Equal(target.Predicate.ExcludedSpellCardTypes, test.wantExcludedTypes) {
				t.Fatalf("excluded spell types = %+v, want %+v", target.Predicate.ExcludedSpellCardTypes, test.wantExcludedTypes)
			}
			if !slices.Equal(target.Predicate.StackObjectKinds, test.wantKinds) {
				t.Fatalf("stack object kinds = %+v, want %+v", target.Predicate.StackObjectKinds, test.wantKinds)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
			}
			counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
			if !ok {
				t.Fatalf("primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
			}
			if counter.Object.Kind() != game.ObjectReferenceTargetStackObject || counter.Object.TargetIndex() != 0 {
				t.Fatalf("counter object = %+v, want target stack object 0", counter.Object)
			}
		})
	}
}

func TestLowerCounterSpellUnionTypeTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		oracleText        string
		wantSpellTypesAny []types.Card
		wantExileInstead  bool
	}{
		{
			name:              "instant or sorcery",
			oracleText:        "Counter target instant or sorcery spell.",
			wantSpellTypesAny: []types.Card{types.Instant, types.Sorcery},
		},
		{
			name:              "artifact or enchantment",
			oracleText:        "Counter target artifact or enchantment spell.",
			wantSpellTypesAny: []types.Card{types.Artifact, types.Enchantment},
		},
		{
			name:              "creature or planeswalker exile instead",
			oracleText:        "Counter target creature or planeswalker spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.",
			wantSpellTypesAny: []types.Card{types.Creature, types.Planeswalker},
			wantExileInstead:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter Union",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			target := mode.Targets[0]
			if target.Allow != game.TargetAllowStackObject {
				t.Fatalf("target allow = %v, want stack object", target.Allow)
			}
			if !slices.Equal(target.Predicate.SpellCardTypesAny, test.wantSpellTypesAny) {
				t.Fatalf("spell card types any = %+v, want %+v", target.Predicate.SpellCardTypesAny, test.wantSpellTypesAny)
			}
			if len(target.Predicate.SpellCardTypes) != 0 {
				t.Fatalf("spell card types = %+v, want none", target.Predicate.SpellCardTypes)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
			}
			counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
			if !ok {
				t.Fatalf("primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
			}
			if counter.ExileInstead != test.wantExileInstead {
				t.Fatalf("exile instead = %v, want %v", counter.ExileInstead, test.wantExileInstead)
			}
		})
	}
}

func TestLowerCounterSpellQualifiedTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		oracleText      string
		wantKinds       []game.StackObjectKind
		wantController  game.ControllerRelation
		wantSourceTypes []types.Card
		wantSupertypes  []types.Super
		wantColorless   bool
	}{
		{
			name:            "activated ability from an artifact source",
			oracleText:      "Counter target activated ability from an artifact source.",
			wantKinds:       []game.StackObjectKind{game.StackActivatedAbility},
			wantSourceTypes: []types.Card{types.Artifact},
		},
		{
			name:           "triggered ability you don't control",
			oracleText:     "Counter target triggered ability you don't control.",
			wantKinds:      []game.StackObjectKind{game.StackTriggeredAbility},
			wantController: game.ControllerNotYou,
		},
		{
			name:           "activated triggered or legendary spell",
			oracleText:     "Counter target activated ability, triggered ability, or legendary spell.",
			wantKinds:      []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
			wantSupertypes: []types.Super{types.Legendary},
		},
		{
			name:          "triggered ability or colorless spell",
			oracleText:    "Counter target triggered ability or colorless spell.",
			wantKinds:     []game.StackObjectKind{game.StackTriggeredAbility, game.StackSpell},
			wantColorless: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Qualified Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			predicate := mode.Targets[0].Predicate
			if !slices.Equal(predicate.StackObjectKinds, test.wantKinds) {
				t.Fatalf("stack object kinds = %+v, want %+v", predicate.StackObjectKinds, test.wantKinds)
			}
			if predicate.Controller != test.wantController {
				t.Fatalf("controller = %v, want %v", predicate.Controller, test.wantController)
			}
			if !slices.Equal(predicate.StackObjectSourceTypes, test.wantSourceTypes) {
				t.Fatalf("source types = %+v, want %+v", predicate.StackObjectSourceTypes, test.wantSourceTypes)
			}
			if !slices.Equal(predicate.SpellSupertypes, test.wantSupertypes) {
				t.Fatalf("spell supertypes = %+v, want %+v", predicate.SpellSupertypes, test.wantSupertypes)
			}
			if predicate.SpellColorless != test.wantColorless {
				t.Fatalf("spell colorless = %v, want %v", predicate.SpellColorless, test.wantColorless)
			}
		})
	}
}

func TestLowerCounterSpellManaValueTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantManaV  compare.Int
	}{
		{
			name:       "exact mana value",
			oracleText: "Counter target spell with mana value 1.",
			wantManaV:  compare.Int{Op: compare.Equal, Value: 1},
		},
		{
			name:       "mana value or less",
			oracleText: "Counter target spell with mana value 2 or less.",
			wantManaV:  compare.Int{Op: compare.LessOrEqual, Value: 2},
		},
		{
			name:       "mana value or greater",
			oracleText: "Counter target spell with mana value 3 or greater.",
			wantManaV:  compare.Int{Op: compare.GreaterOrEqual, Value: 3},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Mana Value Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			predicate := mode.Targets[0].Predicate
			if !slices.Equal(predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) {
				t.Fatalf("stack object kinds = %+v, want [StackSpell]", predicate.StackObjectKinds)
			}
			if !predicate.ManaValue.Exists {
				t.Fatal("mana value predicate missing")
			}
			if predicate.ManaValue.Val != test.wantManaV {
				t.Fatalf("mana value = %+v, want %+v", predicate.ManaValue.Val, test.wantManaV)
			}
		})
	}
}

func TestLowerCounterSpellWithDrawRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dismiss",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Counter target spell. Draw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}

	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if mode.Targets[0].Allow != game.TargetAllowStackObject {
		t.Fatalf("target allow = %v, want stack object", mode.Targets[0].Allow)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want counter plus draw", len(mode.Sequence))
	}
	if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("first primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.Draw); !ok {
		t.Fatalf("second primitive = %T, want game.Draw", mode.Sequence[1].Primitive)
	}
}

func TestLowerArcaneDenialEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Arcane Denial",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{U}",
		OracleText: "Counter target spell. Its controller may draw up to two cards at the beginning of the next turn's upkeep.\nYou draw a card at the beginning of the next turn's upkeep.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %#v, want one target and three instructions", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("first primitive = %T, want CounterObject", mode.Sequence[0].Primitive)
	}
	targetDelayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok ||
		targetDelayed.Trigger.Timing != game.DelayedAtBeginningOfNextUpkeep ||
		targetDelayed.Trigger.Optional {
		t.Fatalf("target delayed trigger = %#v", mode.Sequence[1].Primitive)
	}
	targetSequence := targetDelayed.Trigger.Content.Modes[0].Sequence
	if len(targetSequence) != 2 {
		t.Fatalf("target delayed sequence = %#v, want choose then draw", targetSequence)
	}
	choose, ok := targetSequence[0].Primitive.(game.Choose)
	if !ok ||
		choose.Choice.Kind != game.ResolutionChoiceNumber ||
		choose.Choice.MinNumber != 0 ||
		choose.Choice.MaxNumber != 2 ||
		choose.Choice.PlayerReference == nil {
		t.Fatalf("choice = %#v", targetSequence[0].Primitive)
	}
	if choose.Choice.PlayerReference.Kind() != game.PlayerReferenceCapturedTargetController ||
		choose.Choice.PlayerReference.TargetIndex() != 0 {
		t.Fatalf("choice player = %#v", choose.Choice.PlayerReference)
	}
	draw, ok := targetSequence[1].Primitive.(game.Draw)
	if !ok || !draw.Amount.IsDynamic() {
		t.Fatalf("target draw = %#v", targetSequence[1].Primitive)
	}
	dynamic := draw.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountChosenNumber ||
		game.ChoiceKey(dynamic.ResultKey) != choose.PublishChoice {
		t.Fatalf("target draw amount = %#v", dynamic)
	}
	if draw.Player.Kind() != game.PlayerReferenceCapturedTargetController ||
		draw.Player.TargetIndex() != 0 {
		t.Fatalf("target draw player = %#v", draw.Player)
	}
	controllerDelayed, ok := mode.Sequence[2].Primitive.(game.CreateDelayedTrigger)
	if !ok ||
		controllerDelayed.Trigger.Timing != game.DelayedAtBeginningOfNextUpkeep {
		t.Fatalf("controller delayed trigger = %#v", mode.Sequence[2].Primitive)
	}
	controllerDraw, ok := controllerDelayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || controllerDraw.Amount.IsDynamic() || controllerDraw.Amount.Value() != 1 {
		t.Fatalf("controller draw = %#v", controllerDraw)
	}
}

func TestLowerCounterWithExactNextTurnUpkeepDrawSibling(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sibling Denial",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Counter target spell. You draw two cards at the beginning of the next turn's upkeep.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want counter plus delayed draw", mode.Sequence)
	}
	delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextUpkeep {
		t.Fatalf("delayed instruction = %#v", mode.Sequence[1].Primitive)
	}
	draw, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.IsDynamic() || draw.Amount.Value() != 2 {
		t.Fatalf("delayed draw = %#v, want fixed two", draw)
	}
}

func TestLowerCounterDelayedDrawsRejectUnsupportedShapes(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Counter target spell. Its controller may draw up to two cards at the beginning of the next end step. You draw a card at the beginning of the next turn's upkeep.",
		"Counter target spell. Its controller may draw up to two cards at the beginning of your next upkeep. You draw a card at the beginning of the next turn's upkeep.",
		"Counter target spell. If that spell was countered this way, its controller may draw up to two cards at the beginning of the next turn's upkeep.",
		"Counter target spell. Its controller may draw up to two cards at the beginning of the next turn's upkeep. You gain 1 life at the beginning of the next turn's upkeep.",
		"Counter target spell. Target player draws a card at the beginning of the next turn's upkeep.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Denial",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: text,
			})
			if len(diagnostics) == 0 {
				t.Fatal("unsupported delayed sequence lowered")
			}
		})
	}
}

func TestLowerCounterThenTargetControllerCreatesToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Swan Song",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{U}",
		Colors:     []string{"U"},
		OracleText: "Counter target enchantment, instant, or sorcery spell. Its controller creates a 2/2 blue Bird creature token with flying.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 ||
		!slices.Equal(mode.Targets[0].Predicate.SpellCardTypesAny, []types.Card{
			types.Enchantment, types.Instant, types.Sorcery,
		}) {
		t.Fatalf("targets = %#v", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want counter then create token", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("first primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("second primitive = %T, want game.CreateToken", mode.Sequence[1].Primitive)
	}
	recipientObject, ok := create.Recipient.Val.Object()
	if !create.Recipient.Exists || !ok ||
		recipientObject.Kind() != game.ObjectReferenceTargetStackObject ||
		recipientObject.TargetIndex() != 0 {
		t.Fatalf("recipient = %#v, want controller of target stack object 0", create.Recipient)
	}
	token, ok := create.Source.TokenDefRef()
	if !ok ||
		token.Name != "Bird" ||
		!slices.Equal(token.Types, []types.Card{types.Creature}) ||
		!slices.Equal(token.Subtypes, []types.Sub{types.Bird}) ||
		!slices.Equal(token.Colors, []color.Color{color.Blue}) ||
		!token.Power.Exists || token.Power.Val.Value != 2 ||
		!token.Toughness.Exists || token.Toughness.Val.Value != 2 ||
		len(token.StaticAbilities) != 1 ||
		!game.BodyHasKeyword(&token.StaticAbilities[0], game.Flying) {
		t.Fatalf("token = %#v, want 2/2 blue Bird with flying", token)
	}
}

func TestLowerCounterThenTargetControllerCreatesTreasureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "An Offer You Can't Refuse",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{U}",
		Colors:     []string{"U"},
		OracleText: "Counter target noncreature spell. Its controller creates two Treasure tokens. (They're artifacts with \"{T}, Sacrifice this token: Add one mana of any color.\")",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 ||
		!slices.Equal(mode.Targets[0].Predicate.ExcludedSpellCardTypes, []types.Card{types.Creature}) {
		t.Fatalf("targets = %#v, want noncreature spell", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want counter then create token", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("first primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("second primitive = %T, want game.CreateToken", mode.Sequence[1].Primitive)
	}
	if create.Amount.Value() != 2 {
		t.Fatalf("amount = %d, want 2", create.Amount.Value())
	}
	recipientObject, ok := create.Recipient.Val.Object()
	if !create.Recipient.Exists || !ok ||
		recipientObject.Kind() != game.ObjectReferenceTargetStackObject ||
		recipientObject.TargetIndex() != 0 {
		t.Fatalf("recipient = %#v, want controller of target stack object 0", create.Recipient)
	}
	token, ok := create.Source.TokenDefRef()
	if !ok ||
		token.Name != string(types.Treasure) ||
		!slices.Equal(token.Subtypes, []types.Sub{types.Treasure}) {
		t.Fatalf("token = %#v, want Treasure", token)
	}
}

func TestLowerCounterThenTargetControllerTokenRejectionBoundaries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*contentCtx)
	}{
		{
			name: "sequence cardinality",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects = append(ctx.content.Effects, compiler.CompiledEffect{})
			},
		},
		{
			name: "counter sequence connection",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects[0].Connection = parser.EffectConnectionAnd
			},
		},
		{
			name: "token sequence connection",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects[1].Connection = parser.EffectConnectionAnd
			},
		},
		{
			name: "exact linked counter token envelope",
			mutate: func(ctx *contentCtx) {
				ctx.content.Conditions = []compiler.CompiledCondition{{}}
			},
		},
		{
			name: "exact mandatory counter",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects[0].Optional = true
			},
		},
		{
			name: "spell target shape",
			mutate: func(ctx *contentCtx) {
				ctx.content.Targets[0].Selector.Another = true
				ctx.content.Effects[0].Targets[0].Selector.Another = true
			},
		},
		{
			name: "target controller token recipient",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects[1].Context = parser.EffectContextController
			},
		},
		{
			name: "reference binding to target zero",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects[1].References[0].Occurrence = 1
			},
		},
		{
			name: "subject reference binding to target zero",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects[1].SubjectReferences[0].Occurrence = 1
			},
		},
		{
			name: "unmodified token shape",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects[1].Selector.Tapped = true
			},
		},
		{
			name: "supported token definition",
			mutate: func(ctx *contentCtx) {
				ctx.content.Effects[1].TokenPTKnown = false
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := compiledSwanSongContentCtx(t)
			test.mutate(&ctx)
			if _, ok := lowerCounterThenTargetControllerTokenSequence(ctx); ok {
				t.Fatal("lowered unsupported boundary mutation")
			}
		})
	}
}

func compiledSwanSongContentCtx(t *testing.T) contentCtx {
	t.Helper()
	compilation, diagnostics := compileTestOracle(
		"Counter target enchantment, instant, or sorcery spell. Its controller creates a 2/2 blue Bird creature token with flying.",
		parser.Context{InstantOrSorcery: true},
		compiler.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	return contentCtx{span: content.Span, content: content}
}

func TestLowerCounterThenTargetControllerCreatesTokenFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Counter target blue enchantment, instant, or sorcery spell. Its controller creates a 2/2 blue Bird creature token with flying.",
		"Counter two target enchantment, instant, or sorcery spells. Their controller creates a 2/2 blue Bird creature token with flying.",
		"Counter target enchantment, instant, or sorcery spell. You create a 2/2 blue Bird creature token with flying.",
		"Counter target enchantment, instant, or sorcery spell. Its owner creates a 2/2 blue Bird creature token with flying.",
		"Counter target enchantment, instant, or sorcery spell. Its controller creates a tapped 2/2 blue Bird creature token with flying.",
		"Counter target enchantment, instant, or sorcery spell. Its controller creates a 2/2 blue Bird creature token with flying and haste.",
	} {
		_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Near Swan Song",
			Layout:     "normal",
			TypeLine:   "Instant",
			ManaCost:   "{U}",
			Colors:     []string{"U"},
			OracleText: oracleText,
		}, "n")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Errorf("%q lowered cleanly, want fail closed", oracleText)
		}
	}
}

func TestLowerCounterSpellUnlessPays(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Leak",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Counter target spell unless its controller pays {3}.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowStackObject {
		t.Fatalf("targets = %+v, want one stack-object target", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want pay then counter", len(mode.Sequence))
	}
	pay, ok := mode.Sequence[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("first primitive = %T, want game.Pay", mode.Sequence[0].Primitive)
	}
	if mode.Sequence[0].PublishResult != "unless-paid" {
		t.Fatalf("pay result key = %q, want unless-paid", mode.Sequence[0].PublishResult)
	}
	if !pay.Payment.ManaCost.Exists || !slices.Equal(pay.Payment.ManaCost.Val, cost.Mana{cost.O(3)}) {
		t.Fatalf("payment mana = %+v, want {3}", pay.Payment.ManaCost)
	}
	payer, ok := pay.Payment.Payer.Val.Object()
	if !pay.Payment.Payer.Exists || !ok || payer.Kind() != game.ObjectReferenceTargetStackObject || payer.TargetIndex() != 0 {
		t.Fatalf("payer = %+v, want controller of target stack object 0", pay.Payment.Payer)
	}
	counterObject, ok := mode.Sequence[1].Primitive.(game.CounterObject)
	if !ok {
		t.Fatalf("second primitive = %T, want game.CounterObject", mode.Sequence[1].Primitive)
	}
	if counterObject.Object.Kind() != game.ObjectReferenceTargetStackObject || counterObject.Object.TargetIndex() != 0 {
		t.Fatalf("counter object = %+v, want target stack object 0", counterObject.Object)
	}
	gate := mode.Sequence[1].ResultGate
	if !gate.Exists || gate.Val.Key != "unless-paid" || gate.Val.Succeeded != game.TriFalse {
		t.Fatalf("counter result gate = %+v, want unless-paid succeeded false", gate)
	}
}

func TestLowerCounterSpellUnlessPaysDynamicCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Circular Logic",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		OracleText: "Counter target spell unless its controller pays {1} for each card in your graveyard.\nMadness {U} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want pay then counter", len(mode.Sequence))
	}
	pay, ok := mode.Sequence[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("first primitive = %T, want game.Pay", mode.Sequence[0].Primitive)
	}
	if !pay.Payment.ManaCost.Exists || !slices.Equal(pay.Payment.ManaCost.Val, cost.Mana{cost.O(1)}) {
		t.Fatalf("payment mana = %+v, want {1}", pay.Payment.ManaCost)
	}
	if !pay.Payment.ManaCostMultiplier.Exists || pay.Payment.ManaCostMultiplier.Val == nil {
		t.Fatalf("payment multiplier = %+v, want dynamic count", pay.Payment.ManaCostMultiplier)
	}
	multiplier := pay.Payment.ManaCostMultiplier.Val
	if multiplier.Kind != game.DynamicAmountCountCardsInZone || multiplier.CardZone != zone.Graveyard {
		t.Fatalf("multiplier = %+v, want count of cards in graveyard", multiplier)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.CounterObject); !ok {
		t.Fatalf("second primitive = %T, want game.CounterObject", mode.Sequence[1].Primitive)
	}
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want one (Madness)", len(face.StaticAbilities))
	}
}

func TestLowerCounterSpellColorTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		oracleText         string
		wantColors         []color.Color
		wantExcludedColors []color.Color
		wantColorless      bool
		wantMulticolored   bool
	}{
		{
			name:       "blue spell",
			oracleText: "Counter target blue spell.",
			wantColors: []color.Color{color.Blue},
		},
		{
			name:               "nonblue spell",
			oracleText:         "Counter target nonblue spell.",
			wantExcludedColors: []color.Color{color.Blue},
		},
		{
			name:          "colorless spell",
			oracleText:    "Counter target colorless spell.",
			wantColorless: true,
		},
		{
			name:             "multicolored spell",
			oracleText:       "Counter target multicolored spell.",
			wantMulticolored: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Color Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			predicate := mode.Targets[0].Predicate
			if !slices.Equal(predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) {
				t.Fatalf("stack object kinds = %+v, want [StackSpell]", predicate.StackObjectKinds)
			}
			if !slices.Equal(predicate.SpellColors, test.wantColors) {
				t.Fatalf("spell colors = %+v, want %+v", predicate.SpellColors, test.wantColors)
			}
			if !slices.Equal(predicate.SpellExcludedColors, test.wantExcludedColors) {
				t.Fatalf("spell excluded colors = %+v, want %+v", predicate.SpellExcludedColors, test.wantExcludedColors)
			}
			if predicate.SpellColorless != test.wantColorless {
				t.Fatalf("spell colorless = %v, want %v", predicate.SpellColorless, test.wantColorless)
			}
			if predicate.SpellMulticolored != test.wantMulticolored {
				t.Fatalf("spell multicolored = %v, want %v", predicate.SpellMulticolored, test.wantMulticolored)
			}
			if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
				t.Fatalf("primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
			}
		})
	}
}

func TestLowerCounterSpellRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Counter target monocolored spell.",
		"Counter target blue creature spell.",
		"Counter target spell unless its controller pays {X}.",
		"Counter target activated ability unless its controller pays {1}.",
		"Counter target activated ability. Draw a card.",
		"Counter target spell or ability that targets a creature.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: oracleText,
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected counter-spell diagnostic")
			}
		})
	}
}

func TestLowerNinjutsuAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ninja",
		Layout:     "normal",
		TypeLine:   "Creature — Human Ninja",
		OracleText: "Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control to hand: Put this card onto the battlefield from your hand tapped and attacking.)",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if !ability.ManaCost.Exists || !slices.Equal(ability.ManaCost.Val, cost.Mana{cost.O(1), cost.U}) {
		t.Fatalf("mana cost = %#v, want {1}{U}", ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalReturnUnblockedAttacker {
		t.Fatalf("additional costs = %#v, want return unblocked attacker", ability.AdditionalCosts)
	}
	if len(ability.KeywordAbilities) != 1 {
		t.Fatalf("got %d keyword abilities, want 1", len(ability.KeywordAbilities))
	}
	if _, ok := ability.KeywordAbilities[0].(game.NinjutsuKeyword); !ok {
		t.Fatalf("keyword ability = %T, want game.NinjutsuKeyword", ability.KeywordAbilities[0])
	}
}

func TestLowerNinjutsuAbilityRejectsMalformedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Ninjutsu",
		"Ninjutsu {1}{U} extra text",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Malformed Ninja",
				Layout:     "normal",
				TypeLine:   "Creature — Ninja",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected malformed Ninjutsu diagnostic")
			}
		})
	}
}

func TestLowerSelfCardGraveyardReturnToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dragon",
		Layout:     "normal",
		TypeLine:   "Creature — Dragon",
		OracleText: "{3}{W}{W}: Return this card from your graveyard to your hand.",
		Power:      new("5"),
		Toughness:  new("5"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if ability.ZoneOfFunction != zone.Graveyard {
		t.Fatalf("zone of function = %v, want graveyard", ability.ZoneOfFunction)
	}
	sequence := ability.Content.Modes[0].Sequence
	move, ok := sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceSource || move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
		t.Fatalf("move = %#v", move)
	}
}

func TestLowerSelfCardGraveyardReturnToBattlefieldTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Zombie",
		Layout:     "normal",
		TypeLine:   "Creature — Zombie",
		OracleText: "{1}{B}, Discard two cards: Return this card from your graveyard to the battlefield tapped.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if face.ActivatedAbilities[0].ZoneOfFunction != zone.Graveyard {
		t.Fatalf("zone of function = %v, want graveyard", face.ActivatedAbilities[0].ZoneOfFunction)
	}
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	put, ok := sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("first primitive = %T, want game.PutOnBattlefield", sequence[0].Primitive)
	}
	cardRef, ok := put.Source.CardRef()
	if !ok || cardRef.Kind != game.CardReferenceSource {
		t.Fatalf("source = %#v", put.Source)
	}
	if !put.EntryTapped {
		t.Fatalf("put = %#v, want EntryTapped", put)
	}
}

func TestLowerSelfCardGraveyardReturnToBattlefieldWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Construct",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "{3}{B}: Return this card from your graveyard to the battlefield tapped with two +1/+1 counters on it.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if face.ActivatedAbilities[0].ZoneOfFunction != zone.Graveyard {
		t.Fatalf("zone of function = %v, want graveyard", face.ActivatedAbilities[0].ZoneOfFunction)
	}
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	put, ok := sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("primitive = %T, want game.PutOnBattlefield", sequence[0].Primitive)
	}
	if !put.EntryTapped ||
		len(put.EntryCounters) != 1 ||
		put.EntryCounters[0].Kind != counter.PlusOnePlusOne ||
		put.EntryCounters[0].Amount != 2 {
		t.Fatalf("put = %#v", put)
	}
}

func TestLowerSimpleDelayedOneShotEffects(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cardName   string
		typeLine   string
		oracleText string
		timing     game.DelayedTriggerTiming
		check      func(t *testing.T, primitive game.Primitive)
	}{
		{
			name:       "draw at next upkeep",
			cardName:   "Test Insight",
			typeLine:   "Instant",
			oracleText: "Draw a card at the beginning of the next turn's upkeep.",
			timing:     game.DelayedAtBeginningOfNextUpkeep,
			check: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				draw, ok := primitive.(game.Draw)
				if !ok || draw.Amount.IsDynamic() || draw.Amount.Value() != 1 {
					t.Fatalf("primitive = %#v, want draw one", primitive)
				}
			},
		},
		{
			name:       "self exile at next end step",
			cardName:   "Test Runner",
			typeLine:   "Creature — Human",
			oracleText: "When this creature enters, exile it at the beginning of the next end step.",
			timing:     game.DelayedAtBeginningOfNextEndStep,
			check: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				exile, ok := primitive.(game.Exile)
				if !ok || exile.Object.Kind() != game.ObjectReferenceSourceCard {
					t.Fatalf("primitive = %#v, want source-card exile", primitive)
				}
			},
		},
		{
			name:       "self sacrifice at next end step",
			cardName:   "Test Runner",
			typeLine:   "Creature — Human",
			oracleText: "When this creature enters, sacrifice it at the beginning of the next end step.",
			timing:     game.DelayedAtBeginningOfNextEndStep,
			check: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				sacrifice, ok := primitive.(game.Sacrifice)
				if !ok || sacrifice.Object.Kind() != game.ObjectReferenceSourceCard {
					t.Fatalf("primitive = %#v, want source-card sacrifice", primitive)
				}
			},
		},
		{
			name:       "self return from graveyard at next end step",
			cardName:   "Test God",
			typeLine:   "Legendary Creature — God",
			oracleText: "When Test God dies, return it to its owner's hand at the beginning of the next end step.",
			timing:     game.DelayedAtBeginningOfNextEndStep,
			check: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				move, ok := primitive.(game.MoveCard)
				if !ok ||
					move.Card.Kind != game.CardReferenceSource ||
					move.FromZone != zone.Graveyard ||
					move.Destination != zone.Hand {
					t.Fatalf("primitive = %#v, want source card graveyard-to-hand move", primitive)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       tt.cardName,
				Layout:     "normal",
				TypeLine:   tt.typeLine,
				OracleText: tt.oracleText,
			}
			if strings.Contains(tt.typeLine, "Creature") {
				card.Power = new("2")
				card.Toughness = new("2")
			}
			face := lowerSingleFace(t, card)
			var content game.AbilityContent
			switch {
			case face.SpellAbility.Exists:
				content = face.SpellAbility.Val
			case len(face.TriggeredAbilities) == 1:
				content = face.TriggeredAbilities[0].Content
			default:
				t.Fatalf("lowered face has no single spell or triggered ability: %#v", face)
			}
			if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
				t.Fatalf("content = %#v, want one delayed-trigger instruction", content)
			}
			delayed, ok := content.Modes[0].Sequence[0].Primitive.(game.CreateDelayedTrigger)
			if !ok || delayed.Trigger.Timing != tt.timing {
				t.Fatalf("primitive = %#v, want delayed timing %v", content.Modes[0].Sequence[0].Primitive, tt.timing)
			}
			if len(delayed.Trigger.Content.Modes) != 1 || len(delayed.Trigger.Content.Modes[0].Sequence) != 1 {
				t.Fatalf("delayed content = %#v, want one instruction", delayed.Trigger.Content)
			}
			tt.check(t, delayed.Trigger.Content.Modes[0].Sequence[0].Primitive)
		})
	}
}

func TestLowerDelayedOneShotEffectRejectsTargetReference(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Delay",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature at the beginning of the next end step.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected delayed target effect to remain unsupported")
	}
}

func TestLowerOrderedSequenceWithDelayedOneShotEffect(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sequence",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Draw a card, then you gain 2 life at the beginning of the next end step.",
	})
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 2 {
		t.Fatalf("content = %#v, want two ordered instructions", content)
	}
	if _, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("first primitive = %#v, want draw", content.Modes[0].Sequence[0].Primitive)
	}
	delayed, ok := content.Modes[0].Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("second primitive = %#v, want delayed end-step trigger", content.Modes[0].Sequence[1].Primitive)
	}
	if _, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.GainLife); !ok {
		t.Fatalf("delayed primitive = %#v, want gain life", delayed.Trigger.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerDelayedBlink(t *testing.T) {
	t.Parallel()
	for _, reference := range []string{"that card", "it"} {
		t.Run(reference, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Mist",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: "Exile target creature. Return " + reference + " to the battlefield under its owner's control at the beginning of the next end step.",
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
				t.Fatalf("mode = %#v, want one target and two instructions", mode)
			}
			exile, ok := mode.Sequence[0].Primitive.(game.Exile)
			if !ok || exile.Object != game.TargetPermanentReference(0) || exile.ExileLinkedKey == "" {
				t.Fatalf("exile = %#v, want linked target exile", mode.Sequence[0].Primitive)
			}
			delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
			if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
				t.Fatalf("delayed = %#v, want next-end-step trigger", mode.Sequence[1].Primitive)
			}
			put, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
			key, linked := put.Source.LinkedKey()
			if !ok || !linked || key != exile.ExileLinkedKey {
				t.Fatalf("delayed put = %#v, want linked source %q", put, exile.ExileLinkedKey)
			}
		})
	}
}

func TestLowerMultipleDelayedBlinkPairsUseDistinctKeys(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Double Mist",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Exile target artifact. Return that card to the battlefield under its owner's control at the beginning of the next end step. " +
			"Exile target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 || len(mode.Sequence) != 4 {
		t.Fatalf("mode = %#v, want two targets and four instructions", mode)
	}
	var keys []game.LinkedKey
	for i, targetIndex := range []int{0, 1} {
		exile, ok := mode.Sequence[i*2].Primitive.(game.Exile)
		if !ok || exile.Object != game.TargetPermanentReference(targetIndex) || exile.ExileLinkedKey == "" {
			t.Fatalf("exile %d = %#v, want linked target %d", i, mode.Sequence[i*2].Primitive, targetIndex)
		}
		keys = append(keys, exile.ExileLinkedKey)
		delayed, ok := mode.Sequence[i*2+1].Primitive.(game.CreateDelayedTrigger)
		if !ok {
			t.Fatalf("instruction %d = %#v, want delayed trigger", i*2+1, mode.Sequence[i*2+1].Primitive)
		}
		put, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
		if !ok {
			t.Fatalf("delayed instruction %d = %#v, want put on battlefield", i, delayed.Trigger.Content.Modes[0].Sequence[0].Primitive)
		}
		key, ok := put.Source.LinkedKey()
		if !ok || key != exile.ExileLinkedKey {
			t.Fatalf("put %d linked key = %q (%v), want %q", i, key, ok, exile.ExileLinkedKey)
		}
	}
	if keys[0] == keys[1] {
		t.Fatalf("blink keys = %q/%q, want distinct", keys[0], keys[1])
	}
}

func TestLowerDelayedBlinkRejectsUnsupportedVariants(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Exile target creature. Return it to the battlefield under your control at the beginning of the next end step.",
		"Exile target creature. Return it to the battlefield under its owner's control with a +1/+1 counter on it at the beginning of the next end step.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Unsupported Mist",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: text,
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported blink variant to fail closed")
			}
		})
	}
}

func TestLowerDelayedTargetReturnUsesLinkedReference(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mask",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{3}, {T}: Target creature you control gets +2/+2 until end of turn. Return it to its owner's hand at the beginning of the next end step.",
	})
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %#v, want one target and two instructions", mode)
	}
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok || modify.PublishLinked == "" {
		t.Fatalf("modify = %#v, want published linked target", mode.Sequence[0].Primitive)
	}
	delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("second primitive = %#v, want delayed trigger", mode.Sequence[1].Primitive)
	}
	bounce, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.Bounce)
	if !ok ||
		bounce.Object.Kind() != game.ObjectReferenceLinkedObject ||
		bounce.Object.LinkID() != string(modify.PublishLinked) {
		t.Fatalf("delayed bounce = %#v, want linked object bounce", bounce)
	}
}

func TestLowerConditionAndDelayedReferenceNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, card := range []*ScryfallCard{
		{
			Name:       "Test Pupils",
			Layout:     "normal",
			TypeLine:   "Creature — Human",
			OracleText: "If a creature dealt damage by this creature this turn would die, exile it instead.",
			Power:      new("3"),
			Toughness:  new("3"),
		},
		{
			Name:       "Test Cathar",
			Layout:     "normal",
			TypeLine:   "Creature — Human",
			OracleText: "When this creature dies, return it to the battlefield transformed under your control at the beginning of the next end step.",
			Power:      new("2"),
			Toughness:  new("2"),
		},
		{
			Name:       "Test Orb",
			Layout:     "normal",
			TypeLine:   "Enchantment — Aura",
			OracleText: "Enchant creature\nWhen enchanted creature dies, return that card to the battlefield under its owner's control at the beginning of the next end step.",
		},
		{
			Name:       "Test Ambiguity",
			Layout:     "normal",
			TypeLine:   "Creature — Human",
			OracleText: "It explores.",
			Power:      new("2"),
			Toughness:  new("2"),
		},
		{
			Name:       "Test Invalid Condition Context",
			Layout:     "normal",
			TypeLine:   "Creature — Human",
			OracleText: "This creature has flying unless you control an artifact.",
			Power:      new("2"),
			Toughness:  new("2"),
		},
	} {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(card)
			if len(diagnostics) == 0 {
				t.Fatal("near-miss unexpectedly lowered")
			}
			for _, diagnostic := range diagnostics {
				if diagnostic.Span.End.Offset <= diagnostic.Span.Start.Offset {
					t.Fatalf("diagnostic has no source span: %#v", diagnostic)
				}
			}
		})
	}
}

func TestLowerCounterAbilityInEnterTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		optional   bool
		wantKinds  []game.StackObjectKind
	}{
		{
			name:       "activated ability",
			oracleText: "When this creature enters, counter target activated ability.",
			wantKinds:  []game.StackObjectKind{game.StackActivatedAbility},
		},
		{
			name:       "triggered ability",
			oracleText: "When this creature enters, counter target triggered ability.",
			wantKinds:  []game.StackObjectKind{game.StackTriggeredAbility},
		},
		{
			name:       "activated or triggered ability",
			oracleText: "When this creature enters, counter target activated or triggered ability.",
			wantKinds:  []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility},
		},
		{
			name:       "optional counter activated ability",
			oracleText: "When this creature enters, you may counter target activated ability.",
			optional:   true,
			wantKinds:  []game.StackObjectKind{game.StackActivatedAbility},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter Enter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Wizard",
				OracleText: test.oracleText,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0]
			if trigger.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
				t.Fatalf("trigger event = %v, want entered battlefield", trigger.Trigger.Pattern.Event)
			}
			if trigger.Optional != test.optional {
				t.Fatalf("optional = %v, want %v", trigger.Optional, test.optional)
			}
			if len(trigger.Content.Modes) != 1 {
				t.Fatalf("modes = %d, want 1", len(trigger.Content.Modes))
			}
			mode := trigger.Content.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			target := mode.Targets[0]
			if target.Allow != game.TargetAllowStackObject {
				t.Fatalf("target allow = %v, want stack object", target.Allow)
			}
			if !slices.Equal(target.Predicate.StackObjectKinds, test.wantKinds) {
				t.Fatalf("stack object kinds = %+v, want %+v", target.Predicate.StackObjectKinds, test.wantKinds)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
			}
			counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
			if !ok {
				t.Fatalf("primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
			}
			if counter.Object.Kind() != game.ObjectReferenceTargetStackObject || counter.Object.TargetIndex() != 0 {
				t.Fatalf("counter object = %+v, want target stack object 0", counter.Object)
			}
		})
	}
}

func TestLowerCounterThenExileInstead(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		oracleText        string
		wantExcludedTypes []types.Card
		wantSpellTypes    []types.Card
	}{
		{
			name: "noncreature spell",
			oracleText: "Counter target noncreature spell. If that spell is countered this way, " +
				"exile it instead of putting it into its owner's graveyard.",
			wantExcludedTypes: []types.Card{types.Creature},
		},
		{
			name: "any spell",
			oracleText: "Counter target spell. If that spell is countered this way, " +
				"exile it instead of putting it into its owner's graveyard.",
		},
		{
			name: "creature spell",
			oracleText: "Counter target creature spell. If that spell is countered this way, " +
				"exile it instead of putting it into its owner's graveyard.",
			wantSpellTypes: []types.Card{types.Creature},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter Exile",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			target := mode.Targets[0]
			if !slices.Equal(target.Predicate.SpellCardTypes, test.wantSpellTypes) {
				t.Fatalf("spell types = %+v, want %+v", target.Predicate.SpellCardTypes, test.wantSpellTypes)
			}
			if !slices.Equal(target.Predicate.ExcludedSpellCardTypes, test.wantExcludedTypes) {
				t.Fatalf("excluded spell types = %+v, want %+v", target.Predicate.ExcludedSpellCardTypes, test.wantExcludedTypes)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
			}
			counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
			if !ok {
				t.Fatalf("primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
			}
			if !counter.ExileInstead {
				t.Fatal("ExileInstead = false, want true")
			}
			if counter.Object.Kind() != game.ObjectReferenceTargetStackObject || counter.Object.TargetIndex() != 0 {
				t.Fatalf("counter object = %+v, want target stack object 0", counter.Object)
			}
		})
	}
}

func TestLowerCounterPlainStillGraveyard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Plain Counter",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Counter target spell.",
	})
	countered, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CounterObject)
	if !ok {
		t.Fatalf("primitive = %T, want game.CounterObject", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if countered.ExileInstead {
		t.Fatal("ExileInstead = true, want false for plain counter")
	}
}
