package cardgen

import "testing"

// Captain Howler, Sea Scourge: a discard trigger pumps target creature, and a
// trailing "Whenever that creature deals combat damage to a player this turn,
// you draw a card." rider becomes an object-identity-bound delayed combat-damage
// trigger. The pump publishes the captured permanent under a linked key; the
// delayed trigger's combat-damage pattern binds its source to that captured
// object, so it fires only on combat damage the pumped creature deals.
func TestGenerateCaptainHowlerSeaScourge(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "3"
	card := &ScryfallCard{
		Name:      "Captain Howler, Sea Scourge",
		Layout:    "normal",
		ManaCost:  "{3}{U}{R}",
		TypeLine:  "Legendary Creature — Merfolk Pirate",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Ward—{2}, Pay 2 life.\n" +
			"Whenever you discard one or more cards, target creature gets +2/+0 until end of turn for each card discarded this way. Whenever that creature deals combat damage to a player this turn, you draw a card.",
	}
	generatedSourceContains(t, card, []string{
		"game.WardStaticAbilityWithCosts(",
		"PublishLinked:  game.LinkedKey(\"delayed-target-1\")",
		"Primitive: game.CreateDelayedTrigger{",
		"Event:                game.EventDamageDealt",
		"RequireCombatDamage:  true",
		"DamageRecipient:      game.DamageRecipientPlayer",
		"DamageSourceCaptured: true",
		"DamageSourceObject: opt.Val(game.LinkedObjectReference(\"delayed-target-1\"))",
		"Primitive: game.Draw{",
	})
}
