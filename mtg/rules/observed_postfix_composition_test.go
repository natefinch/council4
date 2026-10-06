package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

type observedPostfixAgent struct {
	libraryPaymentAgent
	answers   []bool
	afterScry func()
}

func (a *observedPostfixAgent) ChooseChoice(observation PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		a.prompts = append(a.prompts, request)
		index := len(a.prompts) - 1
		if index < len(a.answers) && a.answers[index] {
			return []int{1}
		}
		return []int{0}
	}
	if request.Kind == game.ChoiceScry && a.afterScry != nil {
		a.afterScry()
	}
	return request.DefaultSelection
}

func TestCompiledObservedPostfixUnavailableAndSkippedProducts(t *testing.T) {
	body := "Reveal the top card of your library. Scry 1. You may put that card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, put that card into your hand. You gain 1 life."
	for _, outcome := range []string{"empty", "stale before predicate", "declined observation", "declined placement"} {
		t.Run(outcome, func(t *testing.T) {
			text := body
			if outcome == "declined observation" {
				text = "You may reveal" + body[len("Reveal"):]
			}
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Unavailable Postfix Subject", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = 40
			card := addCardToLibrary(g, game.Player1, compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Forest", Layout: "normal", TypeLine: "Basic Land — Forest", OracleText: "{T}: Add {G}.",
			}))
			agent := &observedPostfixAgent{answers: []bool{false}}
			if outcome == "empty" {
				g.Players[game.Player1].Library.Remove(card)
			}
			if outcome == "stale before predicate" {
				agent.afterScry = func() {
					if !moveCardBetweenZones(g, game.Player1, card, zone.Library, zone.Exile) ||
						!moveCardBetweenZones(g, game.Player1, card, zone.Exile, zone.Library) {
						t.Fatal("actual observation leave/reentry failed")
					}
				}
			}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
				ResolutionResults: map[string]game.InstructionResolutionResult{"if-you-do": {Accepted: false}},
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			wantPrompts := 0
			if outcome == "declined observation" || outcome == "declined placement" {
				wantPrompts = 1
			}
			if len(agent.prompts) != wantPrompts || g.Players[game.Player1].Life != 41 ||
				g.Players[game.Player1].Hand.Contains(card) != (outcome == "declined placement") ||
				permanentForCard(g, card) != nil {
				t.Errorf("unavailable/skipped=%s choices=%d hand=%v battlefield=%v life=%d",
					outcome, len(agent.prompts), g.Players[game.Player1].Hand.Contains(card),
					permanentForCard(g, card) != nil, g.Players[game.Player1].Life)
			}
			if len(obj.ResolutionResults) != 1 || obj.ResolutionResults["if-you-do"].Accepted ||
				len(obj.LocalLinkedProducts) != 0 || len(g.LinkedObjects) != 0 {
				t.Error("skipped action consumed or leaked a parent receipt/product")
			}
			t.Logf("availability=%s prompts=%d hand=%v version=%d", outcome, len(agent.prompts),
				g.Players[game.Player1].Hand.Contains(card), g.CardInstances[card].ZoneVersion)
		})
	}
}

func TestCompiledObservedPostfixIndependentOwnersAndModes(t *testing.T) {
	body := "Reveal the top card of your library. Look at the top card of target player's library. You may put that card into that player's hand if it's a creature card with mana value 3 or less. Otherwise, you gain 2 life. Put the revealed card into your hand. You gain 1 life."
	for _, modal := range []bool{false, true} {
		for _, outcome := range []string{"accepted", "declined", "nonqualifying", "unselected"} {
			t.Run(outcome+map[bool]string{false: "/spell", true: "/mode"}[modal], func(t *testing.T) {
				text := body
				if modal {
					text = "Choose one \u2014\n\u2022 " + body + "\n\u2022 You gain 1 life."
				}
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Independent Observations", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
				})
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Players[game.Player1].Life = 40
				ours := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Divination", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleText: "Draw two cards.",
				})
				theirs := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Observed Horse", Layout: "normal", TypeLine: "Creature — Horse", ManaCost: "{3}", Power: new("2"), Toughness: new("2"),
				})
				if outcome == "nonqualifying" {
					theirs = ours
				}
				obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
					Targets:           []game.Target{game.PlayerTarget(game.Player3)},
					ChosenModes:       []int{0},
					ResolutionResults: map[string]game.InstructionResolutionResult{"if-you-do": {Accepted: false}},
				}
				if modal && outcome == "unselected" {
					obj.ChosenModes[0] = 1
				}
				for iteration := range 2 {
					oursID := addCardToLibrary(g, game.Player1, ours)
					theirsID := addCardToLibrary(g, game.Player3, theirs)
					agent := &observedPostfixAgent{answers: []bool{outcome == "accepted"}}
					NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
						[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
					selected := !modal || outcome != "unselected"
					wantPrompts := 0
					if selected && outcome != "nonqualifying" {
						wantPrompts = 1
					}
					if len(agent.prompts) != wantPrompts ||
						g.Players[game.Player1].Hand.Contains(oursID) != selected ||
						g.Players[game.Player3].Hand.Contains(theirsID) != (selected && outcome == "accepted") {
						t.Error("independent observation, selected owner, or decision scope was aliased")
					}
					wantLife := 40 + iteration + 1
					if selected && outcome != "accepted" {
						wantLife += 2 * (iteration + 1)
					}
					if g.Players[game.Player1].Life != wantLife || len(obj.ResolutionResults) != 1 ||
						obj.ResolutionResults["if-you-do"].Accepted || len(obj.LocalLinkedProducts) != 0 || len(g.LinkedObjects) != 0 {
						t.Errorf("invocation=%d life=%d want=%d or local product leaked", iteration, g.Players[game.Player1].Life, wantLife)
					}
				}
			})
		}
	}
}

func TestCompiledObservedShuffleLibraryOwnerIdentity(t *testing.T) {
	warp := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Chaos Warp", Layout: "normal", TypeLine: "Instant", ManaCost: "{2}{R}",
		OracleText: "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.",
	})
	creature := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Owner's Horse", Layout: "normal", TypeLine: "Creature — Horse", ManaCost: "{4}",
		Power: new("2"), Toughness: new("2"),
	})
	divination := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Divination", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleText: "Draw two cards.",
	})
	for _, diverted := range []bool{false, true} {
		t.Run(map[bool]string{false: "actual shuffled permanent", true: "actual nonpermanent after diverted shuffle"}[diverted], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatPermanent(g, game.Player2, creature)
			target.Owner = game.Player3
			g.CardInstances[target.CardInstanceID].Owner = game.Player3
			casterDecoy := addCardToLibrary(g, game.Player1, divination)
			controllerDecoy := addCardToLibrary(g, game.Player2, divination)
			observed := target.CardInstanceID
			if diverted {
				observed = addCardToLibrary(g, game.Player3, divination)
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged, MatchFromZone: true,
					FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Library, ReplaceToZone: zone.Graveyard,
				})
			}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, warp.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{})
			reveals := 0
			for _, event := range g.Events {
				if event.Kind == game.EventCardRevealed {
					reveals++
					if event.Player != game.Player3 || event.CardID != observed {
						t.Errorf("actual reveal=%+v, want owner/card %v/%v", event, game.Player3, observed)
					}
				}
			}
			returned := permanentForCard(g, target.CardInstanceID)
			if reveals != 1 || (returned != nil) == diverted ||
				!g.Players[game.Player1].Library.Contains(casterDecoy) ||
				!g.Players[game.Player2].Library.Contains(controllerDecoy) {
				t.Error("owner-library observation used caster/controller, or put a nonpermanent onto battlefield")
			}
			if returned != nil && (returned.Owner != game.Player3 || returned.Controller != game.Player3 ||
				returned.ObjectID == target.ObjectID) {
				t.Error("returned actual card did not enter as a new owner's incarnation")
			}
			t.Logf("owner=%v observed=%v diverted=%v reveals=%d returned=%v", game.Player3, observed, diverted, reveals, returned != nil)
		})
	}
}
