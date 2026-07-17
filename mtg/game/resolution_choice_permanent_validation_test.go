package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestPermanentChoiceAndChoiceNamedSearchValidate(t *testing.T) {
	t.Parallel()
	player := ControllerReference()
	choose := Choose{
		Choice: ResolutionChoice{
			Kind:            ResolutionChoicePermanent,
			PlayerReference: &player,
			Selection:       &Selection{RequiredTypes: []types.Card{types.Land}, Controller: ControllerYou},
		},
		PublishChoice: ResolutionChosenPermanentChoiceKey,
	}
	if err := choose.validatePrimitive(nil, true); err != nil {
		t.Fatalf("permanent choice validation failed: %v", err)
	}
	search := Search{
		Player: ControllerReference(),
		Spec: SearchSpec{
			SourceZone:     zone.Library,
			Destination:    zone.Battlefield,
			Filter:         Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
			NameFromChoice: ResolutionChosenPermanentChoiceKey,
			EntersTapped:   true,
		},
		Amount: Fixed(2),
	}
	if err := search.validatePrimitive(nil, true); err != nil {
		t.Fatalf("choice-named search validation failed: %v", err)
	}
	search.Spec.Name = "Forest"
	if err := search.validatePrimitive(nil, true); err == nil {
		t.Fatal("search combining fixed and choice-derived names validated")
	}
}

func TestPermanentChoiceRequiresPlayerAndSelection(t *testing.T) {
	t.Parallel()
	if err := (Choose{Choice: ResolutionChoice{Kind: ResolutionChoicePermanent}}).validatePrimitive(nil, true); err == nil {
		t.Fatal("permanent choice without player or selection validated")
	}
}
