package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestLimitedTriggerDeclaredOrigins(t *testing.T) {
	native := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Native Counted Trigger", Layout: "normal", TypeLine: "Creature",
		Power: new("2"), Toughness: new("3"),
		OracleText: "Whenever this creature becomes the target of a spell or ability, you gain 1 life. This ability triggers only twice each turn.",
	})
	for _, scenario := range []string{"native", "merged component", "runtime grants", "copy origin", "static declarations", "static effects", "static bodies"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := copyCardDef(native)
			def.TriggeredAbilities = append(def.TriggeredAbilities, def.TriggeredAbilities[0])
			recipient := addCombatPermanent(g, game.Player1, def)
			switch scenario {
			case "merged component":
				card := addCardToHand(g, game.Player1, native)
				recipient.MergedCards = append(recipient.MergedCards, game.MergedCard{CardInstanceID: card})
			case "runtime grants":
				for range 2 {
					if !applyTypedContinuousEffects(g, &game.StackObject{
						SourceID: recipient.ObjectID, SourceCardID: recipient.CardInstanceID,
						Controller: game.Player1,
					}, recipient, []game.ContinuousEffect{{
						Layer: game.LayerAbility, AddAbilities: []game.Ability{&native.TriggeredAbilities[0]},
					}}, game.DurationUntilEndOfTurn) {
						t.Fatal("runtime counted grant failed")
					}
				}
			case "copy origin":
				source := addCombatPermanent(g, game.Player1, def)
				resolver := &effectResolver{engine: engine, game: g, log: &TurnLog{}, obj: &game.StackObject{
					ID: g.IDGen.Next(), SourceID: recipient.ObjectID, Controller: game.Player1,
					Targets: []game.Target{game.PermanentTarget(source.ObjectID)},
				}}
				if !handleBecomeCopy(resolver, game.BecomeCopy{Object: game.TargetPermanentReference(0)}).succeeded {
					t.Fatal("represented copy action failed")
				}
			case "static declarations", "static effects", "static bodies":
				grant := targetAnnouncementGrant(t)
				static := &grant.StaticAbilities[0]
				switch scenario {
				case "static declarations":
					grant.StaticAbilities = append(grant.StaticAbilities, *static)
				case "static effects":
					static.ContinuousEffects = append(static.ContinuousEffects, static.ContinuousEffects[0])
				case "static bodies":
					effect := &static.ContinuousEffects[0]
					effect.AddAbilities = append(effect.AddAbilities, effect.AddAbilities[0])
				default:
					t.Fatal("unknown declared-coordinate fixture")
				}
				addCombatPermanent(g, game.Player1, grant)
			case "native":
			default:
				t.Fatalf("unknown origin scenario %q", scenario)
			}
			want := 2
			switch scenario {
			case "merged component":
				want = 3
			case "runtime grants", "static declarations", "static effects", "static bodies":
				want = 4
			default:
			}
			for attempt := range 3 {
				obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility,
					Controller: game.Player2, Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)}}
				g.Stack.Push(obj)
				emitTargetEvents(g, obj)
				engine.putTriggeredAbilitiesOnStack(g)
				origins := make(map[game.TriggerOrigin]bool)
				var runtimeOrigins map[id.ID]bool
				if scenario == "runtime grants" {
					runtimeOrigins = make(map[id.ID]bool)
				}
				for _, queued := range g.Stack.Objects() {
					if queued.Kind != game.StackTriggeredAbility {
						continue
					}
					origin := queued.TriggerOrigin()
					if origin.RecipientID != recipient.ObjectID || origin.OriginID == 0 || origins[origin] {
						t.Fatal("missing or colliding declared origin")
					}
					origins[origin] = true
					if origin.Granted && runtimeOrigins != nil {
						if origin.RuntimeID == 0 {
							t.Fatal("runtime template ID used as granted identity")
						}
						runtimeOrigins[origin.RuntimeID] = true
					}
					if scenario == "copy origin" && origin.RuntimeID == 0 {
						t.Fatal("copied native declaration lost actual copy-effect origin")
					}
				}
				expected := want
				if attempt == 2 {
					expected = 0
				}
				if len(origins) != expected || attempt < 2 && runtimeOrigins != nil && len(runtimeOrigins) != 2 {
					t.Fatalf("distinct queued origins=%d, want %d; runtime grants=%d", len(origins), expected, len(runtimeOrigins))
				}
				for g.Stack.Size() > 0 {
					g.Stack.Pop()
				}
			}
		})
	}
}
