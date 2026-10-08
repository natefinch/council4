package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledResolvingSpellRiderDamageAttribution(t *testing.T) {
	for _, card := range []cardgen.ScryfallCard{
		{Name: "Burn the Accursed", Layout: "normal", TypeLine: "Instant", OracleText: "Burn the Accursed deals 5 damage to target creature and 2 damage to that creature's controller. If that creature would die this turn, exile it instead."},
		{Name: "Lava Coil", Layout: "normal", TypeLine: "Sorcery", OracleText: "Lava Coil deals 4 damage to target creature. If that creature would die this turn, exile it instead."},
	} {
		t.Run(card.Name, func(t *testing.T) {
			def := compileUnlessCard(t, card)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCardToHand(g, game.Player1, def)
			target := addCombatCreaturePermanentWithPower(g, game.Player3, 9)
			decoy := addCombatPermanent(g, game.Player4, vanillaCreature(card.Name, 1, 1))
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), AffectedObjectID: decoy.ObjectID, Layer: game.LayerAbility,
				AddKeywords: []game.Keyword{game.Lifelink, game.Deathtouch},
			})
			obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player2,
				SourceID: source, SourceCardID: source, Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			wantDamage := 4
			wantLife := 40
			if card.Name == "Burn the Accursed" {
				wantDamage, wantLife = 5, 38
			}
			if target.MarkedDamage != wantDamage || g.Players[game.Player3].Life != wantLife ||
				g.Players[game.Player2].Life != 40 || g.Players[game.Player4].Life != 40 {
				t.Fatal("spell/rider damage was not actually dealt")
			}
			count := 0
			for _, event := range g.Events {
				if event.Kind != game.EventDamageDealt {
					continue
				}
				count++
				if event.SourceID != source || event.SourceObjectID != obj.ID || event.Controller != game.Player2 {
					t.Fatalf("damage attribution lost resolving spell identity/controller: %+v", event)
				}
			}
			wantEvents := 1
			if card.Name == "Burn the Accursed" {
				wantEvents = 2
			}
			if count != wantEvents {
				t.Fatalf("damage events=%d want=%d", count, wantEvents)
			}
		})
	}
}

func TestCompiledDestroyedTargetPowerAndControllerLKI(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Agonizing Demise", Layout: "normal", TypeLine: "Instant",
		OracleText: "Kicker {1}{R}\nDestroy target nonblack creature. It can't be regenerated. If this spell was kicked, Agonizing Demise deals damage equal to that creature's power to the creature's controller."})
	for _, outcome := range []string{"destroyed", "indestructible", "exile replacement", "modified power", "changed controller", "not kicked", "missing target", "stale incarnation"} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCardToHand(g, game.Player1, def)
			target := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
			target.Controller = game.Player3
			decoy := addCombatCreaturePermanentWithPower(g, game.Player3, 19)
			obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player2,
				SourceID: source, SourceCardID: source, KickerPaid: outcome != "not kicked",
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			switch outcome {
			case "modified power":
				target.TemporaryPowerModifier = 4
			case "changed controller":
				target.Controller = game.Player4
			case "indestructible":
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.Indestructible},
				})
			case "exile replacement":
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
				})
			case "missing target":
				obj.Targets = nil
			case "stale incarnation":
				if !movePermanentToZone(g, target, zone.Graveyard) {
					t.Fatal("fixture departure failed")
				}
				returnObj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{currentCardTarget(t, g, target.CardInstanceID)}}
				newEffectResolver(NewEngine(nil), g, returnObj, [game.NumPlayers]PlayerAgent{}, &TurnLog{}).
					resolveInstruction(&game.Instruction{Primitive: game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget})}})
			case "destroyed", "not kicked":
			default:
				t.Fatal("unknown outcome")
			}
			engine := NewEngine(nil)
			if outcome == "stale incarnation" {
				g.Stack.Push(obj)
				engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			} else {
				engine.resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			}
			want := 40
			if outcome != "not kicked" && outcome != "missing target" && outcome != "stale incarnation" {
				want = 33
			}
			if outcome == "modified power" {
				want = 29
			}
			recipient := game.Player3
			if outcome == "changed controller" {
				recipient = game.Player4
			}
			if g.Players[recipient].Life != want || g.Players[game.Player1].Life != 40 || g.Players[game.Player2].Life != 40 ||
				decoy.MarkedDamage != 0 {
				t.Fatalf("target-power/controller LKI isolation failed: life=%d want%d", g.Players[recipient].Life, want)
			}
			events := 0
			for _, event := range g.Events {
				if event.Kind == game.EventDamageDealt {
					events++
					if event.SourceID != source || event.SourceObjectID != obj.ID || event.Controller != game.Player2 ||
						event.Player != recipient || event.Amount != 40-want {
						t.Fatal("target observation leaked into spell attribution or recipient")
					}
				}
			}
			if events != 0 && want == 40 || events != 1 && want != 40 {
				t.Fatalf("damage events=%d expected life=%d", events, want)
			}
		})
	}
}
