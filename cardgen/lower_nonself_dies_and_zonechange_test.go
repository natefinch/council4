package cardgen

import (
	goparser "go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestLowerCreatureDiesRegressionStillWorks(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Death Drifter",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "When this creature dies, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerWhen {
		t.Fatalf("trigger type = %v, want TriggerWhen", trigger.Trigger.Type)
	}
	if trigger.Trigger.Pattern.Event != game.EventPermanentDied {
		t.Fatalf("trigger event = %v, want EventPermanentDied", trigger.Trigger.Pattern.Event)
	}
}

// TestLowerGainControlForAsLongAsYouControlSourceCardName checks that an
// activated ability whose body is "Gain control of target creature for as long
// as you control [CardName]." lowers to DurationForAsLongAsYouControlSource.
func TestLowerGainControlForAsLongAsYouControlSourceCardName(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Merieke Ri Berit",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Wizard",
		OracleText: "{T}: Gain control of target creature for as long as you control Merieke Ri Berit.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationForAsLongAsYouControlSource)
}

// TestLowerGainControlForAsLongAsYouControlThis checks the "for as long as
// you control this [type]" self-referential variant.
func TestLowerGainControlForAsLongAsYouControlThis(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Control Source",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Wizard",
		OracleText: "{T}: Gain control of target creature for as long as you control this creature.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationForAsLongAsYouControlSource)
}

// TestLowerGainControlAsLongAsSourceOnBattlefield checks the "as long as this
// [type] remains on the battlefield" variant for single-effect spells.
func TestLowerGainControlAsLongAsSourceOnBattlefield(t *testing.T) {
	t.Parallel()
	// Simulate an enchantment that gives control as long as it's on the
	// battlefield, represented as a loyalty ability for simplicity.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Control Aura",
		Layout:     "normal",
		TypeLine:   "Planeswalker — Test",
		OracleText: "−2: Gain control of target creature as long as this planeswalker remains on the battlefield.",
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("loyalty abilities = %d, want 1", len(face.LoyaltyAbilities))
	}
	mode := face.LoyaltyAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationForAsLongAsSourceOnBattlefield)
}

// TestGenerateExecutableCardSourceGainControlForAsLongAsYouControlRenders
// verifies that the rendered Go source contains DurationForAsLongAsYouControlSource.
func TestGenerateExecutableCardSourceGainControlForAsLongAsYouControlRenders(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Merieke Ri Berit",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Wizard",
		OracleText: "{T}: Gain control of target creature for as long as you control Merieke Ri Berit.",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "merieke.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.ApplyContinuous",
		"game.LayerControl",
		"NewController: opt.Val(game.Player1)",
		"game.DurationForAsLongAsYouControlSource",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceGainControlSourceOnBattlefieldRenders
// verifies that the rendered Go source contains DurationForAsLongAsSourceOnBattlefield.
func TestGenerateExecutableCardSourceGainControlSourceOnBattlefieldRenders(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Control Aura",
		Layout:     "normal",
		TypeLine:   "Planeswalker — Test",
		OracleText: "−2: Gain control of target creature as long as this planeswalker remains on the battlefield.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "control_aura.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.ApplyContinuous",
		"game.LayerControl",
		"NewController: opt.Val(game.Player1)",
		"game.DurationForAsLongAsSourceOnBattlefield",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerGainControlAttachmentDuration verifies that the attachment-dependent
// "for as long as that creature is enchanted" wording (Rootwater Matriarch)
// lowers to a control grant with the enchanted-state duration.
func TestLowerGainControlAttachmentDuration(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Rootwater Matriarch",
		Layout:     "normal",
		TypeLine:   "Creature — Merfolk",
		OracleText: "{T}: Gain control of target creature for as long as that creature is enchanted.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationForAsLongAsControlledCreatureEnchanted)
}

// TestLowerGainControlAttachmentDurationRejectsUnsupported ensures that other
// attachment-dependent duration wordings beyond the supported phrase remain
// fail-closed.
func TestLowerGainControlAttachmentDurationRejectsUnsupported(t *testing.T) {
	t.Parallel()
	// "for as long as that Aura is attached to it" (Eriette-style) is an
	// attachment-source duration that is not modeled and must stay unsupported.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Aura Thief",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "{T}: Gain control of target creature for as long as that Aura is attached to it.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported attachment duration, got none")
	}
}

// TestLowerNonSelfDiesTriggerAnotherCreatureYouControl verifies the main
// happy-path non-self dies trigger phrase.
func TestLowerNonSelfDiesTriggerAnotherCreatureYouControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Death Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever another creature you control dies, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerWhenever {
		t.Fatalf("trigger type = %v, want TriggerWhenever", trigger.Trigger.Type)
	}
	pat := trigger.Trigger.Pattern
	if pat.Event != game.EventPermanentDied {
		t.Fatalf("event = %v, want EventPermanentDied", pat.Event)
	}
	if pat.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", pat.Controller)
	}
	if !pat.ExcludeSelf {
		t.Fatal("ExcludeSelf = false, want true")
	}
	wantTypes := []types.Card{types.Creature}
	if !reflect.DeepEqual(pat.SubjectSelection.RequiredTypes, wantTypes) {
		t.Fatalf("SubjectSelection.RequiredTypes = %v, want %v", pat.SubjectSelection.RequiredTypes, wantTypes)
	}
	// Verify the body lowers to a draw effect.
	if len(trigger.Content.Modes) == 0 || len(trigger.Content.Modes[0].Sequence) == 0 {
		t.Fatal("expected non-empty body content")
	}
	if _, ok := trigger.Content.Modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("body primitive = %T, want game.Draw", trigger.Content.Modes[0].Sequence[0].Primitive)
	}
}

// TestLowerNonSelfDiesTriggerEnchantedCreature verifies the attached-permanent
// (enchanted creature) trigger phrase.
func TestLowerNonSelfDiesTriggerEnchantedCreature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Elegy Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nWhen enchanted creature dies, draw a card.",
		Power:      nil,
		Toughness:  nil,
	})
	var ta *game.TriggeredAbility
	for i := range face.TriggeredAbilities {
		if strings.Contains(face.TriggeredAbilities[i].Text, "enchanted creature dies") {
			ta = &face.TriggeredAbilities[i]
		}
	}
	if ta == nil {
		t.Fatal("enchanted-creature-dies triggered ability not lowered")
	}
	if ta.Trigger.Type != game.TriggerWhen {
		t.Fatalf("trigger type = %v, want TriggerWhen", ta.Trigger.Type)
	}
	pat := ta.Trigger.Pattern
	if pat.Event != game.EventPermanentDied {
		t.Fatalf("event = %v, want EventPermanentDied", pat.Event)
	}
	if pat.Source != game.TriggerSourceAttachedPermanent {
		t.Fatalf("source = %v, want TriggerSourceAttachedPermanent", pat.Source)
	}
	wantTypes := []types.Card{types.Creature}
	if !reflect.DeepEqual(pat.SubjectSelection.RequiredTypes, wantTypes) {
		t.Fatalf("SubjectSelection.RequiredTypes = %v, want %v", pat.SubjectSelection.RequiredTypes, wantTypes)
	}
}

// TestLowerNonSelfDiesTriggerUnsupportedControllerDamageFailsClosed verifies
// that a bound source reference does not make unsupported player damage valid.
func TestLowerNonSelfDiesTriggerUnsupportedControllerDamageFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Damage Dealer",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever a creature dies, this creature deals 1 damage to its controller.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for pronoun reference to dying permanent")
	}
	if !strings.Contains(diagnostics[0].Summary, "unsupported damage spell") {
		t.Fatalf("diagnostic summary = %q, want 'unsupported damage spell'", diagnostics[0].Summary)
	}
}

// TestLowerNonSelfDiesTriggerUnrecognisedPhraseFailsClosed verifies that an
// unrecognised trigger phrase produces a fail-closed diagnostic.
func TestLowerNonSelfDiesTriggerUnrecognisedPhraseFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Haunting Creature",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: "Whenever the haunted creature dies, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported dies trigger diagnostic for unrecognised phrase")
	}
	found := false
	for _, d := range diagnostics {
		if strings.Contains(d.Summary, "unsupported") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no unsupported diagnostic found in: %v", diagnostics)
	}
}

// TestLowerNonSelfDiesTriggerACreatureDies verifies "a creature dies".
func TestLowerNonSelfDiesTriggerACreatureDies(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Morbid Counter",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever a creature dies, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	pat := face.TriggeredAbilities[0].Trigger.Pattern
	if pat.Event != game.EventPermanentDied {
		t.Fatalf("event = %v, want EventPermanentDied", pat.Event)
	}
	if pat.Controller != game.TriggerControllerAny {
		t.Fatalf("controller = %v, want TriggerControllerAny", pat.Controller)
	}
	if pat.ExcludeSelf {
		t.Fatal("ExcludeSelf = true, want false for 'a creature dies'")
	}
	if !reflect.DeepEqual(pat.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("SubjectSelection.RequiredTypes = %v", pat.SubjectSelection.RequiredTypes)
	}
}

// TestLowerNonSelfDiesTriggerNontokenCreatureYouControl verifies the nontoken
// creature trigger.
func TestLowerNonSelfDiesTriggerNontokenCreatureYouControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Soul Collector",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever a nontoken creature you control dies, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	pat := face.TriggeredAbilities[0].Trigger.Pattern
	if pat.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", pat.Controller)
	}
	if !pat.SubjectSelection.NonToken {
		t.Fatal("SubjectSelection.NonToken = false, want true")
	}
}

func TestLowerNonSelfDiesTriggerInterveningIfFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Life Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever a creature you control dies, if you have 5 or more life, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("intervening-if non-self dies trigger unexpectedly lowered")
	}
	if !strings.Contains(diagnostics[0].Detail, "does not support this semantic permanent zone-change trigger condition") {
		t.Fatalf("diagnostic = %#v, want intervening-if detail", diagnostics[0])
	}
}

// TestLowerNonSelfDiesSemanticPatterns verifies that every recognized dies
// pattern passes through the shared semantic trigger-pattern lowerer.
func TestLowerNonSelfDiesSemanticPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase      string
		wantSource  game.TriggerSourceFilter
		wantCtrl    game.TriggerControllerFilter
		excludeSelf bool
		wantTypes   []types.Card
		nonToken    bool
		wantKind    game.TriggerType
	}{
		{"enchanted creature dies", game.TriggerSourceAttachedPermanent, game.TriggerControllerAny, false, []types.Card{types.Creature}, false, game.TriggerWhen},
		{"equipped creature dies", game.TriggerSourceAttachedPermanent, game.TriggerControllerAny, false, []types.Card{types.Creature}, false, game.TriggerWhen},
		{"enchanted land dies", game.TriggerSourceAttachedPermanent, game.TriggerControllerAny, false, []types.Card{types.Land, types.Creature}, false, game.TriggerWhen},
		{"enchanted permanent dies", game.TriggerSourceAttachedPermanent, game.TriggerControllerAny, false, []types.Card{types.Creature}, false, game.TriggerWhen},
		{"another creature dies", game.TriggerSourceAny, game.TriggerControllerAny, true, []types.Card{types.Creature}, false, game.TriggerWhenever},
		{"another creature you control dies", game.TriggerSourceAny, game.TriggerControllerYou, true, []types.Card{types.Creature}, false, game.TriggerWhenever},
		{"a creature dies", game.TriggerSourceAny, game.TriggerControllerAny, false, []types.Card{types.Creature}, false, game.TriggerWhenever},
		{"a creature you control dies", game.TriggerSourceAny, game.TriggerControllerYou, false, []types.Card{types.Creature}, false, game.TriggerWhenever},
		{"a creature an opponent controls dies", game.TriggerSourceAny, game.TriggerControllerOpponent, false, []types.Card{types.Creature}, false, game.TriggerWhenever},
		{"a nontoken creature you control dies", game.TriggerSourceAny, game.TriggerControllerYou, false, []types.Card{types.Creature}, true, game.TriggerWhenever},
		{"another nontoken creature you control dies", game.TriggerSourceAny, game.TriggerControllerYou, true, []types.Card{types.Creature}, true, game.TriggerWhenever},
		{"another nontoken creature dies", game.TriggerSourceAny, game.TriggerControllerAny, true, []types.Card{types.Creature}, true, game.TriggerWhenever},
		{"a nontoken creature an opponent controls dies", game.TriggerSourceAny, game.TriggerControllerOpponent, false, []types.Card{types.Creature}, true, game.TriggerWhenever},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			kind := "Whenever "
			if tc.wantKind == game.TriggerWhen {
				kind = "When "
			}
			compilation, diagnostics := compileTestOracle(kind+tc.phrase+", draw a card.", parser.Context{}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := compilation.Abilities[0].Trigger
			pattern, ok := lowerTriggerPattern(&trigger.Pattern)
			if !ok {
				t.Fatalf("lowerTriggerPattern(%q) returned ok=false", tc.phrase)
			}
			triggerType, ok := lowerTriggerKind(trigger.Pattern.Kind)
			if !ok || triggerType != tc.wantKind {
				t.Errorf("triggerType = %v, %v, want %v, true", triggerType, ok, tc.wantKind)
			}
			if pattern.Source != tc.wantSource {
				t.Errorf("source = %v, want %v", pattern.Source, tc.wantSource)
			}
			if pattern.Controller != tc.wantCtrl {
				t.Errorf("controller = %v, want %v", pattern.Controller, tc.wantCtrl)
			}
			if pattern.ExcludeSelf != tc.excludeSelf {
				t.Errorf("ExcludeSelf = %v, want %v", pattern.ExcludeSelf, tc.excludeSelf)
			}
			if !reflect.DeepEqual(pattern.SubjectSelection.RequiredTypes, tc.wantTypes) {
				t.Errorf("SubjectSelection.RequiredTypes = %v, want %v", pattern.SubjectSelection.RequiredTypes, tc.wantTypes)
			}
			if pattern.SubjectSelection.NonToken != tc.nonToken {
				t.Errorf("SubjectSelection.NonToken = %v, want %v", pattern.SubjectSelection.NonToken, tc.nonToken)
			}
		})
	}
}

func TestLowerNonSelfDiesUnknownSemanticPatternReturnsFalse(t *testing.T) {
	t.Parallel()
	unknownPhrases := []string{
		"the haunted creature dies",
		"a madeup dies",
	}
	for _, phrase := range unknownPhrases {
		compilation, diagnostics := compileTestOracle("Whenever "+phrase+", draw a card.", parser.Context{}, compiler.Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("diagnostics = %#v", diagnostics)
		}
		_, ok := lowerTriggerPattern(&compilation.Abilities[0].Trigger.Pattern)
		if ok {
			t.Errorf("lowerTriggerPattern(%q) returned ok=true, want false", phrase)
		}
	}
}

func TestLowerPermanentZoneChangeSemanticPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase string
		want   game.TriggerPattern
	}{
		{
			phrase: "Whenever an artifact is returned to your hand, draw a card.",
			want: game.TriggerPattern{
				Event:         game.EventZoneChanged,
				Player:        game.TriggerPlayerYou,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
				MatchToZone:   true,
				ToZone:        zone.Hand,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Artifact},
				},
			},
		},
		{
			phrase: "Whenever another nontoken legendary green Dragon you control with power 4 or greater enters, draw a card.",
			want: game.TriggerPattern{
				Event:       game.EventPermanentEnteredBattlefield,
				Controller:  game.TriggerControllerYou,
				ExcludeSelf: true,
				SubjectSelection: game.Selection{
					Supertypes:  []types.Super{types.Legendary},
					SubtypesAny: []types.Sub{types.Dragon},
					ColorsAny:   []color.Color{color.Green},
					Power:       opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
					NonToken:    true,
				},
			},
		},
		{
			phrase: "Whenever one or more creatures are put into a graveyard from the battlefield, draw a card.",
			want: game.TriggerPattern{
				Event:         game.EventZoneChanged,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
				MatchToZone:   true,
				ToZone:        zone.Graveyard,
				OneOrMore:     true,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			phrase: "Whenever a creature leaves the battlefield without dying, draw a card.",
			want: game.TriggerPattern{
				Event:         game.EventZoneChanged,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
				ExcludeToZone: true,
				ToZone:        zone.Graveyard,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			phrase: "Whenever a face-down attacking creature dies, draw a card.",
			want: game.TriggerPattern{
				Event:         game.EventPermanentDied,
				MatchFaceDown: true,
				FaceDown:      true,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					CombatState:   game.CombatStateAttacking,
				},
			},
		},
		{
			phrase: "Whenever a creature card is put into your graveyard from anywhere, draw a card.",
			want: game.TriggerPattern{
				Event:       game.EventZoneChanged,
				Player:      game.TriggerPlayerYou,
				MatchToZone: true,
				ToZone:      zone.Graveyard,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.phrase, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileTestOracle(test.phrase, parser.Context{}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			got, ok := lowerTriggerPattern(&compilation.Abilities[0].Trigger.Pattern)
			if !ok || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("pattern = %#v, %v, want %#v, true", got, ok, test.want)
			}
		})
	}
}
