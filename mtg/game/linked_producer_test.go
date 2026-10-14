package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func producerValidationCard() *CardDef {
	content := Mode{Sequence: []Instruction{{Primitive: MovePermanent{
		Object: SourcePermanentReference(), Destination: zone.Exile, PublishLinkedProducer: "test/exile",
	}}}}.Ability()
	content.LinkedProducers = []LinkedProducerKey{"test/exile"}
	return &CardDef{CardFace: CardFace{Name: "Producer", Types: []types.Card{types.Artifact},
		ActivatedAbilities: []ActivatedAbility{{Content: content}},
		TriggeredAbilities: []TriggeredAbility{{
			Trigger: TriggerCondition{Pattern: TriggerPattern{Event: EventPermanentDied, Source: TriggerSourceSelf}},
			Content: Mode{Sequence: []Instruction{{Primitive: PutOnBattlefield{
				Source: OriginalSourceLinkedBattlefieldSource("test/exile"),
			}}}}.Ability(),
		}},
	}}
}

func TestLinkedProducerOwnershipValidation(t *testing.T) {
	for _, scenario := range []string{"valid", "duplicate owners", "undeclared publisher", "dangling consumer", "invalid key", "authored owner"} {
		t.Run(scenario, func(t *testing.T) {
			def := producerValidationCard()
			switch scenario {
			case "duplicate owners":
				def.ActivatedAbilities = append(def.ActivatedAbilities, def.ActivatedAbilities[0])
			case "undeclared publisher":
				def.ActivatedAbilities[0].Content.LinkedProducers = nil
				other := Mode{Sequence: []Instruction{{Primitive: GainLife{Player: ControllerReference(), Amount: Fixed(1)}}}}.Ability()
				other.LinkedProducers = []LinkedProducerKey{"test/exile"}
				def.ActivatedAbilities = append(def.ActivatedAbilities, ActivatedAbility{Content: other})
			case "dangling consumer":
				def.ActivatedAbilities = nil
			case "invalid key":
				def.ActivatedAbilities[0].Content.LinkedProducers = []LinkedProducerKey{" "}
			case "authored owner":
				def.ActivatedAbilities[0].Content.Modes[0] = Mode{Program: ProgramReference{ID: "test/exile", Version: "1"}}
			default:
			}
			valid := scenario == "valid" || scenario == "authored owner"
			if issues := ValidateCardDef(def); (len(issues) == 0) != valid {
				t.Fatalf("validation=%+v, want valid=%t", issues, valid)
			}
		})
	}
}

func TestLinkedProducerNamespacesAreExplicitAndCannotMix(t *testing.T) {
	for _, primitive := range []Primitive{
		MovePermanent{Object: SourcePermanentReference(), Destination: zone.Hand, PublishLinkedProducer: "test/exile"},
		MovePermanent{Object: SourcePermanentReference(), Destination: zone.Exile, PublishLinked: "legacy", PublishLinkedProducer: "test/exile"},
		MoveCard{Card: CardReference{Kind: CardReferenceSource}, FromZone: zone.Graveyard, Destination: zone.Exile,
			PublishLinkedProducer: "test/exile", ReplacePublishedLinked: true},
		PutOnBattlefield{Sources: []BattlefieldSource{OriginalSourceLinkedBattlefieldSource("test/exile"),
			CardBattlefieldSource(CardReference{Kind: CardReferenceSource})}},
		PutOnBattlefield{Source: OriginalSourceLinkedBattlefieldSource("test/exile"), LinkedReturnZones: []zone.Type{zone.Graveyard}},
	} {
		if err := ValidatePrimitive(primitive, nil); err == nil {
			t.Fatalf("invalid channel combination accepted: %#v", primitive)
		}
	}
	body := producerValidationCard().ActivatedAbilities[0].Content
	clone := cloneAbilityContent(body)
	clone.LinkedProducers[0] = "different"
	if body.LinkedProducers[0] != "test/exile" {
		t.Fatal("copied rules data aliased producer declarations")
	}
}
