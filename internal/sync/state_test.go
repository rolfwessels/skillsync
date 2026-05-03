package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateState(t *testing.T) {
	const (
		hashA = "aaaa"
		hashB = "bbbb"
	)

	tests := []struct {
		name              string
		giveRegistryHash  string
		giveLocalHash     string
		giveLockRegHash   string
		giveLockLocalHash string
		want              State
	}{
		{
			name:              "no change on either side",
			giveRegistryHash:  hashA,
			giveLocalHash:     hashA,
			giveLockRegHash:   hashA,
			giveLockLocalHash: hashA,
			want:              StateClean,
		},
		{
			name:              "registry changed, local unchanged",
			giveRegistryHash:  hashB,
			giveLocalHash:     hashA,
			giveLockRegHash:   hashA,
			giveLockLocalHash: hashA,
			want:              StateRegistryModified,
		},
		{
			name:              "local changed, registry unchanged",
			giveRegistryHash:  hashA,
			giveLocalHash:     hashB,
			giveLockRegHash:   hashA,
			giveLockLocalHash: hashA,
			want:              StateLocalModified,
		},
		{
			name:              "both changed",
			giveRegistryHash:  hashB,
			giveLocalHash:     hashB,
			giveLockRegHash:   hashA,
			giveLockLocalHash: hashA,
			want:              StateConflict,
		},
		{
			name:              "no lock entry treats as registry modified",
			giveRegistryHash:  hashA,
			giveLocalHash:     "",
			giveLockRegHash:   "",
			giveLockLocalHash: "",
			want:              StateRegistryModified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateState(tt.giveRegistryHash, tt.giveLocalHash, tt.giveLockRegHash, tt.giveLockLocalHash)
			assert.Equal(t, tt.want, got)
		})
	}
}
