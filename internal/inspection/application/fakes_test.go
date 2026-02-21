package application_test

import (
	"context"
	"sync"

	"github.com/bejayjones/juno/internal/inspection/domain"
)

// fakeInspectionRepo is an in-memory implementation of domain.InspectionRepository.
type fakeInspectionRepo struct {
	mu          sync.Mutex
	inspections map[domain.InspectionID]*domain.Inspection
}

func newFakeRepo() *fakeInspectionRepo {
	return &fakeInspectionRepo{
		inspections: make(map[domain.InspectionID]*domain.Inspection),
	}
}

func (r *fakeInspectionRepo) Save(_ context.Context, insp *domain.Inspection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Store a shallow copy to simulate persistence isolation.
	cp := *insp
	r.inspections[insp.ID] = &cp
	return nil
}

func (r *fakeInspectionRepo) FindByID(_ context.Context, id domain.InspectionID) (*domain.Inspection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	insp, ok := r.inspections[id]
	if !ok {
		return nil, domain.ErrInspectionNotFound
	}
	cp := *insp
	return &cp, nil
}

func (r *fakeInspectionRepo) FindByAppointmentID(_ context.Context, apptID domain.AppointmentID) (*domain.Inspection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, insp := range r.inspections {
		if insp.AppointmentID == apptID {
			cp := *insp
			return &cp, nil
		}
	}
	return nil, domain.ErrInspectionNotFound
}

func (r *fakeInspectionRepo) FindByInspector(_ context.Context, inspectorID domain.InspectorID, filter domain.InspectionFilter) ([]*domain.Inspection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.Inspection
	for _, insp := range r.inspections {
		if insp.InspectorID != inspectorID {
			continue
		}
		if filter.Status != nil && insp.Status != *filter.Status {
			continue
		}
		cp := *insp
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeInspectionRepo) Delete(_ context.Context, id domain.InspectionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.inspections, id)
	return nil
}
