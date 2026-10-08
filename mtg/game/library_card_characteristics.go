package game

import "errors"

// LibraryCardCharacteristics publishes separately available printed values of
// the exact observed card. These keys never denote reveal count or action success.
type LibraryCardCharacteristics struct {
	Power     ResultKey
	Toughness ResultKey
	ManaValue ResultKey
}

func (p LibraryCardCharacteristics) keys() []ResultKey {
	var keys []ResultKey
	if p.Power != "" {
		keys = append(keys, p.Power)
	}
	if p.Toughness != "" {
		keys = append(keys, p.Toughness)
	}
	if p.ManaValue != "" {
		keys = append(keys, p.ManaValue)
	}
	return keys
}

func (p LibraryCardCharacteristics) validate(link LinkedKey) error {
	if len(p.keys()) == 0 {
		return nil
	}
	if link == "" {
		return errors.New("card characteristics require an exact linked observation")
	}
	return p.validateKeys(link)
}

func (p LibraryCardCharacteristics) validateKeys(link LinkedKey) error {
	seen := make(map[ResultKey]bool)
	for _, key := range p.keys() {
		if seen[key] {
			return errors.New("card characteristics require separate result keys")
		}
		seen[key] = true
		if string(link) == string(key) {
			return errors.New("characteristic scalar cannot alias observed object publication")
		}
	}
	return nil
}

func PublishedScalarKeys(primitive Primitive) []ResultKey {
	if primitive == nil {
		return nil
	}
	return primitive.instructionRefs().publishesResults
}
