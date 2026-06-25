package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const undercityInformerOracle = "{1}, Sacrifice a creature: Target player reveals cards from the top of their library until they reveal a land card, then puts those cards into their graveyard."

func TestGenerateExecutableCardSourceUndercityInformer(t *testing.T) {
	t.Parallel()
	power, toughness := "2", "2"
	card := &ScryfallCard{
		Name:       "Undercity Informer",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Creature — Human Rogue",
		OracleText: undercityInformerOracle,
		Power:      &power,
		Toughness:  &toughness,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.RevealUntil{",
		"game.TargetPlayerReference(0)",
		"Destination: zone.Graveyard",
		"types.Land",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}

	face := lowerSingleFace(t, card)
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	primitive, ok := sequence[0].Primitive.(game.RevealUntil)
	if !ok {
		t.Fatalf("primitive = %#v, want game.RevealUntil", sequence[0].Primitive)
	}
	if primitive.Destination != zone.Graveyard {
		t.Fatalf("destination = %v, want Graveyard", primitive.Destination)
	}
	if primitive.Player.Kind() != game.PlayerReferenceTargetPlayer {
		t.Fatalf("player = %#v, want target player", primitive.Player)
	}
	if len(primitive.Until.RequiredTypes) != 1 || primitive.Until.RequiredTypes[0] != types.Land {
		t.Fatalf("until = %#v, want land", primitive.Until)
	}
}

const treasureHuntOracle = "Reveal cards from the top of your library until you reveal a land card. Put all cards revealed this way into your hand."

func TestGenerateExecutableCardSourceTreasureHunt(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Treasure Hunt",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Sorcery",
		OracleText: treasureHuntOracle,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.RevealUntil{",
		"game.ControllerReference()",
		"Destination: zone.Hand",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}

	face := lowerSingleFace(t, card)
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	primitive, ok := sequence[0].Primitive.(game.RevealUntil)
	if !ok {
		t.Fatalf("primitive = %#v, want game.RevealUntil", sequence[0].Primitive)
	}
	if primitive.Destination != zone.Hand {
		t.Fatalf("destination = %v, want Hand", primitive.Destination)
	}
	if primitive.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("player = %#v, want controller", primitive.Player)
	}
}
