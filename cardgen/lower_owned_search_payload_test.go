package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestOwnedSearchPayloadReuse(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		ownedLandSearch,
		"Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
		"Search your library for a creature card with mana value 3 or less, put it onto the battlefield, then shuffle.",
		"Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
		"Search your library for any number of basic land cards, put them onto the battlefield, then shuffle.",
		"Target player searches their library for a basic land card, puts it onto the battlefield tapped, then shuffles.",
		"Each player searches their library for a basic land card, puts it onto the battlefield, then shuffles.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			var payloads []game.Search
			for _, body := range []string{text, "Tap target creature. " + text + " Scry 1."} {
				defs, diagnostics, err := CompileCardDefs(&ScryfallCard{Name: "Search Payload", Layout: "normal", TypeLine: "Sorcery", OracleText: body})
				if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
					t.Fatalf("compile: %v %v, definitions=%d", err, diagnostics, len(defs))
				}
				for _, instruction := range defs[0].SpellAbility.Val.Modes[0].Sequence {
					if search, ok := instruction.Primitive.(game.Search); ok {
						payloads = append(payloads, search)
					}
				}
			}
			if len(payloads) != 2 {
				t.Fatalf("search payloads=%d, want 2", len(payloads))
			}
			if payloads[0].Player.Kind() == game.PlayerReferenceTargetPlayer {
				if payloads[1].Player.TargetIndex() != 1 {
					t.Fatal("search target did not move to its nonzero physical target slot")
				}
				payloads[1].Player = payloads[0].Player
			}
			if !reflect.DeepEqual(payloads[0], payloads[1]) {
				t.Fatalf("aggregate payload changed:\nstandalone=%+v\ncomposed=%+v", payloads[0], payloads[1])
			}
		})
	}
}
