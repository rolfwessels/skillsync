package importcmd

import (
	"fmt"

	"github.com/rolfwessels/skillsync/internal/bundle"
)

type StubPrompter struct {
	Indices   []int
	Decisions []Decision
}

func (s *StubPrompter) PickGroups(all []Group) ([]Group, error) {
	var out []Group
	for _, i := range s.Indices {
		if i >= 0 && i < len(all) {
			out = append(out, all[i])
		}
	}
	return out, nil
}

func (s *StubPrompter) Decide(_ Group, _ []bundle.Bundle) (Decision, error) {
	if len(s.Decisions) == 0 {
		return Decision{}, fmt.Errorf("stub: exhausted decisions")
	}
	d := s.Decisions[0]
	s.Decisions = s.Decisions[1:]
	return d, nil
}

func (StubPrompter) Confirm(string) error { return nil }
