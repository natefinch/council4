package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLibraryGraveyardDamageSourceLowering(t *testing.T) {
	for _, tc := range []struct {
		name     string
		text     string
		explicit bool
	}{
		{"source card", "When Event Source is put into your graveyard from your library, you may exile it. If you do, Event Source deals 3 damage to each opponent and you gain 3 life.", false},
		{"watcher", "Whenever a card is put into your graveyard from your library, Event Source deals 3 damage to each opponent.", true},
		{"self death", "When Event Source dies, Event Source deals 3 damage to each opponent.", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Event Source", Layout: "normal", TypeLine: "Creature", OracleText: tc.text}
			face := lowerSingleFace(t, card)
			found := 0
			for _, instruction := range face.TriggeredAbilities[0].Content.Modes[0].Sequence {
				if damage, ok := instruction.Primitive.(game.Damage); ok {
					found++
					if damage.DamageSource.Exists != tc.explicit ||
						tc.explicit && damage.DamageSource.Val != game.SourcePermanentReference() {
						t.Fatalf("source attribution=%+v want original permanent=%v", damage.DamageSource, tc.explicit)
					}
				}
				if move, ok := instruction.Primitive.(game.MoveCard); ok &&
					move.Card.Kind != game.CardReferenceEvent {
					t.Fatal("owned exile lost its strict event-card incarnation")
				}
			}
			if found != 1 {
				t.Fatalf("damage instructions=%d", found)
			}
		})
	}
}
