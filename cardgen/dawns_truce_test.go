package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const dawnsTruceOracle = "Gift a card (You may promise an opponent a gift as you cast this spell. If you do, they draw a card before its other effects.)\n" +
	"You and permanents you control gain hexproof until end of turn. If the gift was promised, permanents you control also gain indestructible until end of turn."

func dawnsTruceCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Dawn's Truce",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{W}",
		OracleText: dawnsTruceOracle,
	}
}

// TestLowerDawnsTruce proves the compound "you and permanents you control gain
// hexproof" subject composes a player-scoped hexproof rule effect with a group
// keyword grant, and that the gift-gated "permanents you control also gain
// indestructible" rider lowers to a group grant guarded by the gift-promised
// condition. No card name drives the composition.
func TestLowerDawnsTruce(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, dawnsTruceCard())

	// Gift keyword: the promised opponent draws a card before other effects.
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1 (Gift)", len(face.StaticAbilities))
	}
	keyword, ok := game.BodyKeywordAbility(&face.StaticAbilities[0].Body, game.Gift)
	if !ok {
		t.Fatalf("Gift keyword not found in %#v", face.StaticAbilities[0].Body)
	}
	gift, ok := keyword.(game.GiftKeyword)
	if !ok {
		t.Fatalf("keyword = %T, want game.GiftKeyword", keyword)
	}
	if len(gift.Delivery.Modes[0].Sequence) != 1 {
		t.Fatalf("gift delivery = %#v, want a single draw instruction", gift.Delivery)
	}
	draw, ok := gift.Delivery.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Player != game.GiftRecipientReference() {
		t.Fatalf("gift delivery primitive = %#v, want recipient draw", gift.Delivery.Modes[0].Sequence[0].Primitive)
	}

	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("spell ability = %+v, want one mode", face.SpellAbility)
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence length = %d, want 3: %+v", len(sequence), sequence)
	}

	// [0] player hexproof rule effect for you, until end of turn, ungated.
	rule, ok := sequence[0].Primitive.(game.ApplyRule)
	if !ok || rule.Duration != game.DurationUntilEndOfTurn ||
		len(rule.RuleEffects) != 1 ||
		rule.RuleEffects[0].Kind != game.RuleEffectPlayerHexproof ||
		rule.RuleEffects[0].AffectedPlayer != game.PlayerYou {
		t.Fatalf("sequence[0] = %+v, want until-end-of-turn player hexproof for you", sequence[0])
	}
	if sequence[0].Condition.Exists {
		t.Fatalf("player hexproof must be ungated, got %+v", sequence[0].Condition)
	}

	// [1] group hexproof grant over permanents you control, ungated.
	hexGrant := requireControlledGroupKeyword(t, sequence[1], game.Hexproof)
	if sequence[1].Condition.Exists {
		t.Fatalf("permanent hexproof must be ungated, got %+v", sequence[1].Condition)
	}
	_ = hexGrant

	// [2] group indestructible grant over permanents you control, gated on the
	// gift being promised.
	requireControlledGroupKeyword(t, sequence[2], game.Indestructible)
	if !sequence[2].Condition.Exists ||
		!sequence[2].Condition.Val.Condition.Exists ||
		!sequence[2].Condition.Val.Condition.Val.GiftPromised {
		t.Fatalf("sequence[2] must be gated on gift promised, got %+v", sequence[2].Condition)
	}
}

// TestLowerPlayerAndPermanentsHexproofNoGift proves the same compound subject
// lowers without a gift branch: the plain "you and permanents you control gain
// hexproof until end of turn" clause (Lazotep Plating's rider) produces exactly
// the player rule + ungated group grant, with no indestructible and no gift gate.
func TestLowerPlayerAndPermanentsHexproofNoGift(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ward",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{U}",
		OracleText: "You and permanents you control gain hexproof until end of turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2: %+v", len(sequence), sequence)
	}
	rule, ok := sequence[0].Primitive.(game.ApplyRule)
	if !ok || len(rule.RuleEffects) != 1 ||
		rule.RuleEffects[0].Kind != game.RuleEffectPlayerHexproof {
		t.Fatalf("sequence[0] = %+v, want player hexproof rule", sequence[0])
	}
	requireControlledGroupKeyword(t, sequence[1], game.Hexproof)
	for i, instruction := range sequence {
		if instruction.Condition.Exists {
			t.Fatalf("sequence[%d] must be ungated, got %+v", i, instruction.Condition)
		}
	}
}

func TestGenerateDawnsTruceSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(dawnsTruceCard(), "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	for _, want := range []string{
		"game.GiftKeyword{",
		"Player: game.GiftRecipientReference(),",
		"game.RuleEffectPlayerHexproof",
		"AffectedPlayer: game.PlayerYou",
		"game.Hexproof,",
		"game.Indestructible,",
		"GiftPromised: true,",
		"game.DurationUntilEndOfTurn,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestDawnsTruceVariantsFailClosed proves the composition fails closed for
// keywords that have no player-scoped rule-effect equivalent: "you and
// permanents you control" only composes for keywords a player can also carry
// (hexproof, shroud). A trample grant to a player is meaningless and must not
// silently drop the player from the compound subject.
func TestDawnsTruceVariantsFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		// Trample has no player-scoped rule effect: the compound subject cannot
		// grant it to a player, so the whole sentence fails closed.
		"You and permanents you control gain trample until end of turn.",
		// A permanent keyword with no player analogue fails the same way.
		"You and permanents you control gain flying until end of turn.",
	} {
		face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
			Name:       "Test Ward",
			Layout:     "normal",
			TypeLine:   "Instant",
			ManaCost:   "{1}{W}",
			OracleText: oracleText,
		})
		if face.SpellAbility.Exists {
			t.Fatalf("unsupported variant produced a spell ability: %q", oracleText)
		}
	}
}

// requireControlledGroupKeyword asserts the instruction is an until-end-of-turn
// continuous grant of exactly the given keyword over the permanents you control.
func requireControlledGroupKeyword(t *testing.T, instruction game.Instruction, keyword game.Keyword) game.ContinuousEffect {
	t.Helper()
	cont, ok := instruction.Primitive.(game.ApplyContinuous)
	if !ok || cont.Duration != game.DurationUntilEndOfTurn ||
		len(cont.ContinuousEffects) != 1 {
		t.Fatalf("instruction = %+v, want until-end-of-turn continuous grant", instruction)
	}
	grant := cont.ContinuousEffects[0]
	if !grant.Group.Valid() {
		t.Fatalf("grant group = %+v, want the permanents you control", grant.Group)
	}
	if selection := grant.Group.Selection(); selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 0 {
		t.Fatalf("grant selection = %+v, want all permanents you control", grant.Group.Selection())
	}
	if len(grant.AddKeywords) != 1 || grant.AddKeywords[0] != keyword {
		t.Fatalf("grant keywords = %+v, want [%v]", grant.AddKeywords, keyword)
	}
	return grant
}
