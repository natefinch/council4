package parser

import (
	"slices"
	"testing"
)

// searchExact parses a single library-search sentence and reports whether its
// resolving effect round-tripped to an exact, lowerable production.
func searchExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) == 0 || effects[0].Kind != EffectSearch {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactLibrarySearchAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		// Plain and single card-type singular searches.
		"Search your library for a card, put that card into your hand, then shuffle.",
		"Search your library for a creature card, reveal it, put it into your hand, then shuffle.",
		"Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		"Search your library for a land card, put it onto the battlefield tapped, then shuffle.",
		// Basic land subtype unions (fetch and dual lands).
		"Search your library for a Forest or Island card, put it onto the battlefield, then shuffle.",
		"Search your library for a Mountain or Forest card, put it onto the battlefield, then shuffle.",
		"Search your library for a basic Forest card, put it onto the battlefield tapped, then shuffle.",
		"Search your library for a basic Forest, Plains, or Island card, put it onto the battlefield tapped, then shuffle.",
		// "up to N" plural searches with plural destination wording.
		"Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
		"Search your library for up to three basic land cards, put them onto the battlefield, then shuffle.",
		"Search your library for up to two enchantment cards, reveal them, put them into your hand, then shuffle.",
		"Search your library for up to two basic land cards, put those cards onto the battlefield tapped, then shuffle.",
		// Non-basic subtype searches (the subtype implies the card type).
		"Search your library for a Sliver card, reveal it, put it into your hand, then shuffle.",
		"Search your library for an Equipment card, put it onto the battlefield, then shuffle.",
		"Search your library for an Aura or Equipment card, put it into your hand, then shuffle.",
		// A subtype paired with a card type.
		"Search your library for a Myr creature card, put it onto the battlefield, then shuffle.",
		"Search your library for a Dragon creature card, reveal it, put it into your hand, then shuffle.",
		// A color filter on a card type (Green Sun's Zenith family), single color
		// and a color union; the runtime matches the card's color identity.
		"Search your library for a green creature card, put it onto the battlefield, then shuffle.",
		"Search your library for a white creature card, reveal it, put it into your hand, then shuffle.",
		"Search your library for a red or green creature card, put it onto the battlefield, then shuffle.",
		// Planeswalker tutors, singular and "up to N".
		"Search your library for a planeswalker card, reveal it, put it into your hand, then shuffle.",
		"Search your library for up to two planeswalker cards, reveal them, put them into your hand, then shuffle.",
		// Permanent tutors: a plain permanent, a subtype-paired permanent
		// ("Rebel permanent"), and a legendary permanent, to the battlefield.
		"Search your library for a permanent card, put it onto the battlefield, then shuffle.",
		"Search your library for a Goblin permanent card, put it onto the battlefield, then shuffle.",
		"Search your library for an Elf permanent card, put it onto the battlefield, then shuffle.",
		"Search your library for a legendary Spirit permanent card, put it onto the battlefield, then shuffle.",
		// A "with mana value N or less" rider on a permanent, a typed card, or a
		// plural "up to N" search.
		"Search your library for a Rebel permanent card with mana value 5 or less, put it onto the battlefield, then shuffle.",
		"Search your library for an artifact card with mana value 1 or less, reveal it, put it into your hand, then shuffle.",
		"Search your library for up to two creature cards with mana value 1 or less, reveal them, put them into your hand, then shuffle.",
		// A "legendary" supertype on a typed card.
		"Search your library for a legendary creature card, reveal it, put it into your hand, then shuffle.",
		// Singular search-to-top tutors shuffle before replacing the found card.
		"Search your library for a card, then shuffle and put that card on top.",
		"Search your library for an artifact or enchantment card, reveal it, then shuffle and put that card on top.",
		// Instant- and sorcery-card tutors: a single spell type and a spell-type
		// union. The found card is named by the interchangeable "the card"
		// demonstrative as well as "it"/"that card".
		"Search your library for a sorcery card, reveal it, then shuffle and put that card on top.",
		"Search your library for an instant card, reveal it, put it into your hand, then shuffle.",
		"Search your library for an instant or sorcery card, reveal it, then shuffle and put that card on top.",
		"Search your library for a creature card, reveal it, then shuffle and put the card on top.",
		// Search-to-graveyard tutors (Entomb, Buried Alive), singular and "up to
		// N", with a plain card or a typed card filter.
		"Search your library for a card, put that card into your graveyard, then shuffle.",
		"Search your library for a creature card, put it into your graveyard, then shuffle.",
		"Search your library for up to three creature cards, put them into your graveyard, then shuffle.",
		// A trailing rider (random discard or fixed life loss) may sit between the
		// put phrase and the closing "then shuffle." (Gamble, Diabolic Tutor-style
		// life payments); the search clause itself stays exact.
		"Search your library for a card, put that card into your hand, discard a card at random, then shuffle.",
		"Search your library for a creature card, put it into your hand, you lose 2 life, then shuffle.",
	}
	for _, source := range accepted {
		if !searchExact(t, source) {
			t.Errorf("searchExact(%q) = false, want true", source)
		}
	}
}

func TestExactLibrarySearchFailsClosed(t *testing.T) {
	t.Parallel()
	// Each carries a rider the runtime SearchSpec cannot faithfully express, so
	// the round-trip must fail closed rather than lower to a wrong predicate.
	rejected := []string{
		// Non-library or extra source zone.
		"Search your library and graveyard for a creature card, put it into your hand, then shuffle.",
		// A multi-type union exceeds the single-type SearchSpec.
		"Search your library for an artifact creature card, put it onto the battlefield, then shuffle.",
		// Mana-value riders other than a fixed "or less" bound are not modeled.
		"Search your library for a creature card with mana value 3 or greater, put it into your hand, then shuffle.",
		"Search your library for a permanent card with mana value X or less, put it onto the battlefield, then shuffle.",
		// "different names" and variable counts.
		"Search your library for up to two basic land cards with different names, put them onto the battlefield tapped, then shuffle.",
		"Search your library for up to X basic land cards, put them onto the battlefield tapped, then shuffle.",
		// Unsupported destinations and ordering.
		"Search your library for a card, put it on top of your library, then shuffle.",
		"Search your library for a card, put that card on top, then shuffle.",
		"Search your library for up to two cards, then shuffle and put those cards on top.",
		"Search your library for a card, then shuffle and put that card on the bottom.",
		"Search your library for a card, then shuffle and put that card on top at random.",
		"Search your library for a card, then shuffle and put that card in the top three cards of your library.",
	}

	for _, source := range rejected {
		if searchExact(t, source) {
			t.Errorf("searchExact(%q) = true, want false", source)
		}
	}
}

func TestExactLibraryTopSearchCarriesTypedDestinationAndFilter(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Search your library for an artifact or enchantment card, reveal it, then shuffle and put that card on top.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	search := document.Abilities[0].Sentences[0].Effects[0]
	if !search.Exact || search.SearchDestination != EffectDestinationTop {
		t.Fatalf("search = %#v, want exact typed top destination", search)
	}
	if search.Selection.Kind != SelectionArtifact ||
		!slices.Equal(search.Selection.RequiredTypesAny, []CardType{CardTypeArtifact, CardTypeEnchantment}) {
		t.Fatalf("selection = %#v, want artifact-or-enchantment filter", search.Selection)
	}
}

// searchExactOptional parses a single optional library-search sentence ("You may
// search ...") and reports both that its leading search effect carries the
// resolving optionality and whether it round-tripped to an exact production. The
// "you may" prefix must not defeat exact recognition.
func searchExactOptional(t *testing.T, source string) (optional, exact bool) {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) == 0 || effects[0].Kind != EffectSearch {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Optional, effects[0].Exact
}

func TestExactOptionalLibrarySearchAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"You may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		"You may search your library for a creature card, reveal it, put it into your hand, then shuffle.",
		"You may search your library for a Goblin card, reveal it, put it into your hand, then shuffle.",
		"You may search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
		"You may search your library for an instant or sorcery card, reveal it, put it into your hand, then shuffle.",
		"You may search your library for a green creature card, reveal it, put it into your hand, then shuffle.",
	}
	for _, source := range accepted {
		optional, exact := searchExactOptional(t, source)
		if !optional {
			t.Errorf("searchExactOptional(%q) optional = false, want true", source)
		}
		if !exact {
			t.Errorf("searchExactOptional(%q) exact = false, want true", source)
		}
	}
}

func TestExactOptionalLibrarySearchFailsClosed(t *testing.T) {
	t.Parallel()
	// The optional prefix must not relax the filter/shape envelope: an
	// unsupported filter stays non-exact even when wrapped in "you may".
	rejected := []string{
		"You may search your library and graveyard for a creature card, put it into your hand, then shuffle.",
	}
	for _, source := range rejected {
		if _, exact := searchExactOptional(t, source); exact {
			t.Errorf("searchExactOptional(%q) exact = true, want false", source)
		}
	}
}

// riderSearchEffect parses a removal-plus-rider spell ("Exile target creature.
// Its controller may search their library for ...") and returns the search
// effect of the second sentence so a test can assert its optionality and exact
// round-trip. The affected-permanent's-controller searcher ("Its controller may
// search their library") must reconstruct byte-for-byte just like the controller
// "Search your library" form.
func riderSearchEffect(t *testing.T, source string) (optional, exact bool) {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 2 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[1].Effects
	if len(effects) == 0 || effects[0].Kind != EffectSearch {
		t.Fatalf("Parse(%q) rider effects = %#v", source, effects)
	}
	return effects[0].Optional, effects[0].Exact
}

func TestExactControllerSearchRiderAccepts(t *testing.T) {
	t.Parallel()
	// The Path to Exile / Assassin's Trophy rider: the affected permanent's
	// controller optionally fetches a basic land. The "Its controller may search
	// their library" subject is the only modeled non-controller searcher and must
	// round-trip exact with the optionality preserved.
	accepted := []string{
		"Exile target creature. Its controller may search their library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		"Exile target creature. Its controller may search their library for a basic land card, put that card onto the battlefield tapped, then shuffle.",
		"Destroy target nonbasic land. Its controller may search their library for a basic land card, put it onto the battlefield, then shuffle.",
		// The possessive subject generalizes beyond the creature pronoun "Its" to
		// any "That <permanent>'s controller" back-reference (Demolition Field).
		"Destroy target nonbasic land an opponent controls. That land's controller may search their library for a basic land card, put it onto the battlefield, then shuffle.",
		// A color filter behind the rider prefix round-trips exact.
		"Exile target creature. Its controller may search their library for a green creature card, put it onto the battlefield, then shuffle.",
	}
	for _, source := range accepted {
		optional, exact := riderSearchEffect(t, source)
		if !optional {
			t.Errorf("riderSearchEffect(%q) optional = false, want true", source)
		}
		if !exact {
			t.Errorf("riderSearchEffect(%q) exact = false, want true", source)
		}
	}
}

func TestExactControllerSearchRiderFailsClosed(t *testing.T) {
	t.Parallel()
	// The rider prefix relaxes neither the searcher nor the filter/shape
	// envelope. A non-optional rider, a controller-owned-library searcher that is
	// not the bounded "Search your library" form, and an unsupported filter all
	// stay non-exact.
	rejected := []string{
		// No "may": a mandatory "Its controller searches their library" is not the
		// modeled optional rider and must not reconstruct as exact.
		"Exile target creature. Its controller searches their library for a basic land card, puts it onto the battlefield tapped, then shuffles.",
		// Unsupported destination behind the rider prefix.
		"Exile target creature. Its controller may search their library for a basic land card, put it into their graveyard, then shuffle.",
	}
	for _, source := range rejected {
		if _, exact := riderSearchEffect(t, source); exact {
			t.Errorf("riderSearchEffect(%q) exact = true, want false", source)
		}
	}
}

// embeddedSearchExact parses an ability whose library-search clause is not
// sentence-initial (it follows a triggered-ability condition such as "When this
// creature enters, "), so the verb is lowercase ("search"). It returns whether
// that search effect round-tripped to an exact, lowerable production. The
// lowercase verb must not defeat exact recognition: the embedded search lowers to
// the same production as a sentence-initial controller tutor.
func embeddedSearchExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	for _, sentence := range document.Abilities[0].Sentences {
		for i := range sentence.Effects {
			if sentence.Effects[i].Kind == EffectSearch {
				return sentence.Effects[i].Exact
			}
		}
	}
	t.Fatalf("Parse(%q) found no search effect", source)
	return false
}

func TestExactEmbeddedLibrarySearchAccepts(t *testing.T) {
	t.Parallel()
	// A triggered or otherwise non-sentence-initial controller tutor: the
	// leading clause ("When this creature enters, ", "Whenever this creature
	// mutates, ") leaves the search verb lowercase. The bounded filter/shape
	// envelope is unchanged from the sentence-initial form.
	accepted := []string{
		"When this creature enters, search your library for a Forest card, put that card onto the battlefield, then shuffle.",
		"When this creature enters, search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		"When this creature enters, search your library for a card, put it into your hand, then shuffle.",
		"When this enchantment enters, search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
		"When this creature enters, search your library for a basic Swamp card, reveal it, put it into your hand, then shuffle.",
		"When this creature enters, search your library for a green creature card, put it onto the battlefield, then shuffle.",
	}
	for _, source := range accepted {
		if !embeddedSearchExact(t, source) {
			t.Errorf("embeddedSearchExact(%q) = false, want true", source)
		}
	}
}

func TestExactEmbeddedLibrarySearchFailsClosed(t *testing.T) {
	t.Parallel()
	// The lowercase embedded verb relaxes neither the destination nor the
	// remaining shape envelope: an unsupported "on top of library" destination
	// stays non-exact.
	rejected := []string{
		"When this creature enters, search your library for a card, put it on top of your library, then shuffle.",
	}
	for _, source := range rejected {
		if embeddedSearchExact(t, source) {
			t.Errorf("embeddedSearchExact(%q) = true, want false", source)
		}
	}
}

// TestTrimLeadingInterveningCondition verifies the leading intervening-if
// condition clause is removed so the search clause that follows can reconstruct
// byte-exactly. Only a recognized condition intro followed by a comma is
// stripped; ordinary search wording is left intact.
func TestTrimLeadingInterveningCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "if condition before optional search",
			text: "if an opponent controls more lands than you, you may search your library for up to three basic land cards, reveal them, put them into your hand, then shuffle.",
			want: "you may search your library for up to three basic land cards, reveal them, put them into your hand, then shuffle.",
		},
		{
			name: "if condition before mandatory search",
			text: "if you control three or more creatures, search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			want: "search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
		},
		{
			name: "no condition leaves text unchanged",
			text: "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			want: "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := trimLeadingInterveningCondition(test.text); got != test.want {
				t.Fatalf("trimLeadingInterveningCondition(%q) = %q, want %q", test.text, got, test.want)
			}
		})
	}
}

// TestInterveningConditionSearchEffectStaysExact confirms a triggered ability
// that gates a library search behind an intervening-if condition still produces
// a supported (UnsupportedDetail-free) search effect, so the search lowers even
// though its effect text retains the leading condition clause.
func TestInterveningConditionSearchEffectStaysExact(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"At the beginning of your upkeep, if an opponent controls more lands than you, "+
			"you may search your library for up to three basic land cards, reveal them, "+
			"put them into your hand, then shuffle.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	var search *EffectSyntax
	for i := range document.Abilities[0].Sentences {
		for j := range document.Abilities[0].Sentences[i].Effects {
			effect := &document.Abilities[0].Sentences[i].Effects[j]
			if effect.Kind == EffectSearch {
				search = effect
			}
		}
	}
	if search == nil {
		t.Fatalf("no search effect found in %#v", document.Abilities[0].Sentences)
	}
	if search.UnsupportedDetail != "" {
		t.Fatalf("search.UnsupportedDetail = %q, want empty", search.UnsupportedDetail)
	}
}
