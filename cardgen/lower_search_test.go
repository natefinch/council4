package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// loweredSearch lowers a single-spell-ability card and returns the lone Search
// primitive its spell ability produces, failing the test on any diagnostic or
// unexpected shape.
func loweredSearch(t *testing.T, typeLine, oracleText string) game.Search {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Search Test",
		Layout:     "normal",
		TypeLine:   typeLine,
		OracleText: oracleText,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("lowerExecutableFaces(%q) diagnostics = %#v", oracleText, diagnostics)
	}
	if len(faces) != 1 || !faces[0].SpellAbility.Exists {
		t.Fatalf("lowerExecutableFaces(%q) faces = %#v", oracleText, faces)
	}
	modes := faces[0].SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("lowerExecutableFaces(%q) modes = %#v", oracleText, modes)
	}
	search, ok := modes[0].Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("lowerExecutableFaces(%q) primitive = %#v, want game.Search", oracleText, modes[0].Sequence[0].Primitive)
	}
	return search
}

func TestLowerSearchSpellSpecs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		amount     int
		spec       game.SearchSpec
	}{
		{
			name:       "basic land ramp to battlefield tapped, up to two",
			typeLine:   "Sorcery",
			oracleText: "Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
			amount:     2,
			spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				CardType:     opt.Val(types.Land),
				Supertype:    opt.Val(types.Basic),
				EntersTapped: true,
			},
		},
		{
			name:       "dual-land subtype union to battlefield",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a Forest or Island card, put it onto the battlefield, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Battlefield,
				SubtypesAny: []types.Sub{types.Forest, types.Island},
			},
		},
		{
			name:       "creature tutor to hand with reveal",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a creature card, reveal it, put it into your hand, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Creature),
				Reveal:      true,
			},
		},
		{
			name:       "basic triome to battlefield tapped",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a basic Forest, Plains, or Island card, put it onto the battlefield tapped, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				Supertype:    opt.Val(types.Basic),
				SubtypesAny:  []types.Sub{types.Forest, types.Plains, types.Island},
				EntersTapped: true,
			},
		},
		{
			name:       "non-basic subtype tutor to hand",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a Sliver card, reveal it, put it into your hand, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				SubtypesAny: []types.Sub{types.Sliver},
				Reveal:      true,
			},
		},
		{
			name:       "non-basic subtype union to battlefield",
			typeLine:   "Sorcery",
			oracleText: "Search your library for an Aura or Equipment card, put it onto the battlefield, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Battlefield,
				SubtypesAny: []types.Sub{types.Aura, types.Equipment},
			},
		},
		{
			name:       "subtype paired with creature type to battlefield",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a Myr creature card, put it onto the battlefield, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Battlefield,
				CardType:    opt.Val(types.Creature),
				SubtypesAny: []types.Sub{types.Myr},
			},
		},
		{
			name:       "planeswalker tutor to hand, up to two",
			typeLine:   "Sorcery",
			oracleText: "Search your library for up to two planeswalker cards, reveal them, put them into your hand, then shuffle.",
			amount:     2,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Planeswalker),
				Reveal:      true,
			},
		},
		{
			name:       "subtype permanent tutor with mana-value bound to battlefield",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a Rebel permanent card with mana value 5 or less, put it onto the battlefield, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				Permanent:    true,
				SubtypesAny:  []types.Sub{types.Rebel},
				MaxManaValue: opt.Val(5),
			},
		},
		{
			name:       "plain permanent tutor to battlefield",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a permanent card, put it onto the battlefield, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Battlefield,
				Permanent:   true,
			},
		},
		{
			name:       "legendary subtype permanent tutor to battlefield",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a legendary Spirit permanent card, put it onto the battlefield, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Battlefield,
				Permanent:   true,
				Supertype:   opt.Val(types.Legendary),
				SubtypesAny: []types.Sub{types.Spirit},
			},
		},
		{
			name:       "typed card tutor with mana-value bound to hand",
			typeLine:   "Sorcery",
			oracleText: "Search your library for an artifact card with mana value 1 or less, reveal it, put it into your hand, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Hand,
				CardType:     opt.Val(types.Artifact),
				MaxManaValue: opt.Val(1),
				Reveal:       true,
			},
		},
		{
			name:       "split-destination land tutor: one to battlefield tapped, the other to hand",
			typeLine:   "Sorcery",
			oracleText: "Search your library for up to two basic land cards, reveal those cards, put one onto the battlefield tapped and the other into your hand, then shuffle.",
			amount:     2,
			spec: game.SearchSpec{
				SourceZone:       zone.Library,
				Destination:      zone.Battlefield,
				CardType:         opt.Val(types.Land),
				Supertype:        opt.Val(types.Basic),
				Reveal:           true,
				EntersTapped:     true,
				SplitDestination: opt.Val(game.SearchDestination{Zone: zone.Hand}),
			},
		},
		{
			name:       "split-destination land tutor without reveal",
			typeLine:   "Sorcery",
			oracleText: "Search your library for up to two basic land cards, put one into your hand and the other onto the battlefield tapped, then shuffle.",
			amount:     2,
			spec: game.SearchSpec{
				SourceZone:       zone.Library,
				Destination:      zone.Hand,
				CardType:         opt.Val(types.Land),
				Supertype:        opt.Val(types.Basic),
				SplitDestination: opt.Val(game.SearchDestination{Zone: zone.Battlefield, EntersTapped: true}),
			},
		},
		{
			name:       "legendary creature tutor to hand",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a legendary creature card, reveal it, put it into your hand, then shuffle.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Creature),
				Supertype:   opt.Val(types.Legendary),
				Reveal:      true,
			},
		},
		{
			name:       "shared-land-type correlated tutor to battlefield tapped (Myriad Landscape)",
			typeLine:   "Sorcery",
			oracleText: "Search your library for up to two basic land cards that share a land type, put them onto the battlefield tapped, then shuffle.",
			amount:     2,
			spec: game.SearchSpec{
				SourceZone:    zone.Library,
				Destination:   zone.Battlefield,
				CardType:      opt.Val(types.Land),
				Supertype:     opt.Val(types.Basic),
				EntersTapped:  true,
				SharedSubtype: true,
			},
		},
		{
			name:       "artifact or enchantment tutor to top with reveal",
			typeLine:   "Instant",
			oracleText: "Search your library for an artifact or enchantment card, reveal it, then shuffle and put that card on top.",
			amount:     1,
			spec: game.SearchSpec{
				SourceZone:          zone.Library,
				Destination:         zone.Library,
				DestinationPosition: game.SearchPositionTop,
				CardTypesAny:        []types.Card{types.Artifact, types.Enchantment},
				Reveal:              true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			search := loweredSearch(t, test.typeLine, test.oracleText)
			if got := search.Amount.Value(); got != test.amount {
				t.Errorf("amount = %d, want %d", got, test.amount)
			}
			if got := search.Spec; !searchSpecEqual(got, test.spec) {
				t.Errorf("spec = %+v, want %+v", got, test.spec)
			}
		})
	}
}

// TestLowerSpellTypeTutorSpecs covers library-top tutors filtered by the spell
// card types instant and sorcery: a type union ("instant or sorcery card") lowers
// to CardTypesAny, while a single spell type ("a sorcery card") lowers to the
// singular CardType filter. The "the card" demonstrative case confirms the put
// destination accepts that wording alongside "it"/"that card".
func TestLowerSpellTypeTutorSpecs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		spec       game.SearchSpec
	}{
		{
			name:       "instant or sorcery tutor to top with reveal (Mystical Tutor)",
			typeLine:   "Instant",
			oracleText: "Search your library for an instant or sorcery card, reveal it, then shuffle and put that card on top.",
			spec: game.SearchSpec{
				SourceZone:          zone.Library,
				Destination:         zone.Library,
				DestinationPosition: game.SearchPositionTop,
				CardTypesAny:        []types.Card{types.Instant, types.Sorcery},
				Reveal:              true,
			},
		},
		{
			name:       "single sorcery tutor to top with reveal (Personal Tutor)",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a sorcery card, reveal it, then shuffle and put that card on top.",
			spec: game.SearchSpec{
				SourceZone:          zone.Library,
				Destination:         zone.Library,
				DestinationPosition: game.SearchPositionTop,
				CardType:            opt.Val(types.Sorcery),
				Reveal:              true,
			},
		},
		{
			name:       "single creature tutor to top with \"the card\" wording (Worldly Tutor)",
			typeLine:   "Instant",
			oracleText: "Search your library for a creature card, reveal it, then shuffle and put the card on top.",
			spec: game.SearchSpec{
				SourceZone:          zone.Library,
				Destination:         zone.Library,
				DestinationPosition: game.SearchPositionTop,
				CardType:            opt.Val(types.Creature),
				Reveal:              true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			search := loweredSearch(t, test.typeLine, test.oracleText)
			if got := search.Amount.Value(); got != 1 {
				t.Errorf("amount = %d, want 1", got)
			}
			if got := search.Spec; !searchSpecEqual(got, test.spec) {
				t.Errorf("spec = %+v, want %+v", got, test.spec)
			}
		})
	}
}

// TestLowerSplitSearchFailsClosed confirms the split-destination lowerer rejects
// near-miss shapes rather than producing a silently-wrong distribution. Each
// case keeps the "put one ... and the other ..." split but breaks one structural
// requirement, so the whole spell must stay unsupported.
func TestLowerSplitSearchFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		oracle string
	}{
		{
			// More than two cards cannot fill exactly two single-card slots.
			name:   "up to three split",
			oracle: "Search your library for up to three basic land cards, reveal those cards, put one onto the battlefield tapped and the other into your hand, then shuffle.",
		},
		{
			// Myriad Landscape's shared-land-type constraint is not modeled.
			name:   "share a land type",
			oracle: "Search your library for up to two basic land cards that share a land type, reveal those cards, put one onto the battlefield tapped and the other into your hand, then shuffle.",
		},
		{
			// A graveyard slot is not a modeled split destination.
			name:   "graveyard slot",
			oracle: "Search your library for up to two basic land cards, put one onto the battlefield tapped and the other into your graveyard, then shuffle.",
		},
		{
			// An extra trailing clause after the split breaks the envelope.
			name:   "extra trailing clause",
			oracle: "Search your library for up to two basic land cards, put one onto the battlefield tapped and the other into your hand, then shuffle, then draw a card.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Split Test",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: tc.oracle,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected an unsupported diagnostic, got faces = %#v", faces)
			}
		})
	}
}

func searchSpecEqual(a, b game.SearchSpec) bool {
	return a.SourceZone == b.SourceZone &&
		a.Destination == b.Destination &&
		a.DestinationPosition == b.DestinationPosition &&
		a.FailToFindPolicy == b.FailToFindPolicy &&
		a.CardType == b.CardType &&
		a.Supertype == b.Supertype &&
		a.Permanent == b.Permanent &&
		a.MaxManaValue == b.MaxManaValue &&
		a.Reveal == b.Reveal &&
		a.EntersTapped == b.EntersTapped &&
		a.SplitDestination == b.SplitDestination &&
		a.SharedSubtype == b.SharedSubtype &&
		slices.Equal(a.CardTypesAny, b.CardTypesAny) &&
		slices.Equal(a.SubtypesAny, b.SubtypesAny)
}

func TestLowerVampiricTutorSearchThenLoseLife(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Vampiric Tutor",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Search your library for a card, then shuffle and put that card on top. You lose 2 life.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	mode := faces[0].SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want search then life loss", mode.Sequence)
	}
	search, ok := mode.Sequence[0].Primitive.(game.Search)
	if !ok || search.Spec.Destination != zone.Library ||
		search.Spec.DestinationPosition != game.SearchPositionTop ||
		search.Spec.FailToFindPolicy != game.SearchMustFindIfAvailable ||
		search.Spec.Reveal {
		t.Fatalf("first primitive = %#v, want required hidden library-top search", mode.Sequence[0].Primitive)
	}
	lose, ok := mode.Sequence[1].Primitive.(game.LoseLife)
	if !ok || lose.Amount.Value() != 2 || lose.Player != game.ControllerReference() {
		t.Fatalf("second primitive = %#v, want controller lose 2 life", mode.Sequence[1].Primitive)
	}
}

func TestLowerSearchFailToFindPolicies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		want       game.SearchFailToFindPolicy
	}{
		{
			name:       "unrestricted exact card search",
			oracleText: "Search your library for a card, then shuffle and put that card on top.",
			want:       game.SearchMustFindIfAvailable,
		},
		{
			name:       "qualified exact card search",
			oracleText: "Search your library for an artifact or enchantment card, reveal it, then shuffle and put that card on top.",
			want:       game.SearchFailToFindDefault,
		},
		{
			name:       "up to search",
			oracleText: "Search your library for up to two basic land cards, put them into your hand, then shuffle.",
			want:       game.SearchFailToFindDefault,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			search := loweredSearch(t, "Sorcery", test.oracleText)
			if search.Spec.FailToFindPolicy != test.want {
				t.Fatalf("fail-to-find policy = %v, want %v", search.Spec.FailToFindPolicy, test.want)
			}
		})
	}
}

func TestExactSearchEffectSequenceRejectsTruncatedRevealShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		effects []compiler.CompiledEffect
	}{
		{
			name: "missing shuffle",
			effects: []compiler.CompiledEffect{
				{Kind: compiler.EffectSearch},
				{Kind: compiler.EffectReveal},
				{Kind: compiler.EffectPut},
			},
		},
		{
			name: "missing put",
			effects: []compiler.CompiledEffect{
				{Kind: compiler.EffectSearch},
				{Kind: compiler.EffectReveal},
				{Kind: compiler.EffectShuffle},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if shape, ok := exactSearchEffectSequence(test.effects); ok {
				t.Fatalf("exactSearchEffectSequence() = %#v, true; want fail closed", shape)
			}
		})
	}
}

func TestLowerLibraryTopSearchFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Search your library for up to two cards, then shuffle and put those cards on top.",
		"Search your library for a card, then shuffle and put that card on the bottom.",
		"Search your library for a card, then shuffle and put that card on top at random.",
		"Search your library for a card, then shuffle and put that card in the top three cards of your library.",
		"Search your library for a creature card, reveal it, put it into your hand.",
		"Search your library for a card, then shuffle and put that card on top. Draw a card.",
		"Search your library for a card, then shuffle and put that card on top. You gain 2 life.",
	} {
		if faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Near Miss",
			Layout:     "normal",
			TypeLine:   "Instant",
			OracleText: oracleText,
		}); len(diagnostics) == 0 {
			t.Errorf("%q unexpectedly lowered: %#v", oracleText, faces)
		}
	}
}

func TestGenerateTopTutorCardSourceEndToEnd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		card  ScryfallCard
		wants []string
	}{
		{
			card: ScryfallCard{
				Name:       "Vampiric Tutor",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: "Search your library for a card, then shuffle and put that card on top. You lose 2 life.",
			},
			wants: []string{
				"DestinationPosition: game.SearchPositionTop",
				"FailToFindPolicy:",
				"game.SearchMustFindIfAvailable",
				"Primitive: game.LoseLife{",
				"Amount: game.Fixed(2)",
			},
		},
		{
			card: ScryfallCard{
				Name:       "Enlightened Tutor",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: "Search your library for an artifact or enchantment card, reveal it, then shuffle and put that card on top.",
			},
			wants: []string{
				"DestinationPosition: game.SearchPositionTop",
				"CardTypesAny:",
				"[]types.Card{types.Artifact, types.Enchantment}",
				"Reveal:",
			},
		},
	}
	for _, test := range tests {
		source, diagnostics, err := GenerateExecutableCardSource(&test.card, "t")
		if err != nil {
			t.Fatalf("%s: %v", test.card.Name, err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%s diagnostics = %#v", test.card.Name, diagnostics)
		}
		for _, want := range test.wants {
			if !strings.Contains(source, want) {
				t.Fatalf("%s source missing %q:\n%s", test.card.Name, want, source)
			}
		}
	}
}

// TestLowerSharedSubtypeSearchFailsClosed confirms the correlated-search lowerer
// rejects near-miss shapes rather than dropping the shared-land-type constraint.
// Each case keeps the "that share a land type" wording but breaks one structural
// requirement, so the whole spell must stay unsupported.
func TestLowerSharedSubtypeSearchFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		oracle string
	}{
		{
			// A correlation needs two cards; a singular search has only one.
			name:   "singular search",
			oracle: "Search your library for a basic land card that share a land type, put it onto the battlefield tapped, then shuffle.",
		},
		{
			// More than two cards is outside the modeled two-card shape.
			name:   "up to three",
			oracle: "Search your library for up to three basic land cards that share a land type, put them onto the battlefield tapped, then shuffle.",
		},
		{
			// "share a land type" is not meaningful for a non-land filter.
			name:   "non-land filter",
			oracle: "Search your library for up to two creature cards that share a land type, put them onto the battlefield, then shuffle.",
		},
		{
			// A different correlation property is not modeled.
			name:   "share a color",
			oracle: "Search your library for up to two basic land cards that share a color, put them onto the battlefield tapped, then shuffle.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Shared Subtype Test",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: tc.oracle,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected an unsupported diagnostic, got faces = %#v", faces)
			}
		})
	}
}
