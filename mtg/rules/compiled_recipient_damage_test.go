package rules

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledRecipientDamageFrameRuntime(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		text      string
		player    game.PlayerID
		permanent bool
		damage    int
		spell     bool
		composed  bool
	}{
		{"controller fixed", "{0}: This creature deals 2 damage to you.", game.Player2, false, 2, false, false},
		{"controller X", "{X}: This creature deals X damage to you.", game.Player2, false, 3, false, false},
		{"spell controller", "Frame Source deals 2 damage to you.", game.Player2, false, 2, true, false},
		{"event fixed", "Whenever an opponent draws a card, this creature deals 2 damage to that player.", game.Player4, false, 2, false, false},
		{"event X", "Whenever an opponent draws a card, this creature deals X damage to that player.", game.Player4, false, 3, false, false},
		{"event power", "Whenever an opponent casts a noncreature spell, this creature deals damage equal to its power to that player.", game.Player4, false, 5, false, false},
		{"event hand", "Whenever an opponent attacks you, this creature deals damage to that player equal to the number of cards in their hand.", game.Player4, false, 4, false, false},
		{"related fixed", "Whenever this creature blocks or becomes blocked by a creature, this creature deals 2 damage to that creature.", 0, true, 2, false, false},
		{"related X", "Whenever this creature blocks or becomes blocked by a creature, this creature deals X damage to that creature.", 0, true, 3, false, false},
		{"related power", "Whenever this creature blocks or becomes blocked by a creature, this creature deals damage equal to its power to that creature.", 0, true, 5, false, false},
		{"controller composition", "{0}: Draw a card. This creature deals 2 damage to you. You gain 1 life. Scry 1.", game.Player2, false, 2, false, true},
		{"event composition", "Whenever an opponent draws a card, draw a card. This creature deals 2 damage to that player. You gain 1 life. Scry 1.", game.Player4, false, 2, false, true},
		{"related composition", "Whenever this creature blocks or becomes blocked by a creature, draw a card. This creature deals 2 damage to that creature. You gain 1 life. Scry 1.", 0, true, 2, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := carddata.ScryfallCard{Name: "Frame Source", Layout: "normal", TypeLine: "Creature",
				Power: new("5"), Toughness: new("8"), OracleText: tc.text}
			if tc.spell {
				card.TypeLine = "Instant"
			}
			def := legacyCard(t, card)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, def)
			source.Controller = game.Player3
			related := addCombatCreaturePermanentWithPower(g, game.Player4, 9)
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), AffectedObjectID: source.ObjectID, Layer: game.LayerAbility,
				AddKeywords: []game.Keyword{game.Lifelink},
			})
			for range 4 {
				addCardToHand(g, game.Player4, evidenceCard("Event hand", 1))
			}
			for range 3 {
				addCardToLibrary(g, game.Player2, evidenceCard("Controller library", 1))
			}
			obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility,
				Controller: game.Player2, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID, XValue: 3}
			var content game.AbilityContent
			switch {
			case tc.spell:
				obj.Kind = game.StackSpell
				obj.SourceID = source.CardInstanceID
				content = def.SpellAbility.Val
			case len(def.ActivatedAbilities) != 0:
				content = def.ActivatedAbilities[0].Content
			default:
				obj.Kind = game.StackTriggeredAbility
				obj.HasTriggerEvent = true
				obj.TriggerEvent = game.Event{Kind: game.EventCardDrawn, Player: game.Player4}
				if tc.permanent {
					obj.TriggerEvent = game.Event{Kind: game.EventBlockerDeclared, PermanentID: source.ObjectID,
						RelatedPermanentID: related.ObjectID}
				}
				content = def.TriggeredAbilities[0].Content
			}
			engine.resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			for player, state := range g.Players {
				want := 40
				if !tc.permanent && game.PlayerID(player) == tc.player {
					want -= tc.damage
				}
				if !tc.spell && game.PlayerID(player) == game.Player3 {
					want += tc.damage
				}
				if tc.composed && game.PlayerID(player) == game.Player2 {
					want++
				}
				if state.Life != want {
					t.Fatalf("player %d life=%d want=%d", player, state.Life, want)
				}
			}
			wantMarked := 0
			if tc.permanent {
				wantMarked = tc.damage
			}
			if related.MarkedDamage != wantMarked || source.MarkedDamage != 0 {
				t.Fatalf("related/source damage=%d/%d want=%d/0", related.MarkedDamage, source.MarkedDamage, wantMarked)
			}
			if tc.composed && (g.Players[game.Player2].Hand.Size() != 1 || g.Players[game.Player2].Library.Size() != 2) {
				t.Fatal("independent draw/scry did not act on the resolving controller")
			}
			count := 0
			for _, event := range g.Events {
				if event.Kind != game.EventDamageDealt {
					continue
				}
				count++
				object, controller := source.ObjectID, game.Player3
				if tc.spell {
					object, controller = obj.ID, game.Player2
				}
				if event.Amount != tc.damage || event.SourceID != source.CardInstanceID ||
					event.SourceObjectID != object || event.Controller != controller {
					t.Fatalf("damage lost physical source/current source or resolving controller attribution: %+v", event)
				}
			}
			if count != 1 {
				t.Fatalf("damage events=%d want=1", count)
			}
		})
	}
}
