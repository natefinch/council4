package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileTriggeredAbility(t *testing.T) {
	t.Parallel()
	source := "Whenever a creature enters, if it was cast, draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Trigger == nil ||
		ability.Trigger.Kind != TriggerWhenever ||
		ability.Trigger.Event != "a creature enters" {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if ability.Trigger.Condition == nil || !ability.Trigger.Condition.Intervening {
		t.Fatalf("intervening condition = %#v", ability.Trigger.Condition)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
}

// TestCompileAttackBatchEventPlayerDrawGate verifies Firemane Commando's second
// trigger "Whenever another player attacks with two or more creatures, they draw
// a card if none of those creatures attacked you." compiles to an
// opponent-controller attacker-declared pattern scoped by attacker count, an
// event-player draw effect, and a trailing effect-gate condition carrying the
// no-attacker-attacked-controller predicate (not an intervening condition).
func TestCompileAttackBatchEventPlayerDrawGate(t *testing.T) {
	t.Parallel()
	source := "Whenever another player attacks with two or more creatures, they draw a card if none of those creatures attacked you."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil ||
		ability.Trigger.Pattern.Event != TriggerEventAttackerDeclared ||
		ability.Trigger.Pattern.Controller != ControllerOpponent ||
		ability.Trigger.Pattern.AttackerCountAtLeast != 2 {
		t.Fatalf("trigger pattern = %#v", ability.Trigger)
	}
	if ability.Trigger.Condition != nil {
		t.Fatalf("gate must not be an intervening condition: %#v", ability.Trigger.Condition)
	}
	if len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectDraw ||
		ability.Content.Effects[0].Context != parser.EffectContextEventPlayer {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != ConditionPredicateNoAttackerAttackedController {
		t.Fatalf("conditions = %#v", ability.Content.Conditions)
	}
}

func TestCompileTriggeredAbilityWithInternalEventComma(t *testing.T) {
	t.Parallel()
	source := "Whenever you cast a noncreature, nonland spell, draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Trigger == nil ||
		ability.Trigger.Kind != TriggerWhenever ||
		ability.Trigger.Event != "you cast a noncreature, nonland spell" {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
}

// TestCompileCastTriggerFewerThanSelfCounterInterveningIf verifies Runaway
// Steam-Kin's intervening-if "if this creature has fewer than three +1/+1
// counters on it" compiles onto the source-bound object-match predicate with the
// strict upper bound threaded through as CounterCountLessThan, leaving the
// inclusive minimum zero so the two thresholds stay mutually exclusive.
func TestCompileCastTriggerFewerThanSelfCounterInterveningIf(t *testing.T) {
	t.Parallel()
	source := "Whenever you cast a red spell, if this creature has fewer than three +1/+1 counters on it, put a +1/+1 counter on this creature."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.Condition == nil {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	condition := ability.Trigger.Condition
	if condition.Predicate != ConditionPredicateObjectMatches ||
		condition.ObjectBinding != ReferenceBindingSource {
		t.Fatalf("condition = %#v, want source object-match", condition)
	}
	selection := condition.Selection
	if !selection.CounterKindKnown ||
		selection.CounterKind != counter.PlusOnePlusOne ||
		selection.CounterCountLessThan != 3 ||
		selection.CounterCountAtLeast != 0 {
		t.Fatalf("selection = %#v, want +1/+1 count < 3", selection)
	}
}

// TestCompileEnterTriggerNonTokenInterveningIf verifies Life of the Party's
// negated intervening-if "if it's not a token" compiles onto the event-permanent
// object-match predicate with the NonToken selection flag set, so the runtime
// gate rejects token entries.
func TestCompileEnterTriggerNonTokenInterveningIf(t *testing.T) {
	t.Parallel()
	source := "When this creature enters, if it's not a token, draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.Condition == nil {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	condition := ability.Trigger.Condition
	if condition.Predicate != ConditionPredicateObjectMatches ||
		condition.ObjectBinding != ReferenceBindingEventPermanent {
		t.Fatalf("condition = %#v, want event-permanent object-match", condition)
	}
	if !condition.Selection.NonToken || condition.Selection.TokenOnly {
		t.Fatalf("selection = %#v, want NonToken", condition.Selection)
	}
}

// TestCompileAttachedDiesThatSubjectInterveningIf proves the attached-permanent
// dies trigger "When enchanted creature dies, if that creature was a Horror, ..."
// compiles its "that creature was a <subtype>" back-reference to an
// event-permanent object-match intervening condition, the reusable gate Endless
// Evil needs. The subtype is evaluated against the dead creature's last-known
// information at runtime.
func TestCompileAttachedDiesThatSubjectInterveningIf(t *testing.T) {
	t.Parallel()
	source := "When enchanted creature dies, if that creature was a Horror, return this card to its owner's hand."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.Condition == nil {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if ability.Trigger.Pattern.Event != TriggerEventPermanentDied ||
		ability.Trigger.Pattern.Source != TriggerSourceAttachedPermanent {
		t.Fatalf("pattern = %#v, want attached-permanent dies", ability.Trigger.Pattern)
	}
	condition := ability.Trigger.Condition
	if condition.Predicate != ConditionPredicateObjectMatches ||
		condition.ObjectBinding != ReferenceBindingEventPermanent {
		t.Fatalf("condition = %#v, want event-permanent object-match", condition)
	}
	if !slices.Equal(condition.Selection.SubtypesAny, []string{string(types.Horror)}) {
		t.Fatalf("selection = %#v, want Horror subtype", condition.Selection)
	}
}

func TestCompileNonPhaseTriggerUsesNormalizedSyntaxTokens(t *testing.T) {
	t.Parallel()
	source := "Whenever a  creature enters, draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := compilation.Abilities[0].Trigger
	if trigger.Event != "a  creature enters" {
		t.Fatalf("raw event = %q, want exact source metadata", trigger.Event)
	}
	if trigger.Pattern.Event != TriggerEventPermanentEnteredBattlefield {
		t.Fatalf("pattern = %#v, want normalized permanent-enter event", trigger.Pattern)
	}
}

func TestCompileSemanticTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, TriggerPattern)
	}{
		{
			name:   "source self",
			source: "When this creature enters, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Kind != TriggerWhen ||
					pattern.Event != TriggerEventPermanentEnteredBattlefield ||
					pattern.Source != TriggerSourceSelf {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "source self with capitalized subtype",
			source: "When this Aura enters, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Kind != TriggerWhen ||
					pattern.Event != TriggerEventPermanentEnteredBattlefield ||
					pattern.Source != TriggerSourceSelf {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "controller and subject Selection",
			source: "Whenever another nontoken creature you control enters, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Controller != ControllerYou ||
					!pattern.ExcludeSelf ||
					!pattern.SubjectSelection.NonToken ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "step and active player relation",
			source: "At the beginning of each opponent's draw step, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventBeginningOfStep ||
					pattern.Step != TriggerStepDraw ||
					pattern.Controller != ControllerOpponent {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "combat damage role",
			source: "Whenever this creature deals combat damage to a creature, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventDamageDealt ||
					pattern.Source != TriggerSourceSelf ||
					pattern.Subject != TriggerSubjectDamageSource ||
					pattern.CombatQualifier != TriggerCombatDamage ||
					pattern.DamageRecipient != TriggerDamageRecipientPermanent ||
					!slices.Equal(pattern.DamageRecipientSelection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "one or more",
			source: "Whenever one or more artifacts you control enter, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if !pattern.OneOrMore ||
					pattern.Controller != ControllerYou ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Artifact}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "attack subject",
			source: "Whenever a creature an opponent controls attacks, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventAttackerDeclared ||
					pattern.Controller != ControllerOpponent ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "attached blocker",
			source: "Whenever equipped creature blocks, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventBlockerDeclared ||
					pattern.Source != TriggerSourceAttachedPermanent {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "attached permanent dies whenever",
			source: "Whenever equipped creature dies, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Kind != TriggerWhenever ||
					pattern.Event != TriggerEventPermanentDied ||
					pattern.Source != TriggerSourceAttachedPermanent ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "tap controller relation",
			source: "Whenever another artifact you control becomes tapped, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventPermanentTapped ||
					pattern.Controller != ControllerYou ||
					!pattern.ExcludeSelf ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Artifact}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "untap subject",
			source: "Whenever a creature you control becomes untapped, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventPermanentUntapped ||
					pattern.Controller != ControllerYou {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "became target of spell",
			source: "Whenever this creature becomes the target of a spell, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventObjectBecameTarget ||
					pattern.Source != TriggerSourceSelf ||
					pattern.StackObject != TriggerStackObjectSpell {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "noncreature spell mana value less than source power",
			source: "Whenever an opponent casts a noncreature spell with mana value less than this creature's power, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventSpellCast ||
					pattern.Controller != ControllerOpponent ||
					!pattern.CardSelection.ManaValueLessThanSourcePower ||
					pattern.CardSelection.MatchManaValue ||
					!slices.Equal(pattern.CardSelection.ExcludedTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := compilation.Abilities[0].Trigger
			if trigger == nil || trigger.Event == "" {
				t.Fatalf("trigger = %#v", trigger)
			}
			eventText := test.source[trigger.Pattern.Span.Start.Offset:trigger.Pattern.Span.End.Offset]
			if eventText != trigger.Event {
				t.Fatalf("pattern span text = %q, raw diagnostic event = %q", eventText, trigger.Event)
			}
			test.check(t, trigger.Pattern)
		})
	}
}

func TestCompileSpellCastDisjunctionTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, TriggerPattern)
	}{
		{
			name:   "cast spell card-type disjunction",
			source: "Whenever you cast an artifact or enchantment spell, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventSpellCast ||
					!slices.Equal(pattern.CardSelection.RequiredTypesAny,
						[]types.Card{types.Artifact, types.Enchantment}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "cast spell subtype disjunction",
			source: "Whenever you cast an Aura, Equipment, or Vehicle spell, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventSpellCast ||
					!slices.Equal(pattern.CardSelection.SubtypesAny,
						[]types.Sub{types.Aura, types.Equipment, types.Vehicle}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := compilation.Abilities[0].Trigger
			if trigger == nil || trigger.Event == "" {
				t.Fatalf("trigger = %#v", trigger)
			}
			eventText := test.source[trigger.Pattern.Span.Start.Offset:trigger.Pattern.Span.End.Offset]
			if eventText != trigger.Event {
				t.Fatalf("pattern span text = %q, raw diagnostic event = %q", eventText, trigger.Event)
			}
			test.check(t, trigger.Pattern)
		})
	}
}

func TestCompileSelfOrAnotherTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever this creature or another Ally you control enters, draw a card.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := compilation.Abilities[0].Trigger
	if trigger == nil {
		t.Fatal("trigger = nil")
	}
	pattern := trigger.Pattern
	if pattern.Event != TriggerEventPermanentEnteredBattlefield ||
		pattern.Controller != ControllerYou ||
		!pattern.SubjectSelectionOrSelf ||
		pattern.ExcludeSelf ||
		pattern.Source != TriggerSourceAny ||
		!slices.Equal(pattern.SubjectSelection.SubtypesAny, []types.Sub{types.Sub("Ally")}) {
		t.Fatalf("pattern = %#v", pattern)
	}
}

// TestCompileSelfGraveyardOrAnotherTriggerPattern verifies the two-verb
// self-or-another battlefield-to-graveyard union "this creature dies or another
// <Selection> you control is put into a graveyard from the battlefield" (Scrap
// Trawler) compiles to a permanent zone-change pattern widened to the source via
// SubjectSelectionOrSelf, identical to the single-verb shared-subject form.
func TestCompileSelfGraveyardOrAnotherTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever this creature dies or another artifact you control is put into a graveyard from the battlefield, draw a card.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := compilation.Abilities[0].Trigger
	if trigger == nil {
		t.Fatal("trigger = nil")
	}
	pattern := trigger.Pattern
	if pattern.Event != TriggerEventZoneChanged ||
		pattern.Controller != ControllerYou ||
		!pattern.SubjectSelectionOrSelf ||
		pattern.ExcludeSelf ||
		pattern.Source != TriggerSourceAny ||
		!pattern.MatchFromZone || pattern.FromZone != TriggerZoneBattlefield ||
		!pattern.MatchToZone || pattern.ToZone != TriggerZoneGraveyard ||
		!slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("pattern = %#v", pattern)
	}
}

func TestCompileActionTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		check  func(*testing.T, TriggerPattern)
	}{
		{
			source: "Whenever a Forest an opponent controls becomes tapped, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventPermanentTapped ||
					pattern.Controller != ControllerOpponent ||
					!slices.Equal(pattern.SubjectSelection.SubtypesAny, []types.Sub{types.Forest}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "When this creature is turned face up, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Kind != TriggerWhen ||
					pattern.Event != TriggerEventPermanentTurnedFaceUp ||
					pattern.Source != TriggerSourceSelf {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever a creature you control becomes the target of a spell or ability an opponent controls, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventObjectBecameTarget ||
					pattern.Controller != ControllerYou ||
					pattern.CauseController != ControllerOpponent ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever a player cycles a card, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventCycled || pattern.Player != TriggerPlayerAny {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you sacrifice a Clue, you gain 3 life.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventPermanentSacrificed ||
					pattern.Player != TriggerPlayerYou ||
					!slices.Equal(pattern.SubjectSelection.SubtypesAny, []types.Sub{types.Clue}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you scry, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventScry || pattern.Player != TriggerPlayerYou {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you commit a crime, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventCrimeCommitted || pattern.Player != TriggerPlayerYou {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever an opponent commits a crime, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventCrimeCommitted || pattern.Player != TriggerPlayerOpponent {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever an opponent searches their library, you gain 1 life and draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventLibrarySearched || pattern.Player != TriggerPlayerOpponent {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever a player searches their library, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventLibrarySearched || pattern.Player != TriggerPlayerAny {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you create a token, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventTokenCreated ||
					pattern.UnionEvent != TriggerEventUnknown ||
					pattern.Player != TriggerPlayerYou ||
					!pattern.SubjectSelection.TokenOnly {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you create or sacrifice a token, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventTokenCreated ||
					pattern.UnionEvent != TriggerEventPermanentSacrificed ||
					pattern.Player != TriggerPlayerYou ||
					!pattern.SubjectSelection.TokenOnly {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever this creature attacks or becomes the target of a spell, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventObjectBecameTarget ||
					pattern.UnionEvent != TriggerEventAttackerDeclared ||
					pattern.Source != TriggerSourceSelf ||
					pattern.StackObject != TriggerStackObjectSpell {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever an opponent activates an ability of a creature or land that isn't a mana ability, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventAbilityActivated ||
					pattern.Player != TriggerPlayerOpponent ||
					!pattern.ExcludeManaAbility ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypesAny, []types.Card{types.Creature, types.Land}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever this creature attacks alone, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventAttackerDeclared ||
					pattern.Source != TriggerSourceSelf ||
					!pattern.AttackAlone {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you attack with two or more creatures, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventAttackerDeclared ||
					pattern.Controller != ControllerYou ||
					!pattern.OneOrMore ||
					pattern.AttackerCountAtLeast != 2 {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "At the beginning of your next upkeep, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventBeginningOfStep ||
					pattern.Step != TriggerStepUpkeep ||
					pattern.Controller != ControllerYou ||
					!pattern.NextOccurrence {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you gain life during your turn, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventLifeGained ||
					pattern.Player != TriggerPlayerYou ||
					pattern.CastDuringTurn != TriggerCastTurnYours {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you lose life during your turn, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventLifeLost ||
					pattern.Player != TriggerPlayerYou ||
					pattern.CastDuringTurn != TriggerCastTurnYours {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you gain life for the first time during each of your turns, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventLifeGained ||
					pattern.Player != TriggerPlayerYou ||
					pattern.PlayerEventOrdinalThisTurn != 1 ||
					pattern.CastDuringTurn != TriggerCastTurnYours {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever an opponent loses life for the first time during each of their turns, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventLifeLost ||
					pattern.Player != TriggerPlayerOpponent ||
					pattern.PlayerEventOrdinalThisTurn != 1 ||
					pattern.CastDuringTurn != TriggerCastTurnNotYours {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			test.check(t, compilation.Abilities[0].Trigger.Pattern)
		})
	}
}

func TestCompileEnterOrDiesUnionTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When this creature enters or dies, draw a card.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	pattern := compilation.Abilities[0].Trigger.Pattern
	if pattern.Event != TriggerEventPermanentEnteredBattlefield ||
		pattern.UnionEvent != TriggerEventPermanentDied ||
		pattern.Source != TriggerSourceSelf {
		t.Fatalf("pattern = %#v", pattern)
	}
}

func TestCompileSemanticTriggerPatternsFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever creature you control becomes tapped, draw a card.",
		"At the beginning of your declare attackers step, draw a card.",
		"At the beginning of the upkeep of enchanted permanent's controller, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if pattern := compilation.Abilities[0].Trigger.Pattern; pattern.Event != TriggerEventUnknown {
				t.Fatalf("near-miss pattern = %#v, want unknown event", pattern)
			}
		})
	}
}

// TestCompileBecameTargetFirstTimeEachTurn covers the Valiant ability word and
// the Glasskite spirits: the inline "for the first time each turn" ordinal on a
// became-target trigger compiles to the object-became-target event with a
// once-per-turn cap (MaxTriggersPerTurn == 1).
func TestCompileBecameTargetFirstTimeEachTurn(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		source    string
		wantCause ControllerKind
	}{
		{
			name:      "valiant you control",
			source:    "Valiant — Whenever this creature becomes the target of a spell or ability you control for the first time each turn, draw a card.",
			wantCause: ControllerYou,
		},
		{
			name:      "glasskite any controller",
			source:    "Whenever this creature becomes the target of a spell or ability for the first time each turn, draw a card.",
			wantCause: ControllerAny,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := compilation.Abilities[0].Trigger
			if trigger.Pattern.Event != TriggerEventObjectBecameTarget ||
				trigger.Pattern.Source != TriggerSourceSelf ||
				trigger.Pattern.CauseController != test.wantCause {
				t.Fatalf("pattern = %#v", trigger.Pattern)
			}
			if trigger.MaxTriggersPerTurn != 1 {
				t.Fatalf("MaxTriggersPerTurn = %d, want 1", trigger.MaxTriggersPerTurn)
			}
		})
	}
}

func TestCompilePlayerOrdinalTriggerPattern(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		source  string
		event   TriggerEvent
		player  TriggerPlayerRelation
		ordinal int
	}{
		{
			source:  "Whenever you draw your second card each turn, create a 2/2 black Zombie creature token.",
			event:   TriggerEventCardDrawn,
			player:  TriggerPlayerYou,
			ordinal: 2,
		},
		{
			source:  "When an opponent loses life for the first time each turn, draw a card.",
			event:   TriggerEventLifeLost,
			player:  TriggerPlayerOpponent,
			ordinal: 1,
		},
	} {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			pattern := compilation.Abilities[0].Trigger.Pattern
			if pattern.Event != test.event ||
				pattern.Player != test.player ||
				pattern.PlayerEventOrdinalThisTurn != test.ordinal {
				t.Fatalf("pattern = %#v", pattern)
			}
		})
	}
}

func TestCompileNamedSelfEnterTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When Example Card enters, draw a card.",
		pipelineContext{CardName: "Example Card"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	pattern := compilation.Abilities[0].Trigger.Pattern
	if pattern.Event != TriggerEventPermanentEnteredBattlefield || pattern.Source != TriggerSourceSelf {
		t.Fatalf("pattern = %#v", pattern)
	}
}

func TestCompileSemanticTriggerPatternReferencesInterveningCondition(t *testing.T) {
	t.Parallel()
	source := "When this creature enters, if it was kicked, draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := compilation.Abilities[0].Trigger
	if trigger.Condition == nil || trigger.Pattern.InterveningCondition != trigger.Condition {
		t.Fatalf("trigger = %#v, want pattern to reference source-spanned intervening condition", trigger)
	}
}

func TestCompileSagaChapterAbility(t *testing.T) {
	t.Parallel()
	source := "II, III — Draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{Saga: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityChapter || !slices.Equal(ability.Chapters, []int{2, 3}) {
		t.Fatalf("ability = %#v", ability)
	}
	if ability.AbilityWord != "" {
		t.Fatalf("ability word = %q, want empty", ability.AbilityWord)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
}

func TestCompileOptionalTriggeredAbility(t *testing.T) {
	t.Parallel()
	source := "When this creature enters, you may draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if !ability.Optional || source[ability.OptionalSpan.Start.Offset:ability.OptionalSpan.End.Offset] != "you may" {
		t.Fatalf("optional ability = %#v", ability)
	}
}

func TestCompileBecomesMonstrousEventXTargets(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When this creature becomes monstrous, goad up to X target creatures your opponents control.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil ||
		ability.Trigger.Pattern.Event != TriggerEventPermanentBecameMonstrous ||
		ability.Trigger.Pattern.Source != TriggerSourceSelf {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if len(ability.Content.Targets) != 1 {
		t.Fatalf("targets = %#v", ability.Content.Targets)
	}
	target := ability.Content.Targets[0]
	if target.Cardinality != (TargetCardinality{Min: 0, Max: 99, MaxEventX: true}) ||
		target.Selector.Kind != SelectorCreature ||
		target.Selector.Controller != ControllerOpponent {
		t.Fatalf("target = %#v", target)
	}
}
