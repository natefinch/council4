package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceBrowbeat covers the multiplayer "may have"
// causative gate on a sorcery: every player is offered the source's damage
// ("Any player may have Browbeat deal 5 damage to them"), each accepting player
// is dealt it, and the negative consequence resolves only when nobody accepted
// ("If no one does, target player draws three cards"). The offer is an optional
// instruction over the all-players group publishing the collective decision, its
// damage recipient is the offered group member, and the consequence draw is
// gated Accepted TriFalse while keeping the ability's lone player target.
func TestGenerateExecutableCardSourceBrowbeat(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Browbeat",
		Layout:     "normal",
		ManaCost:   "{3}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Any player may have Browbeat deal 5 damage to them. If no one does, target player draws three cards.",
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Constraint: \"target player\"",
		"Primitive: game.Damage{",
		"Amount:    game.Fixed(5)",
		"Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference())",
		"Optional:           true",
		"OptionalActorGroup: opt.Val(game.AllPlayersReference())",
		"PublishResult:      game.ResultKey(\"group-may-have-action\")",
		"Primitive: game.Draw{",
		"Amount: game.Fixed(3)",
		"Player: game.TargetPlayerReference(0)",
		"Key:      \"group-may-have-action\"",
		"Accepted: game.TriFalse",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceBookBurning covers the multiplayer "may have"
// gate whose negative consequence mills a target player: every player is offered
// the source's damage ("Any player may have Book Burning deal 6 damage to them")
// and, if no one accepts, "target player mills six cards". It confirms the offer
// magnitude and group scope, and the Accepted TriFalse mill consequence keeping
// its target.
func TestGenerateExecutableCardSourceBookBurning(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Book Burning",
		Layout:     "normal",
		ManaCost:   "{3}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Any player may have Book Burning deal 6 damage to them. If no one does, target player mills six cards.",
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Constraint: \"target player\"",
		"Amount:    game.Fixed(6)",
		"Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference())",
		"OptionalActorGroup: opt.Val(game.AllPlayersReference())",
		"PublishResult:      game.ResultKey(\"group-may-have-action\")",
		"Primitive: game.Mill{",
		"Player: game.TargetPlayerReference(0)",
		"Accepted: game.TriFalse",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceBreakingPoint covers the multiplayer "may
// have" gate whose negative consequence destroys all creatures and carries a
// credited regeneration rider third sentence ("Creatures destroyed this way
// can't be regenerated."). The rider folds onto the destroy during sentence
// parsing, so the gate tolerates the third sentence, offers each player the
// source's 6 damage, and gates the PreventRegeneration mass destroy on nobody
// having accepted (Accepted TriFalse).
func TestGenerateExecutableCardSourceBreakingPoint(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Breaking Point",
		Layout:     "normal",
		ManaCost:   "{1}{R}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Any player may have Breaking Point deal 6 damage to them. If no one does, destroy all creatures. Creatures destroyed this way can't be regenerated.",
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Amount:    game.Fixed(6)",
		"Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference())",
		"OptionalActorGroup: opt.Val(game.AllPlayersReference())",
		"PublishResult:      game.ResultKey(\"group-may-have-action\")",
		"Primitive: game.Destroy{",
		"PreventRegeneration: true",
		"Key:      \"group-may-have-action\"",
		"Accepted: game.TriFalse",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceVexingDevil covers the multiplayer "may have"
// gate on an enters ability whose actor is the source pronoun and whose scope is
// opponents ("any opponent may have it deal 4 damage to them"), with an
// affirmative consequence ("If a player does, sacrifice this creature"). The
// offer is over the opponents group, its damage source is left unset so the
// runtime resolves it to the entering creature, and the sacrifice is gated
// Accepted TriTrue.
func TestGenerateExecutableCardSourceVexingDevil(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Vexing Devil",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Creature — Devil",
		OracleText: "When this creature enters, any opponent may have it deal 4 damage to them. If a player does, sacrifice this creature.",
	}, "v")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.Damage{",
		"Amount:    game.Fixed(4)",
		"Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference())",
		"OptionalActorGroup: opt.Val(game.OpponentsReference())",
		"PublishResult:      game.ResultKey(\"group-may-have-action\")",
		"Primitive: game.Sacrifice{",
		"Key:      \"group-may-have-action\"",
		"Accepted: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "DamageSource:") {
		t.Fatalf("group offer damage source should be unset:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceLonghornFirebeast covers the second opponents-
// scoped enters variant ("any opponent may have it deal 5 damage to them. If a
// player does, sacrifice this creature"), confirming the family generalizes
// across damage magnitude.
func TestGenerateExecutableCardSourceLonghornFirebeast(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Longhorn Firebeast",
		Layout:     "normal",
		ManaCost:   "{4}{R}{R}",
		TypeLine:   "Creature — Beast",
		OracleText: "When this creature enters, any opponent may have it deal 5 damage to them. If a player does, sacrifice this creature.",
	}, "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Amount:    game.Fixed(5)",
		"Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference())",
		"OptionalActorGroup: opt.Val(game.OpponentsReference())",
		"Primitive: game.Sacrifice{",
		"Accepted: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceGroupMayHaveGateFailsClosed confirms the
// multiplayer "may have" gate fails closed on shapes outside the damage-to-
// accepters family, rather than silently dropping the offer or mis-lowering the
// consequence: Distant Memories leads with a search/exile and a "put into hand"
// actor, and Sin Prodder branches with an Otherwise clause. Each must emit
// diagnostics.
func TestGenerateExecutableCardSourceGroupMayHaveGateFailsClosed(t *testing.T) {
	t.Parallel()
	for _, card := range []ScryfallCard{
		{
			Name:       "Distant Memories",
			Layout:     "normal",
			ManaCost:   "{3}{U}",
			TypeLine:   "Sorcery",
			OracleText: "Search your library for a card, exile it, then shuffle. Any opponent may have you put that card into your hand. If no player does, you draw three cards.",
		},
		{
			Name:       "Sin Prodder",
			Layout:     "normal",
			ManaCost:   "{2}{R}",
			TypeLine:   "Creature — Devil",
			OracleText: "Menace\nAt the beginning of your upkeep, reveal the top card of your library. Any opponent may have you put that card into your graveyard. If a player does, this creature deals damage to that player equal to that card's mana value. Otherwise, put that card into your hand.",
		},
	} {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&card, "x")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("%s: expected fail-closed diagnostics, got none", card.Name)
			}
		})
	}
}
