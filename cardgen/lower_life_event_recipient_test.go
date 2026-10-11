package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func TestLifeEventRecipientRequiresControllerActorProof(t *testing.T) {
	for _, event := range []game.EventKind{game.EventPermanentEnteredBattlefield, game.EventPermanentDied} {
		pattern := game.TriggerPattern{Event: event, Controller: game.TriggerControllerOpponent}
		ctx := contentCtx{enclosingKind: compiler.AbilityTriggered, triggerEvent: event, triggerPattern: &pattern,
			content: compiler.AbilityContent{Effects: []compiler.CompiledEffect{{Context: parser.EffectContextReferencedPlayer}}}}
		references := []compiler.CompiledReference{{Kind: compiler.ReferenceThatPlayer, Binding: compiler.ReferenceBindingEventPermanent}}
		got, ok := lifeEventPlayerRecipient(ctx, references)
		if !ok || got.player != game.ObjectControllerReference(game.EventPermanentReference()) {
			t.Fatalf("qualified controller actor=%+v ok=%v", got, ok)
		}

		for _, tt := range []struct {
			name string
			edit func(*contentCtx)
		}{
			{"missing trigger pattern", func(c *contentCtx) { c.triggerPattern = nil }},
			{"unqualified actor", func(c *contentCtx) { c.triggerPattern.Controller = game.TriggerControllerAny }},
			{"plural event", func(c *contentCtx) { c.triggerPattern.OneOrMore = true }},
			{"nontriggered shell", func(c *contentCtx) { c.enclosingKind = compiler.AbilitySpell }},
		} {
			t.Run(tt.name, func(t *testing.T) {
				mutated := ctx
				cloned := pattern
				mutated.triggerPattern = &cloned
				tt.edit(&mutated)
				if _, ok := lifeEventPlayerRecipient(mutated, references); ok {
					t.Fatal("unsupported actor domain was accepted")
				}
			})
		}
	}
}

func TestLifeEventRecipientCardDefRolesAndUnmodeledAssignment(t *testing.T) {
	for _, tt := range []struct {
		text, kind string
	}{
		{"Whenever a creature an opponent controls dies, that player loses 2 life.", "game.PlayerReferenceObjectController"},
		{"Whenever a land an opponent controls enters, that player loses 2 life.", "game.PlayerReferenceObjectController"},
		{"Whenever an opponent becomes the monarch, that player loses 2 life.", "game.PlayerReferenceEventPlayer"},
		{"Whenever this creature becomes the target of a spell or ability an opponent controls, that player loses 4 life.", "game.PlayerReferenceEventPlayer"},
	} {
		card := &ScryfallCard{Name: "Typed Event Actor", Layout: "normal", TypeLine: "Creature",
			Power: new("3"), Toughness: new("3"), OracleText: tt.text}
		assertCardPaths(t, card, "TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.LoseLife).Player.kind = "+tt.kind)
	}
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Keeper of Tresserhorn", Layout: "normal", TypeLine: "Creature — Avatar", Power: new("6"), Toughness: new("6"),
		OracleText: "Whenever this creature attacks and isn't blocked, it assigns no combat damage this turn and defending player loses 2 life.",
	}, "unsupported life spell")
}
