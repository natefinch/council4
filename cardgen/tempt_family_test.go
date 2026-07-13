package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

func temptWithGloryCard() *ScryfallCard {
	return &ScryfallCard{
		Name:          "Tempt with Glory",
		Layout:        "normal",
		ManaCost:      "{5}{W}",
		TypeLine:      "Sorcery",
		Colors:        []string{"W"},
		ColorIdentity: []string{"W"},
		SetType:       "commander",
		Games:         []string{"paper"},
		OracleText:    "Tempting offer \u2014 Put a +1/+1 counter on each creature you control. Each opponent may put a +1/+1 counter on each creature they control. For each opponent who does, put a +1/+1 counter on each creature you control.",
	}
}

func temptWithImmortalityCard() *ScryfallCard {
	return &ScryfallCard{
		Name:          "Tempt with Immortality",
		Layout:        "normal",
		ManaCost:      "{4}{B}",
		TypeLine:      "Sorcery",
		Colors:        []string{"B"},
		ColorIdentity: []string{"B"},
		SetType:       "commander",
		Games:         []string{"paper", "mtgo"},
		OracleText:    "Tempting offer \u2014 Return a creature card from your graveyard to the battlefield. Each opponent may return a creature card from their graveyard to the battlefield. For each opponent who does, return a creature card from your graveyard to the battlefield.",
	}
}

func temptWithReflectionsCard() *ScryfallCard {
	return &ScryfallCard{
		Name:          "Tempt with Reflections",
		Layout:        "normal",
		ManaCost:      "{3}{U}",
		TypeLine:      "Sorcery",
		Colors:        []string{"U"},
		ColorIdentity: []string{"U"},
		SetType:       "commander",
		Games:         []string{"paper", "mtgo"},
		OracleText:    "Tempting offer \u2014 Choose target creature you control. Create a token that's a copy of that creature. Each opponent may create a token that's a copy of that creature. For each opponent who does, create a token that's a copy of that creature.",
	}
}

// temptingOfferInstruction fetches the sole spell-ability instruction of a landed
// Tempt-cycle card, asserting the ability lowered to exactly one instruction
// flagged TemptingOffer, optional, and offered to the opponents.
func temptingOfferInstruction(t *testing.T, card *ScryfallCard) game.Instruction {
	t.Helper()
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists {
		t.Fatalf("%s produced no spell ability", card.Name)
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("%s spell modes = %#v, want one single-instruction mode", card.Name, modes)
	}
	instr := modes[0].Sequence[0]
	if !instr.TemptingOffer {
		t.Fatalf("%s instruction is not flagged TemptingOffer", card.Name)
	}
	if !instr.Optional {
		t.Fatalf("%s TemptingOffer instruction is not optional", card.Name)
	}
	if !instr.OptionalActorGroup.Exists ||
		instr.OptionalActorGroup.Val.Kind != game.PlayerGroupReferenceOpponents {
		t.Fatalf("%s OptionalActorGroup = %#v, want opponents", card.Name, instr.OptionalActorGroup)
	}
	return instr
}

// TestLowerTemptWithGloryTemptingOffer proves the +1/+1 counter idiom lowers to a
// single group offer whose AddCounter targets every creature the acting player
// controls through GroupOfferMemberReference().
func TestLowerTemptWithGloryTemptingOffer(t *testing.T) {
	t.Parallel()
	instr := temptingOfferInstruction(t, temptWithGloryCard())
	add, ok := instr.Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %#v, want game.AddCounter", instr.Primitive)
	}
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", add.CounterKind)
	}
	if add.Amount.IsDynamic() || add.Amount.Value() != 1 {
		t.Fatalf("amount = %#v, want fixed 1", add.Amount)
	}
	if add.Group.Domain() != game.GroupDomainPlayerControlled {
		t.Fatalf("group domain = %v, want player-controlled", add.Group.Domain())
	}
	player, ok := add.Group.PlayerAnchor()
	if !ok || player.Kind() != game.PlayerReferenceGroupOfferMember {
		t.Fatalf("group player anchor = %#v, want GroupOfferMember", player)
	}
}

// TestLowerTemptWithImmortalityTemptingOffer proves the reanimation idiom lowers
// to a single group offer whose graveyard-return choice reads the acting player's
// own graveyard through GroupOfferMemberReference().
func TestLowerTemptWithImmortalityTemptingOffer(t *testing.T) {
	t.Parallel()
	instr := temptingOfferInstruction(t, temptWithImmortalityCard())
	env, ok := instr.Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %#v, want game.ChooseFromZone", instr.Primitive)
	}
	if env.SourceZone != zone.Graveyard {
		t.Fatalf("source zone = %v, want graveyard", env.SourceZone)
	}
	if env.Destination.Zone != zone.Battlefield {
		t.Fatalf("destination zone = %v, want battlefield", env.Destination.Zone)
	}
	if env.Player.Kind() != game.PlayerReferenceGroupOfferMember {
		t.Fatalf("player = %#v, want GroupOfferMember", env.Player)
	}
}

// TestLowerTemptWithReflectionsTemptingOffer proves the copy idiom lowers to a
// single group offer whose copy token copies the controller's one target creature
// and enters under the acting player's control through GroupOfferMemberReference().
func TestLowerTemptWithReflectionsTemptingOffer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, temptWithReflectionsCard())
	if !face.SpellAbility.Exists {
		t.Fatal("Tempt with Reflections produced no spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want exactly one", mode.Targets)
	}
	instr := mode.Sequence[0]
	if !instr.TemptingOffer {
		t.Fatal("instruction is not flagged TemptingOffer")
	}
	token, ok := instr.Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %#v, want game.CreateToken", instr.Primitive)
	}
	if !token.Recipient.Exists ||
		token.Recipient.Val.Kind() != game.PlayerReferenceGroupOfferMember {
		t.Fatalf("recipient = %#v, want GroupOfferMember", token.Recipient)
	}
	spec, ok := token.Source.TokenCopy()
	if !ok {
		t.Fatalf("token source = %#v, want a copy spec", token.Source)
	}
	if spec.Source != game.TokenCopySourceObject {
		t.Fatalf("copy source = %v, want object copy", spec.Source)
	}
	if spec.Object != game.TargetPermanentReference(0) {
		t.Fatalf("copy object = %#v, want target permanent 0", spec.Object)
	}
}
