package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

type bandReplayMode int

const (
	bandReplayExact bandReplayMode = iota
	bandReplayStaged
	bandReplayDefault
)

// bandReplayPolicy declares every banding creature as attackers. Exact mode
// proposes a complete declaration banding all attackers that share a target;
// staged mode builds the same through choices, and default mode answers each
// band choice with its default (no band).
type bandReplayPolicy struct {
	mode bandReplayMode
}

func (p bandReplayPolicy) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	if _, attacking := legal[0].DeclareAttackersPayload(); attacking {
		if p.mode != bandReplayExact {
			return action.ChooseAttackers()
		}
		for _, candidate := range legal {
			payload, ok := candidate.DeclareAttackersPayload()
			if !ok || len(payload.Attackers) < 2 {
				continue
			}
			band := game.AttackingBand{}
			for _, attack := range payload.Attackers {
				if attack.Target != payload.Attackers[0].Target {
					band.Members = nil
					break
				}
				band.Members = append(band.Members, attack.Attacker)
			}
			if len(band.Members) >= 2 {
				return action.DeclareAttackersWithBands(payload.Attackers, []game.AttackingBand{band})
			}
		}
	}
	return legal[0]
}

func (p bandReplayPolicy) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceCombatDeclaration:
		for _, option := range request.Options {
			if option.Attack.Exists && option.Attack.Val.Target.IsPlayerAttack() {
				return []int{option.Index}
			}
		}
	case game.ChoiceAttackingBand:
		if p.mode == bandReplayStaged && len(request.Options) > 1 {
			return []int{request.Options[1].Index}
		}
	default:
	}
	return agent.BaselineStrategy{}.ChooseChoice(rules.PlayerObservation{}, request)
}

// TestAttackingBandReplayRoundTrip records games whose attackers form bands by
// an exact proposal, by staged choices, and by the staged default (no bands).
// The logged declarations carry their bands, and the JSON replay record
// reproduces each game exactly.
func TestAttackingBandReplayRoundTrip(t *testing.T) {
	for _, test := range []struct {
		name  string
		mode  bandReplayMode
		bands bool
	}{
		{"exact", bandReplayExact, true},
		{"staged", bandReplayStaged, true},
		{"default", bandReplayDefault, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			knight := &game.CardDef{CardFace: game.CardFace{
				Name: "Replay Banding Knight", Types: []types.Card{types.Creature},
				Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
				StaticAbilities: []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Haste, game.Banding)}},
			}}
			cfg := Config{Games: 1, Seed: 3481, TurnLimit: 2}
			for player := range cfg.Configs {
				cfg.Configs[player].Deck = []*game.CardDef{knight, knight, knight}
				for range 5 {
					cfg.Configs[player].Deck = append(cfg.Configs[player].Deck, forest())
				}
			}
			cfg.NewAgents = func(uint64) [game.NumPlayers]rules.PlayerAgent {
				var policies [game.NumPlayers]rules.PlayerAgent
				for player := range policies {
					policies[player] = bandReplayPolicy{mode: test.mode}
				}
				return policies
			}
			result, record := RecordGame(cfg, 0)
			multiAttacks, banded := 0, 0
			for _, turn := range result.Turns {
				for _, logged := range turn.Actions {
					declare, ok := logged.Action.DeclareAttackersPayload()
					if !ok {
						continue
					}
					if len(declare.Attackers) >= 2 {
						multiAttacks++
					}
					for _, band := range declare.Bands {
						members := make(map[id.ID]bool)
						for _, attack := range declare.Attackers {
							members[attack.Attacker] = true
						}
						for _, member := range band.Members {
							if !members[member] {
								t.Fatalf("band member %d was not declared: %+v", member, declare)
							}
						}
						banded++
					}
				}
			}
			if multiAttacks == 0 {
				t.Fatal("no declaration attacked with two creatures")
			}
			if (banded > 0) != test.bands {
				t.Fatalf("recorded %d bands, want bands %v", banded, test.bands)
			}
			data, err := json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			var decoded ReplayRecord
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if replayed := Replay(cfg.Configs, decoded); !reflect.DeepEqual(replayed, result) {
				t.Fatal("attacking band JSON replay diverged")
			}
			if rerun := RunOne(cfg, 0); !reflect.DeepEqual(rerun, result) {
				t.Fatal("same seeded band policies were not reproducible")
			}
		})
	}
}
