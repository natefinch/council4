package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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

func TestFixedOrdinaryOptionalMixedManaCondition(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{Name: "Test Mana", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: Target creature gains trample until end of turn. If you have no cards in hand, you may add {R}{G}."}
	assertCardPaths(t, card,
		"ActivatedAbilities[0].Content.Modes[0].Targets[0].Constraint = \"target creature\"",
		"Sequence[1].Condition.Val.Condition.Val.ControllerHandEmpty = true",
		"Sequence[1].Optional = true",
		"Sequence[1].PublishCondition",
		"Sequence[1].PublishOptionalDecision",
		"Sequence[2].ConditionGate",
		"Sequence[2].OptionalDecisionGate")
	assertCardPathsAbsent(t, card,
		"ActivatedAbilities[1]", "ManaAbilities[", "ActivationCondition.Exists = true",
		"Sequence[3]", "Sequence[0].Condition", "Sequence[0].PublishCondition", "Sequence[0].ConditionGate",
		"Sequence[0].Optional = true", "Sequence[0].PublishOptionalDecision", "Sequence[0].OptionalDecisionGate",
		"Sequence[2].Condition.Val.", "Sequence[2].PublishCondition", "Sequence[2].Optional = true",
		"Sequence[2].PublishOptionalDecision", "ResultGate.Exists = true", "PublishResult")

	defs, diagnostics, err := CompileCardDefs(card)
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile: err=%v diagnostics=%v defs=%d", err, diagnostics, len(defs))
	}
	def := defs[0]
	if len(def.ActivatedAbilities) != 1 || len(def.ManaAbilities) != 0 || def.ActivatedAbilities[0].ActivationCondition.Exists {
		t.Fatal("resolving mana condition became a mana ability or activation restriction")
	}
	content := def.ActivatedAbilities[0].Content
	if len(content.Modes) != 1 || len(content.Modes[0].Targets) != 1 || len(content.Modes[0].Sequence) != 3 {
		t.Fatal("ordinary activated body must retain one creature target and exactly three instructions")
	}
	mode := content.Modes[0]
	target := mode.Targets[0]
	if target.Constraint != "target creature" || target.MinTargets != 1 || target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowPermanent || !target.Selection.Exists ||
		len(target.Selection.Val.RequiredTypesAny) != 1 || target.Selection.Val.RequiredTypesAny[0] != types.Creature {
		t.Fatalf("original target constraint changed: %+v", target)
	}
	trample, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || !trample.Object.Exists || trample.Object.Val != game.TargetPermanentReference(0) ||
		trample.Duration != game.DurationUntilEndOfTurn || len(trample.ContinuousEffects) != 1 {
		t.Fatalf("independent trample instruction lost its target or duration: %+v", mode.Sequence[0])
	}
	effect := trample.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility || len(effect.AddKeywords) != 1 || effect.AddKeywords[0] != game.Trample {
		t.Fatalf("independent target keyword changed: %+v", effect)
	}
	for i, color := range []mana.Color{mana.R, mana.G} {
		add, ok := mode.Sequence[i+1].Primitive.(game.AddMana)
		if !ok || add.Amount.IsDynamic() || add.Amount.Value() != 1 || add.ManaColor != color || add.Player.Exists {
			t.Fatalf("output %d must add one fixed %v mana to the resolving controller: %+v", i+1, color, mode.Sequence[i+1])
		}
	}
	first, second := mode.Sequence[1], mode.Sequence[2]
	if !first.Condition.Exists || !first.Condition.Val.Condition.Exists ||
		!first.Condition.Val.Condition.Val.ControllerHandEmpty || first.Condition.Val.Condition.Val.Negate ||
		first.PublishCondition == "" || first.ConditionGate != "" || first.ConditionGateNegate ||
		second.Condition.Exists || second.PublishCondition != "" ||
		second.ConditionGate != first.PublishCondition || second.ConditionGateNegate {
		t.Fatal("printed hand-empty condition must be evaluated once at the first mana output")
	}
	if !first.Optional || first.PublishOptionalDecision == "" || first.OptionalDecisionGate != "" ||
		second.Optional || second.PublishOptionalDecision != "" ||
		second.OptionalDecisionGate != first.PublishOptionalDecision {
		t.Fatal("both mana outputs must share exactly one acceptance decision")
	}
	for i, instruction := range mode.Sequence {
		if instruction.ResultGate.Exists || instruction.PublishResult != "" {
			t.Fatalf("instruction %d invented a primitive-success gate for condition or acceptance", i)
		}
	}
}

func TestFixedOrdinaryManaRefusals(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"{T}: You may add {R}{R}.",
		"{1}: If this is the third time this ability has resolved this turn, add {R}{R}.",
		"{1}: Draw a card. If this is the third time this ability has resolved this turn, add {R}{R}.",
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
