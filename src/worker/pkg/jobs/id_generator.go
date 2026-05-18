package jobs

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// IDGenerator creates persisted identifiers for jobs package records.
type IDGenerator interface {
	NewID() (string, error)
}

// ULIDGenerator creates ULID identifiers with instance-scoped monotonic entropy.
type ULIDGenerator struct {
	mu      sync.Mutex
	entropy *ulid.MonotonicEntropy
}

// NewULIDGenerator creates a production ULID generator.
func NewULIDGenerator() *ULIDGenerator {
	return &ULIDGenerator{
		entropy: ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0), //nolint:gosec // Non-crypto seed for ULID entropy
	}
}

// NewID returns a new ULID string.
func (g *ULIDGenerator) NewID() (string, error) {
	g.mu.Lock()
	id, err := ulid.New(ulid.Timestamp(time.Now()), g.entropy)
	g.mu.Unlock()

	if err != nil {
		return "", fmt.Errorf("jobs: ulid generation failed: %w", err)
	}

	return id.String(), nil
}
