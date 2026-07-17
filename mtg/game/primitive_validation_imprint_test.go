package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestReplaceLinkedExiledCardValidation(t *testing.T) {
	valid := ReplaceLinkedExiledCard{
		Card:     CardReference{Kind: CardReferenceEvent},
		FromZone: zone.Graveyard,
		LinkID:   "imprint",
	}
	if err := valid.validatePrimitive(nil, true); err != nil {
		t.Fatalf("valid primitive rejected: %v", err)
	}
	for _, invalid := range []ReplaceLinkedExiledCard{
		{Card: valid.Card, FromZone: zone.Exile, LinkID: valid.LinkID},
		{Card: valid.Card, FromZone: zone.Graveyard},
		{FromZone: zone.Graveyard, LinkID: valid.LinkID},
	} {
		if err := invalid.validatePrimitive(nil, true); err == nil {
			t.Fatalf("invalid primitive accepted: %#v", invalid)
		}
	}
}

func TestValidateCardDefRequiresLinkedExileTokenCopyLink(t *testing.T) {
	makeCard := func(link LinkedKey) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:  "Imprint Copy",
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []TriggeredAbility{{
				Trigger: TriggerCondition{Pattern: TriggerPattern{Event: EventPermanentDied}},
				Content: Mode{Sequence: []Instruction{{
					Primitive: ReplaceLinkedExiledCard{
						Card:     CardReference{Kind: CardReferenceEvent},
						FromZone: zone.Graveyard,
						LinkID:   "imprint",
					},
				}}}.Ability(),
			}},
			ActivatedAbilities: []ActivatedAbility{{
				Content: Mode{Sequence: []Instruction{{
					Primitive: CreateToken{
						Amount: Fixed(1),
						Source: TokenCopyOf(TokenCopySpec{
							Source: TokenCopySourceLinkedExiledCard,
							LinkID: link,
						}),
					},
				}}}.Ability(),
			}},
		}}
	}
	if issues := ValidateCardDef(makeCard("imprint")); len(issues) != 0 {
		t.Fatalf("valid linked-exile token copy issues = %#v", issues)
	}
	if issues := ValidateCardDef(makeCard("")); len(issues) == 0 {
		t.Fatal("linked-exile token copy without a link was accepted")
	}
}
