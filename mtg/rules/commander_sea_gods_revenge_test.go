package rules

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Unregistered Engine proving fixture, not a whole-card production admission.
// Source: testdata/card-authoring/commander-choice-apnap.json.
func seaGodsRevengeCommanderFixture(t *testing.T) (*Engine, *game.CardDef) {
	t.Helper()
	reference := game.ProgramReference{ID: "fixture/1906/sea-gods-revenge", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference,
		Resolve: func(ctx cardprogram.Context) error {
			if _, err := ctx.Execute(game.MovePermanent{
				Object: game.AllTargetPermanentsReference(0), Destination: zone.Hand,
			}); err != nil {
				return err
			}
			_, err := ctx.Execute(game.Scry{Amount: game.Fixed(1), Player: game.ControllerReference()})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	def := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sea God's Revenge", Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Blue},
			ManaCost:   opt.Val(cost.Mana{cost.O(5), cost.U}),
			OracleText: "Return up to three target creatures your opponents control to their owners' hands. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)",
			SpellAbility: opt.Val(game.Mode{
				Program: reference,
				Targets: []game.TargetSpec{{
					MinTargets: 0, MaxTargets: 3, Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent,
					}),
				}},
			}.Ability()),
		},
	}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("invalid unregistered fixture: %v", issues)
	}
	return NewEngineWithPrograms(nil, catalog), def
}

func TestSeaGodsRevengeCommanderChoicesBeforeSimultaneousReturnAndScry(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, changedControl := range []bool{false, true} {
			for mask := range 4 {
				t.Run(fmt.Sprintf("reverse-%t/control-%t/accept-%d", reverse, changedControl, mask), func(t *testing.T) {
					engine, def := seaGodsRevengeCommanderFixture(t)
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					second := addCommanderPermanent(g, game.Player2)
					fourth := addCommanderPermanent(g, game.Player4)
					ordinary := addCombatPermanent(g, game.Player3, vanillaCreature("Ordinary return", 2, 2))
					if changedControl {
						second.Controller, fourth.Controller = game.Player4, game.Player2
					}
					permanents := []*game.Permanent{second, ordinary, fourth}
					if reverse {
						slices.Reverse(permanents)
					}
					var targets []game.Target
					for _, permanent := range permanents {
						targets = append(targets, game.PermanentTarget(permanent.ObjectID))
					}
					bottom := addCardToLibrary(g, game.Player1, vanillaCreature("Private scry bottom", 1, 1))
					top := addCardToLibrary(g, game.Player1, vanillaCreature("Private scry top", 1, 1))
					spell := addCardToHand(g, game.Player1, def)
					setMainPhasePriority(g, game.Player1)
					g.Players[game.Player1].ManaPool.Add(mana.C, 5)
					g.Players[game.Player1].ManaPool.Add(mana.U, 1)
					if !engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 0, nil)) {
						t.Fatal("paid three-target Sea God's Revenge fixture cast failed")
					}
					replay := g.Clone()
					branch := g.Clone()
					eventCount := len(g.Events)
					resolve := func(state *game.Game, decisions int) *TurnLog {
						t.Helper()
						wantZone := func(owner game.PlayerID) zone.Type {
							bit := 0
							if owner == game.Player4 {
								bit = 1
							}
							if decisions&(1<<bit) != 0 {
								return zone.Command
							}
							return zone.Hand
						}
						var choices []game.PlayerID
						scried := false
						agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
							switch request.Kind {
							case game.ChoiceMay:
								if len(state.Battlefield) != 3 || len(state.Events) != eventCount {
									t.Fatal("a return committed before all commander choices")
								}
								for _, permanent := range permanents {
									if current, ok := permanentByObjectID(state, permanent.ObjectID); !ok ||
										current.Controller != permanent.Controller {
										t.Fatal("commander choices did not share the original battlefield")
									}
								}
								choices = append(choices, request.Player)
								if wantZone(request.Player) == zone.Command {
									return []int{1}
								}
								return []int{0}
							case game.ChoiceScry:
								if !slices.Equal(choices, []game.PlayerID{game.Player2, game.Player4}) ||
									len(state.Battlefield) != 0 || request.Player != game.Player1 ||
									!request.Private || !request.Subject.Exists || request.Subject.Val.CardID != top {
									t.Fatal("scry preceded the APNAP choices/returns or lost its private top card")
								}
								for _, permanent := range []*game.Permanent{second, fourth} {
									cards, _ := destinationZone(state, permanent.Owner, wantZone(permanent.Owner))
									if !cards.Contains(permanent.CardInstanceID) {
										t.Fatal("scry preceded a commander's final destination")
									}
								}
								if !state.Players[game.Player3].Hand.Contains(ordinary.CardInstanceID) {
									t.Fatal("ordinary return did not precede scry")
								}
								scried = true
								return []int{1}
							default:
								t.Fatalf("unexpected fixture choice: %+v", request)
								return nil
							}
						}}
						log := &TurnLog{}
						engine.resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
						if !scried || len(log.Choices) != 3 || !state.Players[game.Player1].Graveyard.Contains(spell) ||
							!slices.Equal(state.Players[game.Player1].Library.All(), []id.ID{bottom, top}) {
							t.Fatal("fixture did not finish with two owner decisions, scry, and normal spell cleanup")
						}
						var batch id.ID
						var departures []game.Event
						for _, event := range state.Events[eventCount:] {
							if event.Kind == game.EventZoneChanged && event.FromZone == zone.Battlefield {
								if batch == 0 {
									batch = event.SimultaneousID
								}
								if batch == 0 || event.SimultaneousID != batch {
									t.Fatal("returned creatures lost their simultaneous event identity")
								}
								departures = append(departures, event)
							}
						}
						if len(departures) != 3 {
							t.Fatal("fixture did not return all three legal targeted creatures")
						}
						for i, event := range departures {
							if event.PermanentID != permanents[i].ObjectID || event.Player != permanents[i].Owner {
								t.Fatal("APNAP decision order changed the target-to-event correspondence")
							}
						}
						return log
					}
					log := resolve(g, mask)
					replayLog := resolve(replay, mask)
					if !reflect.DeepEqual(log, replayLog) || !reflect.DeepEqual(g.Events, replay.Events) {
						t.Fatal("Sea God's Revenge fixture replay changed decisions or events")
					}
					resolve(branch, mask^3)
					for player := range game.NumPlayers {
						if !slices.Equal(g.Players[player].Hand.All(), replay.Players[player].Hand.All()) ||
							!slices.Equal(g.Players[player].CommandZone.All(), replay.Players[player].CommandZone.All()) {
							t.Fatal("replay or the opposite-decision branch mutated the original destinations")
						}
					}
				})
			}
		}
	}
}
