package cardgen

import (
	"strings"
	"testing"
)

// odricKeywordShareText is Odric, Lunarch Marshal's exact oracle text.
const odricKeywordShareText = "At the beginning of each combat, creatures you control gain first strike until end of turn if a creature you control has first strike. The same is true for flying, deathtouch, double strike, haste, hexproof, indestructible, lifelink, menace, reach, skulk, trample, and vigilance."

// odricSharedKeywords are Odric's thirteen shared runtime keyword names in
// printed order, each of which must lower to its own gated group grant.
var odricSharedKeywords = []string{
	"FirstStrike",
	"Flying",
	"Deathtouch",
	"DoubleStrike",
	"Haste",
	"Hexproof",
	"Indestructible",
	"Lifelink",
	"Menace",
	"Reach",
	"Skulk",
	"Trample",
	"Vigilance",
}

// TestGenerateExecutableCardSourceOdricLunarchMarshal proves the team
// keyword-sharing construct lowers to one continuous group grant per shared
// keyword, each gated on the controller controlling a creature that already has
// that keyword, fired at the beginning of each combat.
func TestGenerateExecutableCardSourceOdricLunarchMarshal(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Odric, Lunarch Marshal",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Soldier",
		ManaCost:   "{3}{W}",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: odricKeywordShareText,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	normalized := normalizeSource(source)
	if !strings.Contains(normalized, normalizeSource("game.StepBeginningOfCombat")) {
		t.Fatalf("source missing beginning-of-combat trigger:\n%s", source)
	}
	if got := strings.Count(normalized, normalizeSource("game.ApplyContinuous{")); got != len(odricSharedKeywords) {
		t.Fatalf("ApplyContinuous count = %d, want %d\n%s", got, len(odricSharedKeywords), source)
	}
	// Every keyword must appear as both the group grant (AddKeywords) to
	// creatures you control and the controller-scoped gate (ControlsMatching)
	// on a creature that already has it — never one without the other.
	for _, kw := range odricSharedKeywords {
		grant := "AddKeywords: []game.Keyword{game." + kw + "}"
		if !strings.Contains(normalized, normalizeSource(grant)) {
			t.Fatalf("source missing group grant for %s:\n%s", kw, source)
		}
		gate := "Keyword: game." + kw + "}"
		if !strings.Contains(normalized, normalizeSource(gate)) {
			t.Fatalf("source missing gate for %s:\n%s", kw, source)
		}
	}
	if !strings.Contains(normalized, normalizeSource("Controller: game.ControllerYou")) {
		t.Fatalf("group grant is not scoped to creatures you control:\n%s", source)
	}
	if !strings.Contains(normalized, normalizeSource("Duration: game.DurationUntilEndOfTurn")) {
		t.Fatalf("shared keywords are not granted until end of turn:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceKeywordShareFailsClosed proves every card that
// shares Odric's "the same is true for ..." surface but differs in trigger,
// subject, gate, or keyword set fails closed — no source is generated — rather
// than silently dropping a keyword or its gate.
func TestGenerateExecutableCardSourceKeywordShareFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		manaCost   string
		power      string
		toughness  string
		oracleText string
		reason     string
	}{
		{
			name:       "Concerted Effort",
			typeLine:   "Enchantment",
			manaCost:   "{2}{W}",
			oracleText: "At the beginning of each upkeep, creatures you control gain flying until end of turn if a creature you control has flying. The same is true for fear, first strike, double strike, landwalk, protection, trample, and vigilance.",
			reason:     "landwalk and protection are not grantable",
		},
		{
			name:       "Oddric, Lunar Marquis",
			typeLine:   "Legendary Creature — Human Soldier",
			manaCost:   "{2}{W}",
			power:      "3",
			toughness:  "3",
			oracleText: `At the beginning of each combat, creatures you control gain banding until end of turn if a creature you control has banding. The same is true for changeling, devoid, fear, flanking, horsemanship, ingest, intimidate, landwalk, shroud, tantrum, wither, and the activated ability "Sacrifice this creature: Add {C}."`,
			reason:     "banding lead keyword is not grantable",
		},
		{
			name:       "Bleeding Effect",
			typeLine:   "Enchantment",
			manaCost:   "{2}{B}",
			oracleText: "At the beginning of combat on your turn, creatures you control gain flying until end of turn if a creature card in your graveyard has flying. The same is true for first strike, double strike, deathtouch, hexproof, indestructible, lifelink, menace, reach, trample, and vigilance.",
			reason:     "gate is a creature card in your graveyard, not a creature you control",
		},
		{
			name:       "Majestic Myriarch",
			typeLine:   "Creature — Chimera",
			manaCost:   "{3}{G}",
			power:      "0",
			toughness:  "0",
			oracleText: "Majestic Myriarch's power and toughness are each equal to twice the number of creatures you control.\nAt the beginning of each combat, this creature gains flying until end of turn if you control a creature with flying. The same is true for first strike, double strike, deathtouch, haste, hexproof, indestructible, lifelink, menace, reach, trample, and vigilance.",
			reason:     "subject is this creature, not creatures you control",
		},
		{
			name:       "Thunderous Orator",
			typeLine:   "Creature — Human Soldier",
			manaCost:   "{2}{W}",
			power:      "2",
			toughness:  "2",
			oracleText: "Vigilance\nWhenever this creature attacks, it gains flying until end of turn if you control a creature with flying. The same is true for first strike, double strike, deathtouch, indestructible, lifelink, menace, and trample.",
			reason:     "attack trigger, not a phase/step trigger",
		},
		{
			name:       "Sproutwatch Dryad",
			typeLine:   "Creature — Dryad",
			manaCost:   "{3}{G}",
			power:      "2",
			toughness:  "3",
			oracleText: "At the beginning of each combat, Sproutwatch Dryad gains flying until end of turn if a creature you control or a card in your hand has flying. The same is true for first strike, double strike, deathtouch, haste, hexproof, indestructible, lifelink, menace, reach, trample, and vigilance.",
			reason:     "self subject and expanded gate",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				ManaCost:   tc.manaCost,
				OracleText: tc.oracleText,
			}
			if tc.power != "" {
				card.Power = new(tc.power)
			}
			if tc.toughness != "" {
				card.Toughness = new(tc.toughness)
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, strings.ToLower(tc.name[:1]))
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("%s did not fail closed (%s): source=%q diagnostics=%#v", tc.name, tc.reason, source, diagnostics)
			}
		})
	}
}
