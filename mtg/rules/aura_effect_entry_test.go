package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type auraEntryChoiceAgent struct{}

func (auraEntryChoiceAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return action.Pass()
}

func (auraEntryChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Choose what the Aura enchants." {
		return []int{len(request.Options) - 1}
	}
	return request.DefaultSelection
}

func TestAuraEffectEntryAdmitsNonTargetingAttachmentBeforeMovement(t *testing.T) {
	for _, scenario := range []string{
		"no host", "wrong type", "shroud", "hexproof", "protection", "legal host",
		"foreign control", "actual controller", "multiple hosts", "diverted entry", "batch no host", "batch legal host",
	} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			spec := &game.TargetSpec{
				Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Creature},
				}),
			}
			if scenario == "foreign control" || scenario == "actual controller" {
				spec.Selection.Val.Controller = game.ControllerYou
			}
			def := &game.CardDef{CardFace: game.CardFace{
				Name: "Incoming Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
				Colors: []color.Color{color.Black}, StaticAbilities: []game.StaticAbility{game.EnchantStaticAbility(spec)},
			}}
			card := &game.CardInstance{ID: g.IDGen.Next(), Owner: game.Player3, Def: def, ZoneVersion: 7}
			g.CardInstances[card.ID] = card
			g.Players[game.Player3].Graveyard.Add(card.ID)
			var host *game.Permanent
			if scenario != "no host" && scenario != "batch no host" {
				hostDef := vanillaCreature("Legal Host", 2, 3)
				switch scenario {
				case "wrong type":
					hostDef.Types = []types.Card{types.Artifact}
				case "shroud":
					hostDef.StaticAbilities = []game.StaticAbility{game.ShroudStaticBody}
				case "hexproof":
					hostDef.StaticAbilities = []game.StaticAbility{game.HexproofStaticBody}
				case "protection":
					hostDef.StaticAbilities = []game.StaticAbility{{
						KeywordAbilities: []game.KeywordAbility{game.ProtectionKeyword{FromColors: []color.Color{color.Black}}},
					}}
				default:
				}
				host = addCombatPermanent(g, game.Player4, hostDef)
				if scenario == "actual controller" {
					host.Controller = game.Player2
				}
				if scenario == "multiple hosts" {
					host = addCombatPermanent(g, game.Player4, vanillaCreature("Chosen Second Host", 2, 3))
				}
			}
			if scenario == "diverted entry" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Graveyard,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
				})
			}
			var agents [game.NumPlayers]PlayerAgent
			for i := range agents {
				agents[i] = auraEntryChoiceAgent{}
			}
			resolver := &effectResolver{
				engine: engine, game: g, obj: &game.StackObject{Controller: game.Player2}, agents: agents,
			}
			beforeID, beforeEvents := g.IDGen.Current(), len(g.Events)
			var entered *game.Permanent
			var success bool
			if scenario == "batch no host" || scenario == "batch legal host" {
				second := &game.CardInstance{ID: g.IDGen.Next(), Owner: game.Player3, Def: def, ZoneVersion: 9}
				g.CardInstances[second.ID] = second
				g.Players[game.Player3].Graveyard.Add(second.ID)
				beforeID = g.IDGen.Current()
				entries := resolver.putResolvedCardsOnBattlefieldCollecting([]resolvedBattlefieldCard{
					{card: card, fromZone: zone.Graveyard, controller: game.Player2},
					{card: second, fromZone: zone.Graveyard, controller: game.Player2},
				}, nil, permanentCreationOptions{})
				if scenario == "batch no host" && (!g.Players[game.Player3].Graveyard.Contains(second.ID) || second.ZoneVersion != 9) {
					t.Fatal("batch no-host refusal moved the second exact card")
				}
				if scenario == "batch legal host" && len(entries) != 2 {
					t.Fatal("simultaneous legal Aura admission did not preserve both entries")
				}
				if len(entries) > 0 {
					entered, success = entries[0], true
				}
			} else {
				entered, success = resolver.putResolvedCardOnBattlefieldValue(
					card, zone.Graveyard, game.Player2, nil, permanentCreationOptions{},
				)
			}
			refused := scenario == "no host" || scenario == "wrong type" || scenario == "protection" ||
				scenario == "foreign control" || scenario == "batch no host"
			if refused {
				if success || entered != nil || !g.Players[game.Player3].Graveyard.Contains(card.ID) ||
					card.ZoneVersion != 7 || g.IDGen.Current() != beforeID || len(g.Events) != beforeEvents ||
					permanentForCard(g, card.ID) != nil {
					t.Fatal("refused Aura entry changed zone/version/identity/events or reported success")
				}
				return
			}
			if scenario == "diverted entry" {
				if success || entered != nil || !g.Players[game.Player3].Exile.Contains(card.ID) ||
					card.ZoneVersion != 8 || len(host.Attachments) != 0 {
					t.Fatal("replacement diversion was lost or attached a nonexistent Aura")
				}
				return
			}
			if !success || entered == nil || entered.Owner != game.Player3 || entered.Controller != game.Player2 ||
				card.ZoneVersion != 8 || g.Players[game.Player3].Graveyard.Contains(card.ID) ||
				!entered.AttachedTo.Exists || entered.AttachedTo.Val != host.ObjectID {
				t.Fatal("legal non-targeting attachment did not enter once under the actual recipient")
			}
			entryEvents := 0
			for _, event := range g.Events[beforeEvents:] {
				if event.CardID == card.ID && event.Kind == game.EventPermanentEnteredBattlefield {
					entryEvents++
				}
			}
			if entryEvents != 1 {
				t.Fatal("legal entry emitted an incorrect number of entry events")
			}
		})
	}
}
