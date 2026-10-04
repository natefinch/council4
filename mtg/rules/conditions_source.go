package rules

import "github.com/natefinch/council4/mtg/game/id"

// conditionSourceObjectID preserves source identity independently of source
// characteristics. Battlefield abilities retain their original ObjectID even
// when the source leaves, phases out, changes control, or returns as a new object.
func conditionSourceObjectID(ctx conditionContext) id.ID {
	if ctx.source != nil {
		return ctx.source.ObjectID
	}
	return ctx.sourceObjectID
}
