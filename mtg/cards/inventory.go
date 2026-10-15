package cards

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// Support separates installed code, behavior evidence and compiler maintenance.
// A legacy name is not an Oracle identity, even if a corpus has one name match.
type Support struct {
	Key              string                  `json:"key"`
	Name             string                  `json:"name"`
	OracleID         string                  `json:"oracle_id,omitempty"`
	Face             game.FaceIndex          `json:"face"`
	Faces            []string                `json:"printed_faces,omitempty"`
	Identity         string                  `json:"identity"`
	Implementation   string                  `json:"implementation"`
	Loadable         bool                    `json:"loadable"`
	BehaviorVerified bool                    `json:"behavior_verified"`
	Verification     string                  `json:"verification,omitempty"`
	Failure          string                  `json:"failure,omitempty"`
	SourceRevision   string                  `json:"source_revision,omitempty"`
	RulesRevision    string                  `json:"rules_revision,omitempty"`
	Programs         []game.ProgramReference `json:"programs,omitempty"`
}

// SupportFor uses the same checked loader as decks. Exact identities are retained;
// name-only legacy entries are never silently promoted to exact face identities.
func (r *Registry) SupportFor(name string) []Support {
	entries := r.identities[name]
	defs, err := r.LookupAllChecked(name)
	out := make([]Support, 0, len(entries))
	for i, entry := range entries {
		support := Support{
			Key: fmt.Sprintf("legacy-name:%s/%d", name, i), Name: name,
			Identity: "name-only", Implementation: "legacy",
			Loadable: err == nil, Face: game.FaceFront,
		}
		if entry.OracleID != "" {
			support.Key = entry.OracleID + "/front"
			support.OracleID = entry.OracleID
			support.Identity = "exact-oracle-face"
			support.Implementation = "authored"
			support.Verification = entry.Verification
			support.SourceRevision = entry.SourceRevision
			support.RulesRevision = entry.RulesRevision
			support.BehaviorVerified = err == nil && entry.Verification != "" &&
				entry.SourceRevision != "" && entry.RulesRevision != ""
		}
		if err != nil {
			support.Failure = err.Error()
		} else if i < len(defs) {
			for _, faceIndex := range defs[i].FaceIndexes() {
				if face, exists := defs[i].Face(faceIndex); exists {
					support.Faces = append(support.Faces, face.Name)
				}
			}
			support.Programs = defs[i].ProgramReferences()
			for _, program := range entry.Programs {
				if !slices.Contains(support.Programs, program.Reference) {
					support.Programs = append(support.Programs, program.Reference)
				}
			}
		}
		out = append(out, support)
	}
	return out
}

// Inventory is a deterministic view of every production registration, not the
// compiler's accepted set. It is intentionally explicit (and eager), unlike
// startup registration or a deck lookup.
func (r *Registry) Inventory() []Support {
	names := r.All()
	slices.Sort(names)
	var out []Support
	for _, name := range names {
		out = append(out, r.SupportFor(name)...)
	}
	return out
}
