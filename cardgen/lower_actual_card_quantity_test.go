package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func TestActualCardQuantityPublisherRejectsNilEnvelope(t *testing.T) {
	envelope := game.Instruction{
		Optional: true, TemptingOffer: true,
		TemptingOfferBody: []game.Instruction{
			{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
		},
	}
	for _, outcome := range []compiler.EffectKind{compiler.EffectExile, compiler.EffectMill, compiler.EffectDiscard} {
		if actualCardQuantityPublisher(envelope.Primitive, outcome) {
			t.Fatalf("nil-primitive multi-body envelope accepted as a card-result publisher for %v", outcome)
		}
	}
}

func TestLowerActualCardQuantityConsumers(t *testing.T) {
	for _, text := range []string{
		"Exile all graveyards. For each creature card exiled this way, create a 2/2 black Zombie creature token.",
		"Exile target player's graveyard. For each artifact or land card exiled this way, create a 1/1 green Insect creature token.",
		"Mill three cards. For each creature card milled this way, draw two cards.",
		"Mill three cards. For each creature card milled this way, you gain 2 life.",
		"Discard two cards. For each nonland card discarded this way, create a 1/1 red Human creature token.",
		"Mill three cards. Draw a card. For each creature card milled this way, you gain 2 life.",
		"Exile target player's graveyard. For each artifact or land card exiled this way, create a 1/1 colorless Insect artifact creature token with flying.",
		"Mill three cards. For each creature card milled this way, create a 2/2 black Necron Warrior artifact creature token.",
		"Exile target opponent's graveyard. For each creature card exiled this way, create a 1/1 colorless Spirit creature token.",
		"Discard two cards, then draw two cards. For each nonland card discarded this way, create a 1/1 red Elemental creature token.",
		"Mill two cards. For each creature card milled this way, create a 1/1 green Saproling creature token.",
		"Discard two cards. You gain 2 life for each card discarded this way.",
		"Mill three cards. For each blue creature card milled this way, you gain 2 life.",
		"Mill three cards. For each permanent card milled this way, you gain 2 life.",
		"Mill three cards. For each nonland permanent card milled this way, you gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			card := &ScryfallCard{Name: "Quantity", Layout: "normal", TypeLine: "Sorcery", OracleText: text}
			face := lowerSingleFace(t, card)
			sequence := face.SpellAbility.Val.Modes[0].Sequence
			last := sequence[len(sequence)-1]
			if sequence[0].PublishResult != actualCardQuantityKey(1) || !last.ResultGate.Exists ||
				last.ResultGate.Val.Key != sequence[0].PublishResult || last.ResultGate.Val.Accepted != game.TriTrue ||
				last.ResultGate.Val.AmountAvailable {
				t.Fatalf("sequence=%#v", sequence)
			}
		})
	}
}

func TestActualCardQuantityPreservesEventCountDomain(t *testing.T) {
	for _, payoff := range []string{
		"draw a card for each card discarded this way.",
		"for each card discarded this way, draw a card.",
	} {
		card := &ScryfallCard{Name: "Event count", Layout: "normal", TypeLine: "Enchantment",
			OracleText: "Whenever you discard one or more cards, " + payoff}
		assertCardPaths(t, card,
			"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Draw).Amount.dynamic.Kind = game.DynamicAmountEventCardCount")
	}
}

func TestLowerActualCardQuantityNearMisses(t *testing.T) {
	for _, text := range []string{
		"For each creature card milled this way, draw a card. Mill three cards.",
		"Exile target creature. For each creature card exiled this way, draw a card.",
		"Mill three cards. For each creature card milled this way with flying, draw a card.",
		"Mill three cards. For each creature card milled this way, draw a card at the beginning of the next end step.",
		"Mill three cards. For each creature milled this way, draw a card.",
		"Mill three cards. For each creature card milled this way, each player draws a card.",
		"Mill three cards. For each creature card milled this way, draw X cards.",
		"Mill three cards. For each creature card with power 2 milled this way, draw a card.",
		"Mill three cards. For each creature card with toughness * milled this way, draw a card.",
		"Mill three cards. For each creature card with mana value 2 milled this way, draw a card.",
		"Mill three cards. For each creature card you control milled this way, draw a card.",
		"Exile two target cards from a graveyard. For each creature card exiled this way, draw a card.",
		"Exile X target creature cards from your graveyard. For each creature card exiled this way, draw a card.",
		"You may mill three cards. If you do, for each creature card milled this way, draw a card.",
		"Mill three cards. For each blue artifact or red creature card milled this way, draw a card.",
		"Mill three cards. For each noncreature artifact or land card milled this way, draw a card.",
		"Mill three cards. For each artifact and land card milled this way, draw a card.",
		"Mill three cards. For each artifact card or land card milled this way, draw a card.",
		"Mill three cards. For each artifact creature card milled this way, you gain 2 life.",
		"Mill two cards. For each creature card milled this way, draw a card and gain 1 life.",
		"Mill two cards. For each creature card milled this way, draw a card, then you gain 1 life.",
		"Mill two cards. For each creature card milled this way, draw a card and create a 1/1 green Insect creature token.",
		"Exile the top card of your library face down. For each creature card exiled this way, you gain 2 life.",
		"Mill a card. If a creature card was put into your graveyard this way, you gain 2 life.",
		"Exile up to two target cards from a single graveyard. If this spell was kicked, instead exile target player's graveyard. Create a 2/2 black Rogue creature token for each creature card exiled this way.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{Name: "Near Miss", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
		})
	}
}
