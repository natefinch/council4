package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerSurveilSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Surveil",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Surveil 2. (Look at the top two cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)",
	})
	mode := face.SpellAbility.Val.Modes[0]
	surveil, ok := mode.Sequence[0].Primitive.(game.Surveil)
	if !ok ||
		surveil.Amount.Value() != 2 ||
		surveil.Player != game.ControllerReference() {
		t.Fatalf("primitive = %+v, want controller surveils two", mode.Sequence[0].Primitive)
	}
}

func TestLowerInvestigateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Investigate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Investigate.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	investigate, ok := mode.Sequence[0].Primitive.(game.Investigate)
	if !ok || investigate.Amount.Value() != 1 {
		t.Fatalf("primitive = %+v, want investigate once", mode.Sequence[0].Primitive)
	}
}

func TestLowerInvestigateTwiceSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Investigate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Investigate twice.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	investigate, ok := mode.Sequence[0].Primitive.(game.Investigate)
	if !ok || investigate.Amount.Value() != 2 {
		t.Fatalf("primitive = %+v, want investigate twice", mode.Sequence[0].Primitive)
	}
}

func TestLowerProliferateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Proliferate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Proliferate.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if _, ok := mode.Sequence[0].Primitive.(game.Proliferate); !ok {
		t.Fatalf("primitive = %T, want game.Proliferate", mode.Sequence[0].Primitive)
	}
}

func TestLowerProliferateTwiceSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Proliferate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Proliferate twice.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	proliferate, ok := mode.Sequence[0].Primitive.(game.Proliferate)
	if !ok || proliferate.Amount.Value() != 2 {
		t.Fatalf("primitive = %+v, want proliferate twice", mode.Sequence[0].Primitive)
	}
}

func TestLowerExploreSourcePermanentTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Scout",
		Layout:     "normal",
		TypeLine:   "Creature — Merfolk Scout",
		OracleText: "When this creature enters, it explores.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	explore, ok := mode.Sequence[0].Primitive.(game.Explore)
	if !ok || explore.Creature.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("primitive = %+v, want event permanent explores", mode.Sequence[0].Primitive)
	}
}

func TestLowerModifyPTEventPermanentTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Guide",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever another creature enters, it gets +2/+0 until end of turn.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok || modify.Object != game.EventPermanentReference() {
		t.Fatalf("primitive = %+v, want event permanent P/T modification", mode.Sequence[0].Primitive)
	}
}

// TestLowerModifyPTEventPermanentInNonZoneChangeTrigger verifies that the
// shared lowerFixedModifyPTSpell path lowers an EventPermanent ModifyPT body
// across non-zone-change trigger shells (generic pattern trigger here).
func TestLowerModifyPTEventPermanentInNonZoneChangeTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "attack trigger it gets",
			oracle: "Whenever a creature attacks, it gets +1/+1 until end of turn.",
		},
		{
			name:   "tapped trigger it gets",
			oracle: "Whenever a creature becomes tapped, it gets +0/+2 until end of turn.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Pump",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
			if !ok || modify.Object != game.EventPermanentReference() {
				t.Fatalf("primitive = %+v, want event permanent P/T modification", mode.Sequence[0].Primitive)
			}
		})
	}
}

// TestLowerModifyPTEventPermanentSharedContentPathRegression verifies that the
// exact ETB non-self "It gets +2/+0 until end of turn." body still lowers
// correctly now that lowerEventPermanentModifyPTBody is removed and the shared
// lowerFixedModifyPTSpell path owns it.
func TestLowerModifyPTEventPermanentSharedContentPathRegression(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Guide",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever another creature enters, it gets +2/+0 until end of turn.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	modify, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modify.Object != game.EventPermanentReference() {
		t.Fatalf("Object = %v, want EventPermanentReference", modify.Object)
	}
	if modify.PowerDelta.Value() != 2 || modify.ToughnessDelta.Value() != 0 {
		t.Fatalf("P/T = %v/%v, want +2/+0", modify.PowerDelta, modify.ToughnessDelta)
	}
	if modify.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("Duration = %v, want DurationUntilEndOfTurn", modify.Duration)
	}
}

// TestLowerModifyPTEventPermanentFailsClosed verifies that EventPermanent
// ModifyPT bodies fail closed when the text does not match the expected form.
func TestLowerModifyPTEventPermanentFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		// Wrong duration.
		"Whenever a creature attacks, it gets +1/+1.",
		// Negated form.
		"Whenever a creature attacks, it doesn't get +1/+1 until end of turn.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Pump",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected diagnostic for unsupported body %q", oracleText)
			}
		})
	}
}

func TestLowerExploreRejectsUnsupportedTargets(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Explore",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature explores.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported explore diagnostic")
	}
}

// TestLowerKeywordAbilityGainLossNotLifeSpell pins issue #499: gain/lose spells
// whose object is a keyword or quoted ability must report a specific keyword/
// ability diagnostic, never the misleading "unsupported life spell".
// TestLowerTemporaryKeywordLossSpell verifies the keyword-removal lowering
// emits an ability-layer continuous effect with RemoveKeywords over the affected
// subject (a controlled/opponent group or a single targeted permanent).
func TestLowerTemporaryKeywordLossSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		oracle   string
		group    bool
		keywords []game.Keyword
	}{
		{
			name:     "opponent permanents",
			oracle:   "Permanents your opponents control lose hexproof and indestructible until end of turn.",
			group:    true,
			keywords: []game.Keyword{game.Hexproof, game.Indestructible},
		},
		{
			name:     "controlled permanents",
			oracle:   "Permanents you control lose hexproof until end of turn.",
			group:    true,
			keywords: []game.Keyword{game.Hexproof},
		},
		{
			name:     "single target",
			oracle:   "Target creature loses flying until end of turn.",
			group:    false,
			keywords: []game.Keyword{game.Flying},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Loss",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: tc.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
			if !ok {
				t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
			}
			effect := apply.ContinuousEffects[0]
			if effect.Layer != game.LayerAbility {
				t.Fatalf("layer = %v, want LayerAbility", effect.Layer)
			}
			if !reflect.DeepEqual(effect.RemoveKeywords, tc.keywords) {
				t.Fatalf("RemoveKeywords = %v, want %v", effect.RemoveKeywords, tc.keywords)
			}
			if tc.group {
				if effect.Group.Empty() {
					t.Fatal("group loss missing affected group")
				}
				if len(mode.Targets) != 0 {
					t.Fatalf("group loss has %d targets, want 0", len(mode.Targets))
				}
			} else if len(mode.Targets) != 1 {
				t.Fatalf("target loss has %d targets, want 1", len(mode.Targets))
			}
		})
	}
}

func TestLowerKeywordAbilityGainLossNotLifeSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		oracle string
		want   string
	}{
		{"Target creature gains shadow until end of turn.", "unsupported keyword or ability grant"},
		{"Lands you control gain \"{T}: Add one mana of any color.\"", "unsupported keyword or ability grant"},
		{"Target creature loses your choice of flying, first strike, or trample until end of turn.", "unsupported keyword or ability loss"},
		{"Target creature loses protection from black until end of turn.", "unsupported keyword or ability loss"},
	}
	for _, test := range tests {
		t.Run(test.oracle, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Keyword",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected diagnostic for %q", test.oracle)
			}
			for _, d := range diagnostics {
				if d.Summary == "unsupported life spell" {
					t.Fatalf("got misleading life-spell diagnostic for %q", test.oracle)
				}
			}
			if diagnostics[0].Summary != test.want {
				t.Fatalf("summary = %q, want %q", diagnostics[0].Summary, test.want)
			}
		})
	}
}

// TestLowerTargetProtectionGrantSpell verifies that a spell granting protection
// from a fixed color or from a color chosen on resolution to a single target
// creature lowers to an anchored ApplyContinuous carrying the protection static
// ability via AddAbilities (the Gods Willing / Mother of Runes shape).
func TestLowerTargetProtectionGrantSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracle      string
		chosenColor bool
		fromColor   color.Color
	}{
		{
			name:      "fixed color",
			oracle:    "Target creature you control gains protection from red until end of turn.",
			fromColor: color.Red,
		},
		{
			name:        "chosen color",
			oracle:      "Target creature you control gains protection from the color of your choice until end of turn.",
			chosenColor: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Protection Grant",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
			if !ok {
				t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
			}
			if apply.Duration != game.DurationUntilEndOfTurn {
				t.Fatalf("duration = %v, want until end of turn", apply.Duration)
			}
			effect := apply.ContinuousEffects[0]
			if len(effect.AddAbilities) != 1 {
				t.Fatalf("abilities = %d, want 1 granted protection ability", len(effect.AddAbilities))
			}
			static, ok := effect.AddAbilities[0].(*game.StaticAbility)
			if !ok {
				t.Fatalf("ability = %T, want *game.StaticAbility", effect.AddAbilities[0])
			}
			prot, ok := game.StaticBodyProtectionKeyword(static)
			if !ok {
				t.Fatalf("body = %+v, want protection keyword", static)
			}
			if tc.chosenColor {
				if !prot.ChosenColor {
					t.Fatalf("protection = %+v, want chosen color", prot)
				}
			} else if len(prot.FromColors) != 1 || prot.FromColors[0] != tc.fromColor {
				t.Fatalf("protection = %+v, want from %v", prot, tc.fromColor)
			}
		})
	}
}

func TestLowerManifestSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Manifest",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Manifest the top card of your library.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	manifest, ok := mode.Sequence[0].Primitive.(game.Manifest)
	if !ok {
		t.Fatalf("primitive = %T, want game.Manifest", mode.Sequence[0].Primitive)
	}
	if manifest.Dread {
		t.Fatal("basic manifest lowered with Dread=true")
	}
}

func TestLowerManifestDreadSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "shorthand",
			oracle: "Manifest Dread.",
		},
		{
			name:   "long form",
			oracle: "Look at the top two cards of your library. Put one of them onto the battlefield face down as a 2/2 creature. Put the other into your graveyard.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Manifest Dread",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			manifest, ok := mode.Sequence[0].Primitive.(game.Manifest)
			if !ok {
				t.Fatalf("primitive = %T, want game.Manifest", mode.Sequence[0].Primitive)
			}
			if !manifest.Dread {
				t.Fatal("manifest dread lowered with Dread=false")
			}
		})
	}
}

func TestLowerRemovalManifestSequence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		dread     bool
		removalIs func(game.Primitive) bool
	}{
		{
			name:   "exile then manifest top card",
			oracle: "Exile target creature. Its controller manifests the top card of their library.",
			dread:  false,
			removalIs: func(p game.Primitive) bool {
				_, ok := p.(game.Exile)
				return ok
			},
		},
		{
			name:   "destroy then manifest dread",
			oracle: "Destroy target creature. Its controller manifests dread.",
			dread:  true,
			removalIs: func(p game.Primitive) bool {
				_, ok := p.(game.Destroy)
				return ok
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Removal Manifest",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Sequence) != 2 {
				t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
			}
			if !test.removalIs(mode.Sequence[0].Primitive) {
				t.Fatalf("removal primitive = %T", mode.Sequence[0].Primitive)
			}
			manifest, ok := mode.Sequence[1].Primitive.(game.Manifest)
			if !ok {
				t.Fatalf("primitive = %T, want game.Manifest", mode.Sequence[1].Primitive)
			}
			if manifest.Dread != test.dread {
				t.Fatalf("manifest Dread = %v, want %v", manifest.Dread, test.dread)
			}
			if manifest.Player.Kind() != game.PlayerReferenceObjectController {
				t.Fatalf("manifest Player kind = %v, want object controller", manifest.Player.Kind())
			}
			object, ok := manifest.Player.Object()
			if !ok || object.Kind() != game.ObjectReferenceTargetPermanent || object.TargetIndex() != 0 {
				t.Fatalf("manifest Player object = %#v", object)
			}
		})
	}
}

func TestLowerManifestRejectsUnsupportedPatterns(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Manifest",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Manifest a card from your hand.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported manifest diagnostic")
	}
}

func TestLowerInterveningTriggerUtilityKeywordBodies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		text      string
		primitive any
	}{
		{
			name:      "scry",
			text:      "When this creature enters, if you control an artifact, scry 2.",
			primitive: game.Scry{Amount: game.Fixed(2), Player: game.ControllerReference()},
		},
		{
			name:      "investigate",
			text:      "When this creature enters, if you control an artifact, investigate.",
			primitive: game.Investigate{Amount: game.Fixed(1)},
		},
		{
			name:      "proliferate",
			text:      "When this creature enters, if you control an artifact, proliferate.",
			primitive: game.Proliferate{Amount: game.Fixed(1)},
		},
		{
			name:      "explore",
			text:      "When this creature enters, if you control an artifact, it explores.",
			primitive: game.Explore{Creature: game.EventPermanentReference()},
		},
		{
			name:      "manifest",
			text:      "When this creature enters, if you control an artifact, manifest the top card of your library.",
			primitive: game.Manifest{},
		},
		{
			name:      "mill",
			text:      "When this creature enters, if you control an artifact, mill two cards.",
			primitive: game.Mill{Amount: game.Fixed(2), Player: game.ControllerReference()},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Utility",
				Layout:     "normal",
				TypeLine:   "Creature — Human Wizard",
				OracleText: tc.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			got := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
			if !reflect.DeepEqual(got, tc.primitive) {
				t.Fatalf("primitive = %+v, want %+v", got, tc.primitive)
			}
		})
	}
}

func TestLowerVariableMillSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mill",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Mill X cards, where X is the number of creatures you control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	mill, ok := mode.Sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("primitive = %T, want game.Mill", mode.Sequence[0].Primitive)
	}
	dynamic := mill.Amount.DynamicAmount()
	if !dynamic.Exists {
		t.Fatalf("mill amount = %+v, want dynamic controlled creature count", mill.Amount)
	}
	selection := dynamic.Val.Group.Selection()
	if dynamic.Val.Kind != game.DynamicAmountCountSelector ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature ||
		selection.Controller != game.ControllerYou {
		t.Fatalf("mill amount = %+v, want dynamic controlled creature count", mill.Amount)
	}
}
