package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
)

func TestLibraryOwnerDestinationRequiresMatchingDomains(t *testing.T) {
	for _, mutate := range []string{"none", "different producer", "different index", "wrong player binding", "wrong card kind", "missing owner", "unclaimed owner"} {
		t.Run(mutate, func(t *testing.T) {
			effect := compiler.CompiledEffect{LibraryOwnerDestination: true}
			references := []compiler.CompiledReference{
				{Kind: compiler.ReferenceThatObject, Binding: compiler.ReferenceBindingPriorInstructionResult,
					ProducerClauseID: 9, PriorInstruction: 2},
				{Kind: compiler.ReferenceThatPlayer, Binding: compiler.ReferenceBindingLibraryOwner,
					ProducerClauseID: 9, PriorInstruction: 2},
			}
			switch mutate {
			case "none":
			case "different producer":
				references[1].ProducerClauseID = 8
			case "different index":
				references[1].PriorInstruction = 1
			case "wrong player binding":
				references[1].Binding = compiler.ReferenceBindingTarget
			case "wrong card kind":
				references[0].Kind = compiler.ReferenceThatPlayer
			case "missing owner":
				references = references[:1]
			case "unclaimed owner":
				effect.LibraryOwnerDestination = false
			default:
				t.Fatalf("unknown mutation %q", mutate)
			}
			_, ok := libraryCardActionSubject(effect, references)
			if ok != (mutate == "none") {
				t.Fatalf("subject proof=%v for %s", ok, mutate)
			}
		})
	}
}

func TestLibraryOwnerDestinationRefusesAmbiguousOrUnmodeledOwner(t *testing.T) {
	for _, text := range []string{
		"Look at the top card of your library. If it's a land card, put it into that player's hand.",
		"Look at the top two cards of target player's library. Put that card into that player's graveyard.",
		"Look at the top card of target player's library. Target opponent draws a card. Put it into that player's graveyard.",
		"Look at the top card of target player's library. Put it into another player's graveyard.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unproven Library Owner", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
			})
		})
	}
}
