package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func temptWithVengeanceCard() *ScryfallCard {
	return &ScryfallCard{
		Name:          "Tempt with Vengeance",
		Layout:        "normal",
		ManaCost:      "{X}{R}",
		TypeLine:      "Sorcery",
		Colors:        []string{"R"},
		ColorIdentity: []string{"R"},
		SetType:       "commander",
		Games:         []string{"paper", "mtgo"},
		OracleText:    "Tempting offer \u2014 Create X 1/1 red Elemental creature tokens with haste. Each opponent may create X 1/1 red Elemental creature tokens with haste. For each opponent who does, create X 1/1 red Elemental creature tokens with haste.",
	}
}

// TestGenerateExecutableCardSourceTemptWithVengeance proves Tempt with Vengeance
// generates end to end: the "Tempting offer" idiom lowers to a single
// TemptingOffer instruction that offers the opponents an X-count red Elemental
// token creation, addressing the acting player through
// GroupOfferMemberReference().
func TestGenerateExecutableCardSourceTemptWithVengeance(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(temptWithVengeanceCard(), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"TemptingOffer:      true,",
		"OptionalActorGroup: opt.Val(game.OpponentsReference()),",
		"Recipient: opt.Val(game.GroupOfferMemberReference()),",
		"Kind: game.DynamicAmountX,",
		"game.HasteStaticBody,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerTemptWithVengeanceTemptingOffer proves the idiom lowers to exactly one
// optional group-offer instruction flagged TemptingOffer, offering the opponents
// a controller-addressed X-count token creation.
func TestLowerTemptWithVengeanceTemptingOffer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, temptWithVengeanceCard())
	if !face.SpellAbility.Exists {
		t.Fatal("Tempt with Vengeance produced no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("spell modes = %#v, want one single-instruction mode", modes)
	}
	instr := modes[0].Sequence[0]
	if !instr.TemptingOffer {
		t.Fatal("instruction is not flagged TemptingOffer")
	}
	if !instr.Optional {
		t.Fatal("TemptingOffer instruction is not optional")
	}
	if !instr.OptionalActorGroup.Exists ||
		instr.OptionalActorGroup.Val.Kind != game.PlayerGroupReferenceOpponents {
		t.Fatalf("OptionalActorGroup = %#v, want opponents", instr.OptionalActorGroup)
	}
	token, ok := instr.Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %#v, want game.CreateToken", instr.Primitive)
	}
	if !token.Recipient.Exists ||
		token.Recipient.Val.Kind() != game.PlayerReferenceGroupOfferMember {
		t.Fatalf("recipient = %#v, want GroupOfferMember", token.Recipient)
	}
	if amount := token.Amount.DynamicAmount(); !amount.Exists || amount.Val.Kind != game.DynamicAmountX {
		t.Fatalf("token amount = %#v, want dynamic X", token.Amount)
	}
}

// TestTemptCycleSiblingsFailClosed proves every Tempt-cycle card the backend does
// not yet model fails closed rather than mis-lowering. Each carries the "Tempting
// offer" ability word (or the capital-O "Tempting Offer" variant), which must
// never fall through to generic lowering and silently drop the each-opponent
// offer or the reward repeat.
func TestTemptCycleSiblingsFailClosed(t *testing.T) {
	t.Parallel()
	siblings := []*ScryfallCard{
		{Name: "Tempt with Discovery", Layout: "normal", ManaCost: "{3}{G}", TypeLine: "Sorcery", Colors: []string{"G"}, ColorIdentity: []string{"G"}, SetType: "commander", Games: []string{"paper", "mtgo"},
			OracleText: "Tempting offer \u2014 Search your library for a land card and put it onto the battlefield. Each opponent may search their library for a land card and put it onto the battlefield. For each opponent who searches a library this way, search your library for a land card and put it onto the battlefield. Then each player who searched a library this way shuffles."},
		{Name: "Tempt with Mayhem", Layout: "normal", ManaCost: "{1}{R}{R}", TypeLine: "Instant", Colors: []string{"R"}, ColorIdentity: []string{"R"}, SetType: "commander", Games: []string{"paper", "mtgo"},
			OracleText: "Tempting offer \u2014 Choose target instant or sorcery spell. Each opponent may copy that spell and may choose new targets for the copy they control. You copy that spell once plus an additional time for each opponent who copied the spell this way. You may choose new targets for the copies you control."},
		{Name: "Tempt with Bunnies", Layout: "normal", ManaCost: "{2}{W}", TypeLine: "Sorcery", Colors: []string{"W"}, ColorIdentity: []string{"W"}, SetType: "commander", Games: []string{"paper", "mtgo"},
			OracleText: "Tempting Offer \u2014 Draw a card and create a 1/1 white Rabbit creature token. Then each opponent may draw a card and create a 1/1 white Rabbit creature token. For each opponent who does, you draw a card and you create a 1/1 white Rabbit creature token."},
	}
	for _, card := range siblings {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, card)
			if face.SpellAbility.Exists {
				t.Fatalf("%s produced a spell ability but must fail closed", card.Name)
			}
		})
	}
}
