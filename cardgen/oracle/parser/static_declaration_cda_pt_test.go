package parser

import "testing"

func TestParseCharacteristicDefiningPowerToughnessDeclarationMeaning(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
		card   string
		want   StaticDeclarationDynamicValueKind
	}{
		{
			name:   "this creature cards in hand",
			source: "This creature's power and toughness are each equal to the number of cards in your hand.",
			want:   StaticDeclarationDynamicValueControllerHandSize,
		},
		{
			name:   "this creature cards in graveyard",
			source: "This creature's power and toughness are each equal to the number of cards in your graveyard.",
			want:   StaticDeclarationDynamicValueControllerGraveyardSize,
		},
		{
			name:   "this creature lands you control",
			source: "This creature's power and toughness are each equal to the number of lands you control.",
			want:   StaticDeclarationDynamicValueControllerLandCount,
		},
		{
			name:   "this creature creatures you control",
			source: "This creature's power and toughness are each equal to the number of creatures you control.",
			want:   StaticDeclarationDynamicValueControllerCreatureCount,
		},
		{
			name:   "this creature artifacts you control",
			source: "This creature's power and toughness are each equal to the number of artifacts you control.",
			want:   StaticDeclarationDynamicValueControllerArtifactCount,
		},
		{
			name:   "this creature creatures on the battlefield",
			source: "This creature's power and toughness are each equal to the number of creatures on the battlefield.",
			want:   StaticDeclarationDynamicValueAllBattlefieldCreatureCount,
		},
		{
			name:   "self name cards in hand",
			source: "Psychosis Crawler's power and toughness are each equal to the number of cards in your hand.",
			card:   "Psychosis Crawler",
			want:   StaticDeclarationDynamicValueControllerHandSize,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, Context{CardName: tc.card})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want exactly one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationCharacteristicDefiningPowerToughness {
				t.Fatalf("kind = %q, want characteristic-defining power/toughness", declaration.Kind)
			}
			if declaration.Subject.Kind != StaticDeclarationSubjectSourceCreature {
				t.Fatalf("subject = %q, want source creature", declaration.Subject.Kind)
			}
			if declaration.DynamicValue != tc.want {
				t.Fatalf("dynamic value = %q, want %q", declaration.DynamicValue, tc.want)
			}
		})
	}
}

func TestParseCharacteristicDefiningPowerToughnessForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		source        string
		card          string
		wantValue     StaticDeclarationDynamicValueKind
		wantSetsPower bool
		wantSetsTough bool
		wantOffset    int
	}{
		{
			name:          "power only creatures you control",
			source:        "This creature's power is equal to the number of creatures you control.",
			wantValue:     StaticDeclarationDynamicValueControllerCreatureCount,
			wantSetsPower: true,
		},
		{
			name:          "both equal sets both",
			source:        "This creature's power and toughness are each equal to the number of cards in your hand.",
			wantValue:     StaticDeclarationDynamicValueControllerHandSize,
			wantSetsPower: true,
			wantSetsTough: true,
		},
		{
			name:          "toughness offset card types among all graveyards",
			source:        "Tarmogoyf's power is equal to the number of card types among cards in all graveyards and its toughness is equal to that number plus 1.",
			card:          "Tarmogoyf",
			wantValue:     StaticDeclarationDynamicValueCardTypesAmongAllGraveyards,
			wantSetsPower: true,
			wantSetsTough: true,
			wantOffset:    1,
		},
		{
			name:          "both equal creature cards in all graveyards",
			source:        "This creature's power and toughness are each equal to the number of creature cards in all graveyards.",
			wantValue:     StaticDeclarationDynamicValueCreatureCardsInAllGraveyards,
			wantSetsPower: true,
			wantSetsTough: true,
		},
		{
			name:          "both equal cards in all graveyards",
			source:        "This creature's power and toughness are each equal to the number of cards in all graveyards.",
			wantValue:     StaticDeclarationDynamicValueAllGraveyardsSize,
			wantSetsPower: true,
			wantSetsTough: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, Context{CardName: tc.card})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want exactly one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationCharacteristicDefiningPowerToughness {
				t.Fatalf("kind = %q, want characteristic-defining power/toughness", declaration.Kind)
			}
			if declaration.DynamicValue != tc.wantValue {
				t.Fatalf("dynamic value = %q, want %q", declaration.DynamicValue, tc.wantValue)
			}
			if declaration.DynamicSetsPower != tc.wantSetsPower {
				t.Fatalf("sets power = %v, want %v", declaration.DynamicSetsPower, tc.wantSetsPower)
			}
			if declaration.DynamicSetsToughness != tc.wantSetsTough {
				t.Fatalf("sets toughness = %v, want %v", declaration.DynamicSetsToughness, tc.wantSetsTough)
			}
			if declaration.DynamicToughnessOffset != tc.wantOffset {
				t.Fatalf("toughness offset = %d, want %d", declaration.DynamicToughnessOffset, tc.wantOffset)
			}
		})
	}
}

func TestParseCharacteristicDefiningPowerToughnessNewCounts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		source      string
		card        string
		wantValue   StaticDeclarationDynamicValueKind
		wantSubtype TriggerSubtype
		wantColor   Color
	}{
		{
			name:      "creature cards in your graveyard",
			source:    "This creature's power and toughness are each equal to the number of creature cards in your graveyard.",
			wantValue: StaticDeclarationDynamicValueControllerCreatureCardsInGraveyard,
		},
		{
			name:      "instant and sorcery cards in your graveyard",
			source:    "This creature's power and toughness are each equal to the number of instant and sorcery cards in your graveyard.",
			wantValue: StaticDeclarationDynamicValueControllerInstantOrSorceryCardsInGraveyard,
		},
		{
			name:      "land cards in your graveyard",
			source:    "This creature's power is equal to the number of land cards in your graveyard.",
			wantValue: StaticDeclarationDynamicValueControllerLandCardsInGraveyard,
		},
		{
			name:      "permanent cards in your graveyard",
			source:    "This creature's power and toughness are each equal to the number of permanent cards in your graveyard.",
			wantValue: StaticDeclarationDynamicValueControllerPermanentCardsInGraveyard,
		},
		{
			name:      "card types among cards in your graveyard",
			source:    "This creature's power and toughness are each equal to the number of card types among cards in your graveyard.",
			wantValue: StaticDeclarationDynamicValueControllerCardTypesInGraveyard,
		},
		{
			name:      "basic land types among lands you control",
			source:    "This creature's power and toughness are each equal to the number of basic land types among lands you control.",
			wantValue: StaticDeclarationDynamicValueControllerBasicLandTypeCount,
		},
		{
			name:      "life total",
			source:    "This creature's power and toughness are each equal to your life total.",
			wantValue: StaticDeclarationDynamicValueControllerLifeTotal,
		},
		{
			name:      "all players hands",
			source:    "This creature's power and toughness are each equal to the total number of cards in all players' hands.",
			wantValue: StaticDeclarationDynamicValueAllPlayersHandSize,
		},
		{
			name:        "swamps you control",
			source:      "Korlash's power and toughness are each equal to the number of Swamps you control.",
			card:        "Korlash",
			wantValue:   StaticDeclarationDynamicValueControllerSubtypeCount,
			wantSubtype: "Swamp",
		},
		{
			name:        "forests you control",
			source:      "This creature's power and toughness are each equal to the number of Forests you control.",
			wantValue:   StaticDeclarationDynamicValueControllerSubtypeCount,
			wantSubtype: "Forest",
		},
		{
			name:        "goblins you control",
			source:      "This creature's power and toughness are each equal to the number of Goblins you control.",
			wantValue:   StaticDeclarationDynamicValueControllerSubtypeCount,
			wantSubtype: "Goblin",
		},
		{
			name:        "spirits you control",
			source:      "This creature's power is equal to the number of Spirits you control.",
			wantValue:   StaticDeclarationDynamicValueControllerSubtypeCount,
			wantSubtype: "Spirit",
		},
		{
			name:      "red permanents you control",
			source:    "This creature's power and toughness are each equal to the number of red permanents you control.",
			wantValue: StaticDeclarationDynamicValueControllerColorPermanentCount,
			wantColor: ColorRed,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, tc.source, Context{CardName: tc.card})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want exactly one", declarations)
			}
			declaration := declarations[0]
			if declaration.Kind != StaticDeclarationCharacteristicDefiningPowerToughness {
				t.Fatalf("kind = %q, want characteristic-defining power/toughness", declaration.Kind)
			}
			if declaration.DynamicValue != tc.wantValue {
				t.Fatalf("dynamic value = %q, want %q", declaration.DynamicValue, tc.wantValue)
			}
			if string(declaration.DynamicValueSubtype) != string(tc.wantSubtype) {
				t.Fatalf("dynamic value subtype = %q, want %q", declaration.DynamicValueSubtype, tc.wantSubtype)
			}
			if declaration.DynamicValueColor != tc.wantColor {
				t.Fatalf("dynamic value color = %q, want %q", declaration.DynamicValueColor, tc.wantColor)
			}
		})
	}
}

func TestParseCharacteristicDefiningPowerToughnessFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "non-source subject",
			source: "Enchanted creature's power and toughness are each equal to the number of cards in your hand.",
		},
		{
			name:   "unsupported count",
			source: "This creature's power and toughness are each equal to the number of nonbasic lands you control.",
		},
		{
			name:   "compound count",
			source: "This creature's power and toughness are each equal to the number of cards in your hand plus the number of cards in your graveyard.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.source, Context{})
			for _, ability := range document.Abilities {
				for _, declaration := range ability.StaticDeclarations {
					if declaration.Kind == StaticDeclarationCharacteristicDefiningPowerToughness {
						t.Fatalf("source %q produced an unexpected characteristic-defining declaration", tc.source)
					}
				}
			}
		})
	}
}
