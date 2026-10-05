package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestFixedManaChannelBoundary(t *testing.T) {
	t.Parallel()
	assertCardPaths(t, &ScryfallCard{Name: "Plain Life", Layout: "normal", TypeLine: "Sorcery", OracleText: "You gain 1 life."},
		"Sequence[0].Primitive.(game.GainLife).Amount.fixed = 1")
	assertCardUnsupported(t, &ScryfallCard{Name: "Timed Life", Layout: "normal", TypeLine: "Sorcery", OracleText: "Until end of turn, you gain 1 life."},
		"unsupported life action duration")
	text := "Until end of turn, any time you could activate a mana ability, you may pay 1 life. If you do, add {C}."
	assertCardUnsupported(t, &ScryfallCard{Name: "Channel", Layout: "normal", TypeLine: "Sorcery", OracleText: text},
		"unsupported life action duration")
}

func TestFixedManaActivationRestrictions(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{Name: "Restricted Mana Body", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: Target creature gains trample until end of turn. Add {R}{R}. Activate only if you have no cards in hand."}
	assertCardPaths(t, card, "ActivatedAbilities[0].ActivationCondition.Val.ControllerHandEmpty = true")
	assertCardPathsAbsent(t, card, "ManaAbilities[", "Sequence[0].Condition", "Sequence[1].Condition")
	card.OracleText = "{1}: Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, add {R}{R}. Activate only if you have no cards in hand."
	assertCardUnsupported(t, card, "unsupported activation condition")
}

const flamekinBody = "{2}: Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}{R}{R}{R}{R}{R}{R}."

func TestFixedOrdinaryManaCompositions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, typeLine, text, path string
		optional                   bool
		amount, total              int
	}{
		{"ordinal optional", "Creature", flamekinBody, "ActivatedAbilities[0]", true, 8, 8},
		{"ordinal mandatory", "Creature", "{R}: Target creature you control gains trample until end of turn. If this is the third time this ability has resolved this turn, add {R}{R}{R}{R}.", "ActivatedAbilities[0]", false, 1, 4},
		{"state activated", "Artifact", "{1}: Target creature gains first strike until end of turn. If you have no cards in hand, you may add {R}{R}.", "ActivatedAbilities[0]", true, 2, 2},
		{"unless activated", "Artifact", "{1}: Target creature gains trample until end of turn. Unless you control another creature, add {R}{R}.", "ActivatedAbilities[0]", false, 1, 2},
		{"group spell", "Instant", "If you have no cards in hand, draw a card, then add {R}{R}.", "SpellAbility", false, 1, 2},
		{"optional spell", "Instant", "If you have no cards in hand, you may add {R}{R}.", "SpellAbility", true, 2, 2},
		{"trigger body", "Artifact", "At the beginning of your upkeep, draw a card. If you control a creature, you may add {R}{R}.", "TriggeredAbilities[0]", true, 2, 2},
		{"ordinal trigger", "Artifact", "Whenever a creature you control enters, draw a card. If this is the third time this ability has resolved this turn, you may add {R}{R}.", "TriggeredAbilities[0]", true, 2, 2},
		{"ordinal mode", "Artifact", "{1}: Choose one \u2014\n\u2022 Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}.\n\u2022 You gain 2 life.", "ActivatedAbilities[0]", true, 2, 2},
		{"optional result", "Instant", "You may add {R}{R}. If you do, draw a card.", "SpellAbility", true, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Test Mana", Layout: "normal", TypeLine: tc.typeLine, OracleText: tc.text}
			assertCardPaths(t, card, tc.path)
			defs, diagnostics, err := CompileCardDefs(card)
			if err != nil || len(diagnostics) != 0 {
				t.Fatalf("compile: %v %v", err, diagnostics)
			}
			face := &defs[0].CardFace
			var content game.AbilityContent
			switch {
			case len(face.ActivatedAbilities) != 0:
				content = face.ActivatedAbilities[0].Content
				if len(face.ManaAbilities) != 0 || face.ActivatedAbilities[0].ActivationCondition.Exists {
					t.Fatal("targeted body was classified as mana or its resolving gate as an activation restriction")
				}
			case len(face.TriggeredAbilities) != 0:
				content = face.TriggeredAbilities[0].Content
			default:
				content = face.SpellAbility.Val
			}
			found := false
			total := 0
			for _, instruction := range content.Modes[0].Sequence {
				add, ok := instruction.Primitive.(game.AddMana)
				if !ok {
					continue
				}
				found = true
				total += add.Amount.Value()
				if add.ManaColor != mana.R || add.Amount.Value() != tc.amount || instruction.Optional != tc.optional {
					t.Fatalf("mana output: %+v, optional=%v", add, instruction.Optional)
				}
			}
			if !found || total != tc.total {
				t.Fatalf("total fixed mana=%d, want %d", total, tc.total)
			}
		})
	}
}

func TestFixedOrdinaryManaRefusals(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"{T}: You may add {R}{R}.",
		"{1}: If this is the third time this ability has resolved this turn, add {R}{R}.",
		"{1}: Draw a card. If this is the third time this ability has resolved this turn, add {R}{R}.",
		"{1}: Target creature gains trample until end of turn. If you have no cards in hand, you may add {R}{G}.",
		"{1}: Target creature gains trample until end of turn. If you have no cards in hand, add {R}{X}.",
		"{1}: Target creature gains trample until end of turn. If you have no cards in hand, target opponent adds {R}{R}.",
		"{1}: Target creature gains trample until end of turn. If you have no cards in hand, add {R}{R} instead.",
		"You may add {R}{R}. If you do, draw a card. You gain 2 life.",
		"If you have no cards in hand, you may add {R}{R}. You gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Test Mana", Layout: "normal", TypeLine: "Artifact", OracleText: text})
		})
	}
}

func TestSoulbrightFixedMana(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{Name: "Soulbright Flamekin", Layout: "normal", TypeLine: "Creature \u2014 Elemental Shaman", ManaCost: "{1}{R}", Power: new("2"), Toughness: new("1"), OracleText: flamekinBody}
	assertCardPaths(t, card,
		"ActivatedAbilities[0].CountsResolutionsThisTurn = true",
		"Sequence[1].Primitive.(game.AddMana).Amount.fixed = 8",
		"Sequence[1].Primitive.(game.AddMana).ManaColor = mana.R",
		"Sequence[1].Optional = true")
	assertCardPathsAbsent(t, card, "ManaAbilities[", "ActivationCondition", "Sequence[0].Condition", "Sequence[2]")
}
