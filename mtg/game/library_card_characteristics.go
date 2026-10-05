package game

import "errors"

// LibraryCardCharacteristics publishes separately available printed values of
// the exact observed card. These keys never denote reveal count or action success.
type LibraryCardCharacteristics struct {
	Power     ResultKey
	Toughness ResultKey
}

func (p LibraryCardCharacteristics) keys() []ResultKey {
	var keys []ResultKey
	if p.Power != "" {
		keys = append(keys, p.Power)
	}
	if p.Toughness != "" {
		keys = append(keys, p.Toughness)
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
	if p.Power != "" && p.Power == p.Toughness {
		return errors.New("power and toughness require separate result keys")
	}
	if string(link) == string(p.Power) || string(link) == string(p.Toughness) {
		return errors.New("characteristic scalar cannot alias observed object publication")
	}
	return nil
}

func PublishedScalarKeys(primitive Primitive) []ResultKey {
	if primitive == nil {
		return nil
	}
	return primitive.instructionRefs().publishesResults
}
