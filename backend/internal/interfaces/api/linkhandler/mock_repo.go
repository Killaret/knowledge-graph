package linkhandler

import (
	"context"
	"sync"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
)

type mockLinkRepo struct {
	mu           sync.RWMutex
	links        map[uuid.UUID]*link.Link
	suppressions []*link.Suppression
}

func newMockLinkRepo() *mockLinkRepo {
	return &mockLinkRepo{
		links: make(map[uuid.UUID]*link.Link),
	}
}

func (m *mockLinkRepo) Save(ctx context.Context, l *link.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.links[l.ID()] = l
	return nil
}

func (m *mockLinkRepo) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.links[id]
	if !ok {
		return nil, nil
	}
	return l, nil
}

func (m *mockLinkRepo) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*link.Link
	for _, l := range m.links {
		if l.SourceNoteID() == sourceID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockLinkRepo) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*link.Link
	for _, l := range m.links {
		if l.TargetNoteID() == targetID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockLinkRepo) FindByPair(ctx context.Context, sourceID, targetID uuid.UUID) ([]*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*link.Link
	for _, l := range m.links {
		if l.SourceNoteID() == sourceID && l.TargetNoteID() == targetID {
			result = append(result, l)
		}
	}
	return result, nil
}

// SaveUserLink mirrors the postgres create-or-promote semantics for handler
// tests: same-type gamma is promoted in place, other-type gammas on the pair
// transfer provenance into the new row, matching suppressions are lifted.
func (m *mockLinkRepo) SaveUserLink(ctx context.Context, l *link.Link) (*link.Link, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var sameType *link.Link
	var gammas []*link.Link
	for _, existing := range m.links {
		samePair := existing.SourceNoteID() == l.SourceNoteID() && existing.TargetNoteID() == l.TargetNoteID()
		reversePair := existing.SourceNoteID() == l.TargetNoteID() && existing.TargetNoteID() == l.SourceNoteID()
		if !samePair && !reversePair {
			continue
		}
		if existing.LinkType().String() == l.LinkType().String() {
			if existing.SourceType().IsUser() {
				return nil, false, link.ErrDuplicateLink
			}
			sameType = existing
		}
		if existing.SourceType().IsGamma() {
			gammas = append(gammas, existing)
		}
	}

	if sameType != nil {
		sameType.PromoteToUser(l.CreatorID(), l.LinkType(), l.Weight(), l.Metadata())
		if sameType.SourceNoteID() != l.SourceNoteID() || sameType.TargetNoteID() != l.TargetNoteID() {
			sameType = link.ReconstructLinkWithCreator(sameType.ID(), l.SourceNoteID(), l.TargetNoteID(),
				sameType.LinkType(), sameType.Weight(), sameType.Metadata(), sameType.SourceType(),
				sameType.CreatorID(), sameType.CreatedAt(), sameType.UpdatedAt(), sameType.LastWeightUpdate())
		}
		for _, g := range gammas {
			if g.ID() != sameType.ID() {
				delete(m.links, g.ID())
			}
		}
		m.links[sameType.ID()] = sameType
		m.liftSuppressions(l.SourceNoteID(), l.TargetNoteID(), l.LinkType().String())
		return sameType, false, nil
	}

	if len(gammas) > 0 {
		l.InheritGammaProvenance(gammas[0])
		for _, g := range gammas {
			delete(m.links, g.ID())
		}
	}
	m.links[l.ID()] = l
	m.liftSuppressions(l.SourceNoteID(), l.TargetNoteID(), l.LinkType().String())
	return l, true, nil
}

func (m *mockLinkRepo) DeleteAndSuppress(ctx context.Context, l *link.Link, s *link.Suppression) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s != nil {
		m.suppressions = append(m.suppressions, s)
	}
	delete(m.links, l.ID())
	return nil
}

func (m *mockLinkRepo) SaveSuppression(ctx context.Context, s *link.Suppression) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.suppressions {
		if existing.NoteAID() == s.NoteAID() && existing.NoteBID() == s.NoteBID() {
			sameType := (existing.LinkType() == nil && s.LinkType() == nil) ||
				(existing.LinkType() != nil && s.LinkType() != nil && *existing.LinkType() == *s.LinkType())
			if sameType {
				return nil
			}
		}
	}
	m.suppressions = append(m.suppressions, s)
	return nil
}

func (m *mockLinkRepo) FindSuppressionsForNotes(ctx context.Context, noteIDs []uuid.UUID) ([]*link.Suppression, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	set := make(map[uuid.UUID]bool, len(noteIDs))
	for _, id := range noteIDs {
		set[id] = true
	}
	var result []*link.Suppression
	for _, s := range m.suppressions {
		if set[s.NoteAID()] || set[s.NoteBID()] {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockLinkRepo) liftSuppressions(sourceID, targetID uuid.UUID, linkType string) {
	noteA, noteB := link.NormalizePair(sourceID, targetID)
	kept := m.suppressions[:0]
	for _, s := range m.suppressions {
		lifted := s.NoteAID() == noteA && s.NoteBID() == noteB &&
			(s.LinkType() == nil || *s.LinkType() == linkType)
		if !lifted {
			kept = append(kept, s)
		}
	}
	m.suppressions = kept
}

func (m *mockLinkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.links, id)
	return nil
}

func (m *mockLinkRepo) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, l := range m.links {
		if l.SourceNoteID() == sourceID {
			delete(m.links, id)
		}
	}
	return nil
}

func (m *mockLinkRepo) Update(ctx context.Context, l *link.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[l.ID()]; !ok {
		return link.ErrLinkNotFound
	}
	m.links[l.ID()] = l
	return nil
}

func (m *mockLinkRepo) List(ctx context.Context) ([]*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*link.Link
	for _, l := range m.links {
		result = append(result, l)
	}
	return result, nil
}

func (m *mockLinkRepo) FindAll(ctx context.Context) ([]*link.Link, error) {
	return m.List(ctx)
}

func (m *mockLinkRepo) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allLinks []*link.Link
	for _, l := range m.links {
		allLinks = append(allLinks, l)
	}

	total := int64(len(allLinks))

	if offset >= len(allLinks) {
		return []*link.Link{}, total, nil
	}

	end := offset + limit
	if end > len(allLinks) {
		end = len(allLinks)
	}
	if limit == 0 {
		end = len(allLinks)
	}

	return allLinks[offset:end], total, nil
}

// Create is a convenience method for tests
func (m *mockLinkRepo) Create(ctx context.Context, sourceID, targetID uuid.UUID, linkType string, weight float64, metadata map[string]interface{}, sourceType string) *link.Link {
	// Convert types to domain types
	linkTypeDomain, _ := link.NewLinkType(linkType)
	weightDomain, _ := link.NewWeight(weight)
	metadataDomain, _ := link.NewMetadata(metadata)

	newLink := link.NewLink(sourceID, targetID, linkTypeDomain, weightDomain, metadataDomain)
	m.Save(ctx, newLink)
	return newLink
}
