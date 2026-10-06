package game

import "fmt"

func validateOptionalDecisions(sequence []Instruction) error {
	published := make(map[OptionalDecisionKey]int)
	for i, instruction := range sequence {
		if instruction.OptionalDecisionGate != "" {
			if _, exists := published[instruction.OptionalDecisionGate]; !exists {
				return fmt.Errorf("instruction[%d]: OptionalDecisionGate references key %q not yet published", i, instruction.OptionalDecisionGate)
			}
		}
		if key := instruction.PublishOptionalDecision; key != "" {
			if !instruction.Optional {
				return fmt.Errorf("instruction[%d]: PublishOptionalDecision requires Optional", i)
			}
			if instruction.OptionalActorGroup.Exists || instruction.TemptingOffer || instruction.ForEachPlayerGroup.Exists {
				return fmt.Errorf("instruction[%d]: optional decision publication requires one deciding player", i)
			}
			if previous, duplicate := published[key]; duplicate {
				return fmt.Errorf("instruction[%d]: duplicate optional decision key %q (first used at index %d)", i, key, previous)
			}
			published[key] = i
		}
		for _, body := range instruction.TemptingOfferBody {
			if body.PublishOptionalDecision != "" || body.OptionalDecisionGate != "" {
				return fmt.Errorf("instruction[%d]: optional decision inside TemptingOfferBody not supported", i)
			}
		}
	}
	return nil
}
