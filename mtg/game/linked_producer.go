package game

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
)

// LinkedProducerKey is a stable, implementation-qualified producer identity.
// Each key is owned by exactly one declared ability body on a Card Definition.
// Repeated resolutions and copies of that body accumulate in the same channel
// on the original permanent occurrence; reentry starts a different channel.
type LinkedProducerKey string

func (key LinkedProducerKey) valid() bool {
	return key != "" && !strings.ContainsAny(string(key), " \t\r\n")
}

// PublishedLinkedProducer reports the opt-in original-source publisher.
func PublishedLinkedProducer(primitive Primitive) LinkedProducerKey {
	if publisher, ok := primitive.(interface{ LinkedProducerPublication() LinkedProducerKey }); ok {
		return publisher.LinkedProducerPublication()
	}
	return ""
}

// LinkedProducerPublication reports the declared exile channel.
func (p MovePermanent) LinkedProducerPublication() LinkedProducerKey { return p.PublishLinkedProducer }

// LinkedProducerPublication reports the declared exile channel.
func (p MoveCard) LinkedProducerPublication() LinkedProducerKey { return p.PublishLinkedProducer }

// ValidateLinkedProducerOperation enforces ownership for authored operations and
// declarative instructions alike. A key on another body grants no publication rights.
func ValidateLinkedProducerOperation(primitive Primitive, declared []LinkedProducerKey) error {
	if producer := PublishedLinkedProducer(primitive); producer != "" {
		if !producer.valid() || !slices.Contains(declared, producer) {
			return fmt.Errorf("linked producer %q is not declared by this ability body", producer)
		}
	}
	return nil
}

func validateProducerPublication(producer LinkedProducerKey, legacy LinkedKey, destination zone.Type) error {
	if producer == "" {
		return nil
	}
	if !producer.valid() || legacy != "" || destination != zone.Exile {
		return errors.New("original-source producer publication requires a valid producer, no legacy link and an exile destination")
	}
	return nil
}

func consumedLinkedProducers(primitive Primitive) []LinkedProducerKey {
	if consumer, ok := primitive.(interface{ LinkedProducerConsumption() []LinkedProducerKey }); ok {
		return consumer.LinkedProducerConsumption()
	}
	return nil
}

// LinkedProducerConsumption reports the selected original-source channels.
func (p PutOnBattlefield) LinkedProducerConsumption() []LinkedProducerKey {
	sources := p.Sources
	if p.Source.Valid() {
		sources = []BattlefieldSource{p.Source}
	}
	var producers []LinkedProducerKey
	for _, source := range sources {
		if producer, ok := source.LinkedProducer(); ok {
			producers = append(producers, producer)
		}
	}
	return producers
}

// NestedAbilityContents exposes resolving bodies without knowing concrete primitives.
func (p RepeatProcess) NestedAbilityContents() []AbilityContent { return []AbilityContent{p.Body} }

// NestedAbilityContents exposes the scheduled body.
func (p CreateDelayedTrigger) NestedAbilityContents() []AbilityContent {
	return []AbilityContent{p.Trigger.Content}
}

// NestedAbilityContents exposes the reflexive body.
func (p CreateReflexiveTrigger) NestedAbilityContents() []AbilityContent {
	return []AbilityContent{p.Trigger.Content}
}

func walkLinkedProducerContent(content AbilityContent, visit func(AbilityContent)) {
	visit(content)
	var sequence func([]Instruction)
	sequence = func(instructions []Instruction) {
		for _, instruction := range instructions {
			sequence(instruction.TemptingOfferBody)
			if nested, ok := instruction.Primitive.(interface{ NestedAbilityContents() []AbilityContent }); ok {
				for _, body := range nested.NestedAbilityContents() {
					walkLinkedProducerContent(body, visit)
				}
			}
		}
	}
	for _, mode := range content.Modes {
		sequence(mode.Sequence)
	}
}

func (v *cardDefValidator) validateDeclaredLinkedProducers() {
	var contents []AbilityContent
	for _, face := range []*CardFace{&v.card.CardFace, faceIfPresent(v.card.Back.Exists, &v.card.Back.Val), faceIfPresent(v.card.Alternate.Exists, &v.card.Alternate.Val)} {
		if face == nil {
			continue
		}
		for index := range face.AbilityCount() {
			walkLinkedProducerContent(BodyContent(face.BodyAt(index)), func(content AbilityContent) {
				contents = append(contents, content)
			})
		}
	}
	owners := make(map[LinkedProducerKey]bool)
	for _, content := range contents {
		for _, producer := range content.LinkedProducers {
			if !producer.valid() || owners[producer] {
				v.add(v.card.Name, "LinkedProducers", CardDefIssueInvalidAbilityBody,
					fmt.Sprintf("linked producer %q must have one valid declared owner", producer))
			}
			owners[producer] = true
		}
	}
	for _, content := range contents {
		var sequence func([]Instruction)
		sequence = func(instructions []Instruction) {
			for _, instruction := range instructions {
				if err := ValidateLinkedProducerOperation(instruction.Primitive, content.LinkedProducers); err != nil {
					v.add(v.card.Name, "LinkedProducers", CardDefIssueInvalidAbilityBody, err.Error())
				}
				for _, producer := range consumedLinkedProducers(instruction.Primitive) {
					if !owners[producer] {
						v.add(v.card.Name, "LinkedProducers", CardDefIssueInvalidAbilityBody,
							fmt.Sprintf("linked producer %q has no declared owner", producer))
					}
				}
				sequence(instruction.TemptingOfferBody)
			}
		}
		for _, mode := range content.Modes {
			sequence(mode.Sequence)
		}
	}
}
