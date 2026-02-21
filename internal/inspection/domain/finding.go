package domain

import "time"

type FindingID string
type PhotoID string

type PhotoRef struct {
	ID          PhotoID
	StoragePath string
	CapturedAt  time.Time
}

type Finding struct {
	ID           FindingID
	Narrative    string
	IsDeficiency bool
	Photos       []PhotoRef
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewFinding(id FindingID, narrative string, isDeficiency bool, now time.Time) Finding {
	return Finding{
		ID:           id,
		Narrative:    narrative,
		IsDeficiency: isDeficiency,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (f *Finding) Update(narrative string, isDeficiency bool, now time.Time) {
	f.Narrative = narrative
	f.IsDeficiency = isDeficiency
	f.UpdatedAt = now
}

func (f *Finding) AddPhoto(photo PhotoRef) {
	f.Photos = append(f.Photos, photo)
}

func (f *Finding) RemovePhoto(id PhotoID) bool {
	for i, p := range f.Photos {
		if p.ID == id {
			f.Photos = append(f.Photos[:i], f.Photos[i+1:]...)
			return true
		}
	}
	return false
}
