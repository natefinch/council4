package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerYouCreateControllerSubject verifies that a controller-form token
// creation written with the "you create" subject ("..., you create a Treasure
// token.") lowers identically to the bare imperative "Create" wording. The two
// surface wordings describe the same controller effect, so the token is created
// for the controller with no recipient group. It backs Monologue Tax, Wizened
// Mentor, and the other "you create" cards.
func TestLowerYouCreateControllerSubject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
		token  string
	}{
		{
			name:   "named artifact",
			oracle: "You create a Treasure token.",
			token:  "Treasure",
		},
		{
			name:   "creature",
			oracle: "You create a 1/1 white Soldier creature token.",
			token:  "Soldier",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "You Create " + test.name,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
			create := createTokenPrimitive(t, face)
			if create.Recipient.Exists {
				t.Fatalf("recipient = %+v, want none (controller)", create.Recipient)
			}
			if create.RecipientGroup.Kind != game.PlayerGroupReferenceNone {
				t.Fatalf("recipient group = %+v, want none", create.RecipientGroup)
			}
			def, ok := create.Source.TokenDefRef()
			if !ok {
				t.Fatal("token source is not a token definition")
			}
			if def.Name != test.token {
				t.Fatalf("token name = %q, want %q", def.Name, test.token)
			}
		})
	}
}

// TestLowerEachPlayerRecipientGroup verifies that the player-group recipient
// forms ("Each player creates ...", "Each opponent creates ...") lower to a
// CreateToken whose RecipientGroup widens the recipient to every member of the
// group. It backs Grismold, Marching Duodrone, Edge Rover, Elephant-Mandrill
// (each player) and Slaughter Specialist (each opponent).
func TestLowerEachPlayerRecipientGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
		token  string
		want   game.PlayerGroupReferenceKind
	}{
		{
			name:   "each player creature",
			oracle: "Each player creates a 1/1 green Plant creature token.",
			token:  "Plant",
			want:   game.PlayerGroupReferenceAllPlayers,
		},
		{
			name:   "each player named",
			oracle: "Each player creates a Treasure token.",
			token:  "Treasure",
			want:   game.PlayerGroupReferenceAllPlayers,
		},
		{
			name:   "each opponent creature",
			oracle: "Each opponent creates a 1/1 white Human creature token.",
			token:  "Human",
			want:   game.PlayerGroupReferenceOpponents,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Group " + test.name,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
			create := createTokenPrimitive(t, face)
			if create.Recipient.Exists {
				t.Fatalf("recipient = %+v, want none", create.Recipient)
			}
			if create.RecipientGroup.Kind != test.want {
				t.Fatalf("recipient group kind = %d, want %d", create.RecipientGroup.Kind, test.want)
			}
			def, ok := create.Source.TokenDefRef()
			if !ok {
				t.Fatal("token source is not a token definition")
			}
			if def.Name != test.token {
				t.Fatalf("token name = %q, want %q", def.Name, test.token)
			}
		})
	}
}

// TestLowerEachOtherPlayerTokenUnsupported confirms the player-group widening
// stays closed for shapes it cannot represent: "each other player" has no
// PlayerGroupReference, and a "each player other than target player" form
// carries a target the simple group recipient does not model. Both must fail
// closed.
func TestLowerEachOtherPlayerTokenUnsupported(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "each other player",
			oracle: "Each other player creates a Treasure token.",
		},
		{
			name:   "each player other than target",
			oracle: "Each player other than target player creates a 1/1 red Dragon creature token with flying.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Unsupported " + test.name,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracle,
			})
		})
	}
}
