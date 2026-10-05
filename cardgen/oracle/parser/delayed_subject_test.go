package parser

import "testing"

func TestDelayedSubjectOwnership(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Whenever this creature blocks or becomes blocked by a creature, destroy that creature at end of combat.",
		"Whenever this creature deals combat damage to a creature, destroy that creature at end of combat.",
		"Create two 1/1 green Insect creature tokens. You gain 2 life. Exile them at end of combat.",
	} {
		t.Run(text, func(t *testing.T) {
			document, _ := Parse(text, Context{})
			for _, sentence := range document.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					if effect.DelayedTiming == DelayedTimingNone {
						continue
					}
					immediate := effect
					immediate.Tokens, _ = cutDelayedTiming(immediate.Tokens)
					immediate.DelayedTiming = DelayedTimingNone
					t.Logf("kind=%v exact=%t optional=%t context=%v duration=%v delayed=%v ownership=%+v refs=%+v tokens=%q immediate=%q", effect.Kind, effect.Exact, effect.Optional, effect.Context, effect.Duration, effect.DelayedTiming, effect.DelayedSubject, effect.References, joinTokens(effect.Tokens), exactEffectClauseText(&immediate))
					if !effect.Exact {
						t.Fatal("modeled delayed action is not exact")
					}
				}
			}
		})
	}
}
