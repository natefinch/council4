package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

type combatReplayPolicy struct {
	staged bool
}

func (p combatReplayPolicy) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	if _, attacking := legal[0].DeclareAttackersPayload(); attacking {
		if p.staged {
			return action.ChooseAttackers()
		}
		for _, first := range legal {
			payload, present := first.DeclareAttackersPayload()
			if !present || len(payload.Attackers) < 3 {
				continue
			}
			for _, second := range legal {
				other, exists := second.DeclareAttackersPayload()
				if exists && len(other.Attackers) > 0 && payload.Attackers[0].Target != other.Attackers[0].Target {
					mixed := []game.AttackDeclaration{payload.Attackers[0], payload.Attackers[1], payload.Attackers[2]}
					mixed[2].Target = other.Attackers[0].Target
					return action.DeclareAttackers(mixed)
				}
			}
		}
	}
	if _, blocking := legal[0].DeclareBlockersPayload(); blocking {
		if p.staged {
			return action.ChooseBlockers()
		}
		for _, first := range legal {
			payload, present := first.DeclareBlockersPayload()
			if !present || len(payload.Blockers) != 1 {
				continue
			}
			for _, second := range legal {
				other, exists := second.DeclareBlockersPayload()
				if exists && len(other.Blockers) == 1 &&
					payload.Blockers[0].Blocker != other.Blockers[0].Blocker && payload.Blockers[0].Blocking != other.Blockers[0].Blocking {
					return action.DeclareBlockers([]game.BlockDeclaration{payload.Blockers[0], other.Blockers[0]})
				}
			}
		}
	}
	return legal[0]
}

func (combatReplayPolicy) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceCombatDeclaration {
		for _, option := range request.Options {
			if option.Attack.Exists || option.Block.Exists {
				return []int{option.Index}
			}
		}
	}
	return agent.BaselineStrategy{}.ChooseChoice(rules.PlayerObservation{}, request)
}

func TestCompleteCombatReplayRoundTrip(t *testing.T) {
	for _, staged := range []bool{false, true} {
		name := "exact"
		if staged {
			name = "staged"
		}
		t.Run(name, func(t *testing.T) {
			bear := &game.CardDef{CardFace: game.CardFace{
				Name: "Replay Vigilant Bear", Types: []types.Card{types.Creature},
				Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 3}),
				StaticAbilities: []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Haste, game.Vigilance)}},
			}}
			cfg := Config{Games: 1, Seed: 3126, TurnLimit: 2}
			for player := range cfg.Configs {
				cfg.Configs[player].Deck = []*game.CardDef{bear, bear, bear}
				for range 5 {
					cfg.Configs[player].Deck = append(cfg.Configs[player].Deck, forest())
				}
			}
			cfg.NewAgents = func(uint64) [game.NumPlayers]rules.PlayerAgent {
				var policies [game.NumPlayers]rules.PlayerAgent
				for player := range policies {
					policies[player] = combatReplayPolicy{staged: staged}
				}
				return policies
			}
			result, record := RecordGame(cfg, 0)
			capturedAttack, capturedBlock, choices := false, false, 0
			for _, seat := range record.Seats {
				choices += len(seat.Choices)
				for _, proposal := range seat.Announcements {
					if proposal.Attack != nil {
						for _, attack := range proposal.Attack.Attackers {
							capturedAttack = capturedAttack || attack.Target != proposal.Attack.Attackers[0].Target
						}
					}
					capturedBlock = capturedBlock || proposal.Block != nil && len(proposal.Block.Blockers) >= 2
				}
			}
			// Staged markers are sampled; exact mixed declarations are not.
			if !staged && (!capturedAttack || !capturedBlock) {
				t.Fatalf("did not record complete proposals: attack %v, block %v", capturedAttack, capturedBlock)
			}
			if staged && choices == 0 {
				t.Fatal("did not record staged combat choices")
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
				t.Fatal("combat declaration JSON replay diverged")
			}
			if rerun := RunOne(cfg, 0); !reflect.DeepEqual(rerun, result) {
				t.Fatal("same seeded combat policies were not reproducible")
			}
		})
	}
}
