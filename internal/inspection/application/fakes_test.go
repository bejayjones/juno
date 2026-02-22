package application_test

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/bejayjones/juno/internal/inspection/domain"
	"github.com/bejayjones/juno/pkg/storage"
)

// fakePhotoStorage is an in-memory PhotoStorage for tests.
type fakePhotoStorage struct {
	mu    sync.Mutex
	files map[string]string // storagePath → content
}

func newFakeStorage() *fakePhotoStorage {
	return &fakePhotoStorage{files: make(map[string]string)}
}

func (s *fakePhotoStorage) Save(_ context.Context, photoID, ext string, data io.Reader) (string, error) {
	b, _ := io.ReadAll(data)
	path := photoID + ext
	s.mu.Lock()
	s.files[path] = string(b)
	s.mu.Unlock()
	return path, nil
}

func (s *fakePhotoStorage) Get(_ context.Context, storagePath string) (io.ReadCloser, error) {
	s.mu.Lock()
	content, ok := s.files[storagePath]
	s.mu.Unlock()
	if !ok {
		return nil, storage.ErrNotImplemented
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func (s *fakePhotoStorage) Delete(_ context.Context, storagePath string) error {
	s.mu.Lock()
	delete(s.files, storagePath)
	s.mu.Unlock()
	return nil
}

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

func (r *fakeInspectionRepo) FindPhotoMeta(_ context.Context, photoID domain.PhotoID) (storagePath, mimeType string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, insp := range r.inspections {
		for _, section := range insp.Systems {
			for _, item := range section.Items {
				for _, f := range item.Findings {
					for _, p := range f.Photos {
						if p.ID == photoID {
							return p.StoragePath, p.MimeType, nil
						}
					}
				}
			}
		}
	}
	return "", "", domain.ErrPhotoNotFound
}

func (r *fakeInspectionRepo) Delete(_ context.Context, id domain.InspectionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.inspections, id)
	return nil
}
