package game

import (
	"fmt"
	"slices"
)

func validateLocalProducts(instruction *Instruction) error {
	for _, key := range instruction.LocalProducts.Results {
		if key == "" || instruction.PublishResult != key && !slices.Contains(PublishedScalarKeys(instruction.Primitive), key) {
			return fmt.Errorf("local result %q is not published by this instruction", key)
		}
	}
	for _, key := range instruction.LocalProducts.Links {
		if key == "" || instruction.Primitive == nil || PublishedLinkedKey(instruction.Primitive) != key {
			return fmt.Errorf("local link %q is not published by this instruction", key)
		}
	}
	return nil
}
