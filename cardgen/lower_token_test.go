package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerMultiColorMultiSubtypeToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Multi",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 blue and white Bird creature token.",
		Colors:     []string{"W", "U"},
	})
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	def, _ := create.Source.TokenDefRef()
	if len(def.Colors) != 2 || def.Colors[0] != color.Blue || def.Colors[1] != color.White {
		t.Fatalf("token colors = %v, want [Blue White]", def.Colors)
	}
	if def.Name != "Bird" {
		t.Fatalf("token name = %q, want Bird", def.Name)
	}
}

func TestLowerMultipleCreatureTokens(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tokens",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create two 1/1 white Soldier creature tokens.",
		Colors:     []string{"W"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if create.Amount.Value() != 2 {
		t.Fatalf("amount = %d, want 2", create.Amount.Value())
	}
}

func TestLowerForEachTokenCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test ForEach Battlefield",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "For each Shrine you control, create a 1/1 white Monk creature token.",
		Colors:     []string{"W"},
	})
	create := createTokenPrimitive(t, face)
	if !create.Amount.IsDynamic() {
		t.Fatalf("amount = %d, want a dynamic per-Shrine count", create.Amount.Value())
	}
	want := game.DynamicAmount{
		Kind:       game.DynamicAmountCountSelector,
		Multiplier: 1,
		Group: game.BattlefieldGroup(game.Selection{
			SubtypesAny: []types.Sub{types.Sub("Shrine")},
			Controller:  game.ControllerYou,
		}),
	}
	if got := create.Amount.DynamicAmount().Val; !reflect.DeepEqual(got, want) {
		t.Fatalf("dynamic amount = %+v, want %+v", got, want)
	}
}

func TestLowerForEachGraveyardCardCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test ForEach Graveyard",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "For each creature card in your graveyard, create a 1/1 white Soldier creature token.",
		Colors:     []string{"W"},
	})
	create := createTokenPrimitive(t, face)
	if !create.Amount.IsDynamic() {
		t.Fatalf("amount = %d, want a dynamic per-card count", create.Amount.Value())
	}
	player := game.ControllerReference()
	want := game.DynamicAmount{
		Kind:       game.DynamicAmountCountCardsInZone,
		Multiplier: 1,
		Player:     &player,
		CardZone:   zone.Graveyard,
		Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}
	if got := create.Amount.DynamicAmount().Val; !reflect.DeepEqual(got, want) {
		t.Fatalf("dynamic amount = %+v, want %+v", got, want)
	}
}

// TestLowerTrailingForEachLandTokenCount verifies that a create-token whose
// count is a TRAILING "for each <permanent> you control" phrase (Avenger of
// Zendikar) lowers to a dynamic per-permanent count without folding the counted
// permanent's type into the token's own type line.
func TestLowerTrailingForEachLandTokenCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Trailing ForEach Land",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 0/1 green Plant creature token for each land you control.",
		Colors:     []string{"G"},
	})
	create := createTokenPrimitive(t, face)
	if !create.Amount.IsDynamic() {
		t.Fatalf("amount = %d, want a dynamic per-land count", create.Amount.Value())
	}
	want := game.DynamicAmount{
		Kind:       game.DynamicAmountCountSelector,
		Multiplier: 1,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Land},
			Controller:    game.ControllerYou,
		}),
	}
	if got := create.Amount.DynamicAmount().Val; !reflect.DeepEqual(got, want) {
		t.Fatalf("dynamic amount = %+v, want %+v", got, want)
	}
}

func createTokenPrimitive(t *testing.T, face loweredFaceAbilities) game.CreateToken {
	t.Helper()
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	return create
}

// TestLowerNamedTokenChoice verifies that "create a X token or a Y token" and
// the N-way "create your choice of a X token, a Y token, or a Z token" forms
// lower to a choose-one modal ability with one CreateToken mode per predefined
// artifact-token alternative.
func TestLowerNamedTokenChoice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
		want   []string
	}{
		{
			name:   "two-way",
			oracle: "Create a Food token or a Treasure token.",
			want:   []string{string(types.Food), string(types.Treasure)},
		},
		{
			name:   "three-way choice of",
			oracle: "Create your choice of a Clue token, a Food token, or a Treasure token.",
			want:   []string{string(types.Clue), string(types.Food), string(types.Treasure)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Provisioner",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability not lowered")
			}
			content := face.SpellAbility.Val
			if len(content.Modes) != len(test.want) || content.MinModes != 1 || content.MaxModes != 1 {
				t.Fatalf("modal shape = modes %d min %d max %d, want %d/1/1",
					len(content.Modes), content.MinModes, content.MaxModes, len(test.want))
			}
			for i, mode := range content.Modes {
				create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
				if !ok {
					t.Fatalf("mode %d primitive = %T, want game.CreateToken", i, mode.Sequence[0].Primitive)
				}
				if create.Amount.Value() != 1 {
					t.Fatalf("mode %d amount = %d, want 1", i, create.Amount.Value())
				}
				def, ok := create.Source.TokenDefRef()
				if !ok || def.Name != test.want[i] {
					t.Fatalf("mode %d token def = %+v, want %s", i, create.Source, test.want[i])
				}
			}
		})
	}
}

// TestLowerActivatedNamedTokenChoice verifies that an activated ability whose
// effect is an N-way named-token choice ("{T}: Create your choice of a Blood
// token, a Clue token, or a Food token.") lowers to a choose-one modal ability
// body with one CreateToken mode per alternative.
func TestLowerActivatedNamedTokenChoice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Font",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Create your choice of a Blood token, a Clue token, or a Food token.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	content := face.ActivatedAbilities[0].Content
	want := []string{string(types.Blood), string(types.Clue), string(types.Food)}
	if len(content.Modes) != len(want) || content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("modal shape = modes %d min %d max %d, want %d/1/1",
			len(content.Modes), content.MinModes, content.MaxModes, len(want))
	}
	for i, mode := range content.Modes {
		create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
		if !ok {
			t.Fatalf("mode %d primitive = %T, want game.CreateToken", i, mode.Sequence[0].Primitive)
		}
		def, ok := create.Source.TokenDefRef()
		if !ok || def.Name != want[i] {
			t.Fatalf("mode %d token def = %+v, want %s", i, create.Source, want[i])
		}
	}
}

func TestLowerSingleCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 white Soldier creature token.",
		Colors:     []string{"W"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %+v, want one instruction", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", mode.Sequence[0].Primitive)
	}
	if create.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1", create.Amount.Value())
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if def.Name != "Soldier" ||
		!def.Power.Exists || def.Power.Val.Value != 1 ||
		!def.Toughness.Exists || def.Toughness.Val.Value != 1 {
		t.Fatalf("token def = %+v, want 1/1 Soldier", def.CardFace)
	}
	if len(def.Types) != 1 || def.Types[0] != types.Creature {
		t.Fatalf("token types = %v, want [Creature]", def.Types)
	}
	if len(def.Subtypes) != 1 || def.Subtypes[0] != types.Soldier {
		t.Fatalf("token subtypes = %v, want [Soldier]", def.Subtypes)
	}
	if len(def.Colors) != 1 || def.Colors[0] != color.White {
		t.Fatalf("token colors = %v, want [White]", def.Colors)
	}
}

// TestLowerConditionalCreateTokenIgnoresLeakedDuration verifies that a triggered
// create-token whose intervening "if you attacked this turn" condition leaks a
// spurious DurationThisTurn onto the create effect still lowers: the token is
// created and the intervening condition is preserved on the trigger. Creating a
// token is instantaneous, so the leaked turn-scoped duration is provably not part
// of the create clause.
func TestLowerConditionalCreateTokenIgnoresLeakedDuration(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Conditional Token",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Whenever this creature attacks, if you attacked this turn, create a 1/1 white Soldier creature token.",
		Power:      new("2"),
		Toughness:  new("2"),
		Colors:     []string{"W"},
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf == "" || !trigger.InterveningCondition.Exists {
		t.Fatalf("trigger = %+v, want intervening condition preserved", trigger)
	}
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want one create instruction", content)
	}
	if _, ok := content.Modes[0].Sequence[0].Primitive.(game.CreateToken); !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", content.Modes[0].Sequence[0].Primitive)
	}
}

// TestCreateTokenDurationOK verifies the guard that tolerates only a spurious
// turn-scoped duration leaked from an intervening condition. A create-token
// clause is instantaneous, so DurationNone is normal and DurationThisTurn is the
// one provably spurious leak; an "until end of turn"/"until your next turn"
// duration cannot leak from such a clause and stays fail-closed.
func TestCreateTokenDurationOK(t *testing.T) {
	t.Parallel()
	tests := []struct {
		duration compiler.DurationKind
		want     bool
	}{
		{compiler.DurationNone, true},
		{compiler.DurationThisTurn, true},
		{compiler.DurationUntilEndOfTurn, false},
		{compiler.DurationUntilYourNextTurn, false},
	}
	for _, tc := range tests {
		if got := createTokenDurationOK(tc.duration); got != tc.want {
			t.Errorf("createTokenDurationOK(%v) = %v, want %v", tc.duration, got, tc.want)
		}
	}
}

func TestGenerateExecutableCardSourceTokenReferencedControllerRecipient(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Within",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Instant",
		OracleText: "Destroy target permanent. Its controller creates a 3/3 green Beast creature token.",
		Colors:     []string{"G"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Destroy{",
		"Primitive: game.CreateToken{",
		"Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),",
		`Name:      "Beast",`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTokenReferencedControllerRebased(t *testing.T) {
	t.Parallel()
	// The recipient must point at the destroyed permanent (game target 1), not the
	// tapped creature (game target 0): the antecedent index is rebased.
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Rebase",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Instant",
		OracleText: "Tap target creature. Destroy target permanent. Its controller creates a 3/3 green Beast creature token.",
		Colors:     []string{"G"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(1))),") {
		t.Fatalf("recipient not rebased to target 1:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceTokenTargetOpponentRecipient(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Hunted",
		Layout:     "normal",
		ManaCost:   "{2}{W}{W}",
		TypeLine:   "Sorcery",
		OracleText: "Target opponent creates a 4/4 black Horror creature token.",
		Colors:     []string{"W"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Constraint: "Target opponent",`,
		"Allow:      game.TargetAllowPlayer,",
		"Player: game.PlayerOpponent,",
		"Primitive: game.CreateToken{",
		"Recipient: opt.Val(game.TargetPlayerReference(0)),",
		`Name:      "Horror",`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestLowerTargetPlayerNamedTokenRecipient verifies the targeted-player form of a
// predefined named token ("Target opponent creates two Treasure tokens.") lowers
// to a count-2 Treasure CreateToken delivered to the targeted player.
func TestLowerTargetPlayerNamedTokenRecipient(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wanted",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target opponent creates two Treasure tokens.",
	})
	create := createTokenPrimitive(t, face)
	if create.Amount.Value() != 2 {
		t.Fatalf("amount = %d, want 2", create.Amount.Value())
	}
	if !create.Recipient.Exists ||
		create.Recipient.Val.Kind() != game.PlayerReferenceTargetPlayer {
		t.Fatalf("recipient = %+v, want target-player reference", create.Recipient)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok || def.Name != string(types.Treasure) {
		t.Fatalf("token def = %+v, want Treasure", create.Source)
	}
	targets := face.SpellAbility.Val.Modes[0].Targets
	if len(targets) != 1 || targets[0].Predicate.Player != game.PlayerOpponent {
		t.Fatalf("targets = %+v, want one opponent target", targets)
	}
}

func TestGenerateExecutableCardSourceCopyOfTargetCreatureToken(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Copy",
		Layout:     "normal",
		ManaCost:   "{1}{G}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Create a token that's a copy of target creature you control.",
		Colors:     []string{"G", "U"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Constraint: \"target creature you control\",",
		"PermanentTypes: []types.Card{types.Creature},",
		"Controller:     game.ControllerYou,",
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceObject,",
		"Object: game.TargetPermanentReference(0),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCopyOfReferenceToken(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Self Copy",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Creature — Insect",
		OracleText: "{T}: Create a token that's a copy of this creature.",
		Power:      new("1"),
		Toughness:  new("1"),
		Colors:     []string{"G"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceObject,",
		"Object: game.SourcePermanentReference(),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCopyOfTargetExceptNotLegendary(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Imposter",
		Layout:     "normal",
		ManaCost:   "{3}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Create a token that's a copy of target creature you control, except it isn't legendary.",
		Colors:     []string{"U"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Object:          game.TargetPermanentReference(0),",
		"SetNotLegendary: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCopyOfAttachedExceptNotLegendaryGainsKeyword(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Helm of the Host",
		Layout:   "normal",
		ManaCost: "{4}",
		TypeLine: "Artifact — Equipment",
		OracleText: "At the beginning of combat on your turn, create a token that's a copy of equipped creature, " +
			"except the token isn't legendary. That token gains haste.\nEquip {5}",
	}, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Object:          game.SourceAttachedPermanentReference(),",
		"SetNotLegendary: true,",
		"AddKeywords:     []game.Keyword{game.Haste},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceConditionalCopyTokenInstead(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Scute Swarm",
		Layout:   "normal",
		ManaCost: "{2}{G}",
		TypeLine: "Creature — Insect",
		OracleText: "Landfall — Whenever a land you control enters, create a 1/1 green Insect creature token. " +
			"If you control six or more lands, create a token that's a copy of this creature instead.",
		Power:     new("1"),
		Toughness: new("1"),
		Colors:    []string{"G"},
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	// The conditional "instead" form must gate the vanilla 1/1 token on the
	// negation of the six-or-more-lands condition and the copy token on the
	// condition, so exactly one of the two tokens is created.
	for _, wanted := range []string{
		"Source: game.TokenDef(scuteSwarmToken),",
		"Negate: true,",
		"Object: game.SourcePermanentReference(),",
		"MinCount:  6,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCreatureTokenCompiles(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Token",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Sorcery",
		OracleText: "Create a 2/2 green Bear creature token.",
		Colors:     []string{"G"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenDef(testTokenToken)",
		"var testTokenToken = &game.CardDef{",
		`Name:      "Bear",`,
		"Subtypes:  []types.Sub{types.Bear},",
		"Power:     opt.Val(game.PT{Value: 2}),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestLowerArtifactCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Artifact Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 colorless Thopter artifact creature token with flying.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if !reflect.DeepEqual(def.Types, []types.Card{types.Artifact, types.Creature}) {
		t.Fatalf("token types = %v, want [Artifact Creature]", def.Types)
	}
	if len(def.Colors) != 0 {
		t.Fatalf("token colors = %v, want colorless (empty)", def.Colors)
	}
	if len(def.Subtypes) != 1 || def.Subtypes[0] != types.Thopter {
		t.Fatalf("token subtypes = %v, want [Thopter]", def.Subtypes)
	}
	if len(def.StaticAbilities) != 1 || !reflect.DeepEqual(def.StaticAbilities[0], game.FlyingStaticBody) {
		t.Fatalf("token static abilities = %v, want [flying]", def.StaticAbilities)
	}
}

func TestLowerEnchantmentCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Enchantment Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 white Glimmer enchantment creature token.",
		Colors:     []string{"W"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if !reflect.DeepEqual(def.Types, []types.Card{types.Enchantment, types.Creature}) {
		t.Fatalf("token types = %v, want [Enchantment Creature]", def.Types)
	}
	if len(def.Colors) != 1 || def.Colors[0] != color.White {
		t.Fatalf("token colors = %v, want [White]", def.Colors)
	}
}

func TestLowerColorlessCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Colorless Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create four 1/1 colorless Hero creature tokens.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if create.Amount.Value() != 4 {
		t.Fatalf("amount = %d, want 4", create.Amount.Value())
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if !reflect.DeepEqual(def.Types, []types.Card{types.Creature}) {
		t.Fatalf("token types = %v, want [Creature]", def.Types)
	}
	if len(def.Colors) != 0 {
		t.Fatalf("token colors = %v, want colorless (empty)", def.Colors)
	}
}

func TestGenerateExecutableCardSourceArtifactCreatureTokenCompiles(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Artifact Token",
		Layout:     "normal",
		ManaCost:   "{2}",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 colorless Thopter artifact creature token with flying.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Types:     []types.Card{types.Artifact, types.Creature}") {
		t.Fatalf("source missing artifact creature token types:\n%s", source)
	}
}

func TestLowerCreatureTokenWithKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Keyword Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 4/4 red Dragon creature token with flying.",
		Colors:     []string{"R"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if len(def.StaticAbilities) != 1 {
		t.Fatalf("token static abilities = %v, want one (flying)", def.StaticAbilities)
	}
	if !reflect.DeepEqual(def.StaticAbilities[0], game.FlyingStaticBody) {
		t.Fatalf("token static ability = %+v, want game.FlyingStaticBody", def.StaticAbilities[0])
	}
}

func TestLowerMultiKeywordCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Multi Keyword Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 2/1 black Spider creature token with menace and reach.",
		Colors:     []string{"B"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	want := []game.StaticAbility{game.MenaceStaticBody, game.ReachStaticBody}
	if !reflect.DeepEqual(def.StaticAbilities, want) {
		t.Fatalf("token static abilities = %v, want [menace reach]", def.StaticAbilities)
	}
}

func TestLowerMultiKeywordOxfordSeriesToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Oxford Keyword Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 4/4 white Angel creature token with flying, vigilance, and indestructible.",
		Colors:     []string{"W"},
	})
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	want := []game.StaticAbility{
		game.FlyingStaticBody, game.VigilanceStaticBody, game.IndestructibleStaticBody,
	}
	if !reflect.DeepEqual(def.StaticAbilities, want) {
		t.Fatalf("token static abilities = %v, want [flying vigilance indestructible]", def.StaticAbilities)
	}
}

// TestLowerNamedCreatureToken verifies that a creature token with an explicit
// Oracle name ("... creature token[s] named <Name>") lowers to a TokenDef whose
// CardFace.Name is the printed name rather than the joined subtypes, including
// the multi-count, colorless artifact-creature, two-subtype/keyword, and
// two-color forms. The trailing keyword still grants its static ability.
func TestLowerNamedCreatureToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracle     string
		colors     []string
		wantName   string
		wantStatic []game.StaticAbility
	}{
		{
			name:     "single",
			oracle:   "Create a 3/1 red Beast creature token named Carnivore.",
			colors:   []string{"R"},
			wantName: "Carnivore",
		},
		{
			name:     "multi count",
			oracle:   "Create four 3/3 blue Serpent creature tokens named Koma's Coil.",
			colors:   []string{"U"},
			wantName: "Koma's Coil",
		},
		{
			name:     "colorless artifact creature",
			oracle:   "Create a 1/1 colorless Sliver artifact creature token named Metallic Sliver.",
			wantName: "Metallic Sliver",
		},
		{
			name:       "two subtypes with keyword",
			oracle:     "Create a 0/1 blue Plant Wall creature token with defender named Kelp.",
			colors:     []string{"U"},
			wantName:   "Kelp",
			wantStatic: []game.StaticAbility{game.DefenderStaticBody},
		},
		{
			name:     "multi-word name",
			oracle:   "Create a 0/1 red Kobold creature token named Kobolds of Kher Keep.",
			colors:   []string{"R"},
			wantName: "Kobolds of Kher Keep",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Named Token",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
				Colors:     test.colors,
			})
			create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
			if !ok {
				t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
			}
			def, ok := create.Source.TokenDefRef()
			if !ok {
				t.Fatal("token source is not a token definition")
			}
			if def.Name != test.wantName {
				t.Fatalf("token name = %q, want %q", def.Name, test.wantName)
			}
			if !reflect.DeepEqual(def.StaticAbilities, test.wantStatic) {
				t.Fatalf("token static abilities = %v, want %v", def.StaticAbilities, test.wantStatic)
			}
		})
	}
}

// TestNamedTokenWithGrantedAbilityFailsClosed verifies that a named creature
// token carrying a quoted granted ability ("... named X with \"...\"") stays
// unsupported rather than silently dropping the ability into a misnamed token.
func TestNamedTokenWithGrantedAbilityFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Quoted Named Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: `Create a 2/2 blue Demon creature token named Blue Horror with "Whenever you cast an instant or sorcery spell, this token deals 1 damage to any target."`,
		Colors:     []string{"U"},
	})
	if !hasDiagnosticSummary(diagnostics, "unsupported token creation") {
		t.Fatalf("expected unsupported token creation diagnostic, got %#v", diagnostics)
	}
}

func hasDiagnosticSummary(diagnostics []shared.Diagnostic, summary string) bool {
	for i := range diagnostics {
		if diagnostics[i].Summary == summary {
			return true
		}
	}
	return false
}

func TestGenerateExecutableCardSourceKeywordTokenCompiles(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Keyword Token",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Create a 4/4 red Dragon creature token with flying.",
		Colors:     []string{"R"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"StaticAbilities: []game.StaticAbility{",
		"game.FlyingStaticBody,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestLowerNamedArtifactToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		oracle  string
		subtype types.Sub
		mana    bool // mana ability (Treasure) vs activated ability
	}{
		{"Create a Treasure token.", types.Treasure, true},
		{"Create a Food token.", types.Food, false},
		{"Create a Clue token.", types.Clue, false},
		{"Create a Blood token.", types.Blood, false},
		{"Create a Gold token.", types.Gold, true},
		{"Create a Lander token.", types.Lander, false},
		{"Create a Mutagen token.", types.Mutagen, false},
	}
	for _, test := range tests {
		t.Run(string(test.subtype), func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test " + string(test.subtype),
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
			create := createTokenPrimitive(t, face)
			def, ok := create.Source.TokenDefRef()
			if !ok {
				t.Fatal("token source is not a token definition")
			}
			if def.Name != string(test.subtype) {
				t.Fatalf("token name = %q, want %q", def.Name, test.subtype)
			}
			if len(def.Types) != 1 || def.Types[0] != types.Artifact {
				t.Fatalf("token types = %v, want [Artifact]", def.Types)
			}
			if len(def.Subtypes) != 1 || def.Subtypes[0] != test.subtype {
				t.Fatalf("token subtypes = %v, want [%s]", def.Subtypes, test.subtype)
			}
			if def.Power.Exists || def.Toughness.Exists {
				t.Fatalf("named token has unexpected P/T: %+v", def.CardFace)
			}
			if test.mana {
				if len(def.ManaAbilities) != 1 || len(def.ActivatedAbilities) != 0 {
					t.Fatalf("Treasure abilities = mana %d activated %d, want one mana ability", len(def.ManaAbilities), len(def.ActivatedAbilities))
				}
			} else {
				if len(def.ActivatedAbilities) != 1 || len(def.ManaAbilities) != 0 {
					t.Fatalf("%s abilities = activated %d mana %d, want one activated ability", test.subtype, len(def.ActivatedAbilities), len(def.ManaAbilities))
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceTreasureTokenCompiles(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Treasure",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Create two Treasure tokens.",
		Colors:     []string{"R"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Amount: game.Fixed(2),",
		"var testTreasureToken = &game.CardDef{",
		`Name:     "Treasure",`,
		"Types:    []types.Card{types.Artifact},",
		"Subtypes: []types.Sub{types.Treasure},",
		"ManaAbilities: []game.ManaAbility{",
		"Kind:               cost.AdditionalSacrificeSource,",
		"game.ResolutionChoiceMana,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceFoodCluBloodTokensCompile(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		oracle      string
		wantSubtype string
	}{
		{"Create a Food token.", "Subtypes: []types.Sub{types.Food},"},
		{"Create a Clue token.", "Subtypes: []types.Sub{types.Clue},"},
		{"Create a Blood token.", "Subtypes: []types.Sub{types.Blood},"},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Token",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: tc.oracle,
		}, "t")
		if err != nil {
			t.Fatalf("%q: %v", tc.oracle, err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%q: diagnostics = %#v", tc.oracle, diagnostics)
		}
		for _, wanted := range []string{
			"Primitive: game.CreateToken{",
			"Types:    []types.Card{types.Artifact},",
			"ActivatedAbilities: []game.ActivatedAbility{",
			"Kind:               cost.AdditionalSacrificeSource,",
			tc.wantSubtype,
		} {
			if !strings.Contains(source, wanted) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, wanted, source)
			}
		}
	}
}

func TestGenerateExecutableCardSourceGoldLanderMutagenTokensCompile(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		oracle string
		wanted []string
	}{
		{"Create a Gold token.", []string{
			"Subtypes: []types.Sub{types.Gold},",
			"ManaAbilities: []game.ManaAbility{",
			"game.ResolutionChoiceMana,",
			"Kind:   cost.AdditionalSacrificeSource,",
		}},
		{"Create a Lander token.", []string{
			"Subtypes: []types.Sub{types.Lander},",
			"ActivatedAbilities: []game.ActivatedAbility{",
			"Primitive: game.Search{",
			"EntersTapped: true,",
		}},
		{"Create a Mutagen token.", []string{
			"Subtypes: []types.Sub{types.Mutagen},",
			"ActivatedAbilities: []game.ActivatedAbility{",
			"Timing:         game.SorceryOnly,",
			"Primitive: game.AddCounter{",
			"CounterKind: counter.PlusOnePlusOne,",
		}},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Token",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: tc.oracle,
		}, "t")
		if err != nil {
			t.Fatalf("%q: %v", tc.oracle, err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%q: diagnostics = %#v", tc.oracle, diagnostics)
		}
		for _, wanted := range append([]string{
			"Primitive: game.CreateToken{",
			"Types:    []types.Card{types.Artifact},",
		}, tc.wanted...) {
			if !strings.Contains(source, wanted) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, wanted, source)
			}
		}
	}
}

func TestLowerTappedCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapped",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a tapped 2/2 black Zombie creature token.",
		Colors:     []string{"B"},
	})
	create := createTokenPrimitive(t, face)
	if !create.EntryTapped {
		t.Fatal("EntryTapped = false, want true")
	}
	if create.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1", create.Amount.Value())
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if def.Name != "Zombie" ||
		!def.Power.Exists || def.Power.Val.Value != 2 ||
		!def.Toughness.Exists || def.Toughness.Val.Value != 2 {
		t.Fatalf("token def = %+v, want 2/2 Zombie", def.CardFace)
	}
}

func TestLowerTappedMultipleTokens(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapped Many",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create two tapped 1/1 white Dog creature tokens.",
		Colors:     []string{"W"},
	})
	create := createTokenPrimitive(t, face)
	if !create.EntryTapped {
		t.Fatal("EntryTapped = false, want true")
	}
	if create.Amount.Value() != 2 {
		t.Fatalf("amount = %d, want 2", create.Amount.Value())
	}
}

func TestLowerTappedNamedToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapped Treasure",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a tapped Treasure token.",
	})
	create := createTokenPrimitive(t, face)
	if !create.EntryTapped {
		t.Fatal("EntryTapped = false, want true")
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if def.Name != string(types.Treasure) {
		t.Fatalf("token name = %q, want Treasure", def.Name)
	}
}

func TestLowerTappedKeywordToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapped Keyword",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create three tapped 1/1 white Spirit creature tokens with flying.",
		Colors:     []string{"W"},
	})
	create := createTokenPrimitive(t, face)
	if !create.EntryTapped {
		t.Fatal("EntryTapped = false, want true")
	}
	if create.Amount.Value() != 3 {
		t.Fatalf("amount = %d, want 3", create.Amount.Value())
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if len(def.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1 (flying)", len(def.StaticAbilities))
	}
}

func TestGenerateExecutableCardSourceTappedTokenRendersEntryTapped(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Tapped Render",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Create a tapped 2/2 black Zombie creature token.",
		Colors:     []string{"B"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"EntryTapped: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceAttackingTokenRendersEntryAttacking(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Attacking Render",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 white Soldier creature token that's tapped and attacking.",
		Colors:     []string{"W"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"EntryAttacking: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if !strings.Contains(source, "EntryTapped:") {
		t.Fatalf("tapped-and-attacking token should enter tapped:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceAttackingOnlyTokenRendersEntryAttacking(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Attacking Only",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 white Cat Soldier creature token with vigilance that's attacking.",
		Colors:     []string{"W"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "EntryAttacking: true,") {
		t.Fatalf("source missing EntryAttacking:\n%s", source)
	}
	if strings.Contains(source, "EntryTapped: true,") {
		t.Fatalf("attacking-only token should not enter tapped:\n%s", source)
	}
}

func TestCreateTokenFailsClosedForUnsupportedShapes(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Create a Powerstone token.", // named token without a representable ability
		"Create a 1/1 white Soldier creature token with flying and protection from red.", // parameterized keyword rider not representable
		"Create a 2/2 green Boar creature token that's tapped and blocking.",             // blocking entry not representable
		"Each opponent creates a 1/1 white Human creature token.",                        // player-group recipient not a single player reference
	} {
		_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Token",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: oracle,
		}, "t")
		if err != nil {
			t.Fatalf("%q: %v", oracle, err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("%q: expected fail-closed, got supported", oracle)
		}
	}
}

func TestLowerTrailingForEachToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Trailing ForEach",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 green Elf Warrior creature token for each Elf you control.",
		Colors:     []string{"G"},
	})
	create := createTokenPrimitive(t, face)
	want := game.DynamicAmount{
		Kind:       game.DynamicAmountCountSelector,
		Multiplier: 1,
		Group: game.BattlefieldGroup(game.Selection{
			SubtypesAny: []types.Sub{types.Sub("Elf")},
			Controller:  game.ControllerYou,
		}),
	}
	if got := create.Amount.DynamicAmount().Val; !reflect.DeepEqual(got, want) {
		t.Fatalf("dynamic amount = %+v, want %+v", got, want)
	}
	def, _ := create.Source.TokenDefRef()
	if def.Name != "Elf Warrior" {
		t.Fatalf("token name = %q, want Elf Warrior", def.Name)
	}
}

func TestLowerNumberOfEqualToToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Equal Count",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a number of 1/1 white Soldier creature tokens equal to the number of opponents you have.",
		Colors:     []string{"W"},
	})
	create := createTokenPrimitive(t, face)
	if !create.Amount.IsDynamic() ||
		create.Amount.DynamicAmount().Val.Kind != game.DynamicAmountOpponentCount {
		t.Fatalf("amount = %+v, want a dynamic opponent count", create.Amount)
	}
}

func TestLowerWhereXToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Where X",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create X 1/1 white Soldier creature tokens, where X is the number of creatures you control.",
		Colors:     []string{"W"},
	})
	create := createTokenPrimitive(t, face)
	if !create.Amount.IsDynamic() ||
		create.Amount.DynamicAmount().Val.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("amount = %+v, want a dynamic creature count", create.Amount)
	}
}

func TestLowerVariableXToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Variable X",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create X 1/1 red Goblin creature tokens.",
		Colors:     []string{"R"},
	})
	create := createTokenPrimitive(t, face)
	if !create.Amount.IsDynamic() ||
		create.Amount.DynamicAmount().Val.Kind != game.DynamicAmountX {
		t.Fatalf("amount = %+v, want the variable X amount", create.Amount)
	}
}

func TestCreateTokenFailsClosedForUnrepresentableDynamicCount(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		// "half the number of Zombies" is not a representable dynamic count.
		"Create X 2/2 black Zombie creature tokens, where X is half the number of Zombies you control, rounded down.",
	} {
		_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Token",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: oracle,
		}, "t")
		if err != nil {
			t.Fatalf("%q: %v", oracle, err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("%q: expected fail-closed, got supported", oracle)
		}
	}
}
