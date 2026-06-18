package cardgen

import (
	"reflect"
	"strings"
	"testing"

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

func TestCreateTokenFailsClosedForUnsupportedShapes(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Create a Powerstone token.", // named token without a representable ability
		"Create a 1/1 white Soldier creature token with flying and vigilance.", // multiple keywords
		"Create a 2/2 green Boar creature token that's tapped and attacking.",  // attacking entry not representable
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
