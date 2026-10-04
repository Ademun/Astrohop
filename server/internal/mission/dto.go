package mission

import (
	"astrohop/internal/astronomy/coordinates"
	"time"
)

type MissionFullDTO struct {
	ID          string         `json:"mission_id"`
	Information InformationDTO `json:"information"`
	Data        DataDTO        `json:"data"`
	Map         *MapDTO        `json:"map,omitempty"`
	Overrides   *Overrides     `json:"overrides,omitempty"`
	IsPublic    bool           `json:"is_public"`
	CreatedAt   time.Time      `json:"created_at"`
	ModifiedAt  *time.Time     `json:"modified_at,omitempty"`
}

type MissionShortDTO struct {
	ID          string         `json:"mission_id"`
	Information InformationDTO `json:"information"`
	IsPublic    bool           `json:"is_public"`
	CreatedAt   time.Time      `json:"created_at"`
}

type InformationDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type DataDTO struct {
	Location   coordinates.GeoLocation `json:"location"`
	Time       time.Time               `json:"time"`
	Objectives []ObjectiveDTO          `json:"objectives"`
	Conditions ConditionsDTO           `json:"conditions"`
}

type HorizontalDTO struct {
	Azimuth  float64 `json:"azimuth"`
	Altitude float64 `json:"altitude"`
}

type MapDTO struct {
	MoonPosition HorizontalDTO           `json:"moon_position"`
	Positions    map[int64]HorizontalDTO `json:"positions"`
	Tour         []int64                 `json:"tour"`
}

type ObjectiveDTO struct {
	OID  int64  `json:"oid" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type ConditionsDTO struct {
	LimitingMagnitude float32 `json:"limiting_magnitude"`
}

func newMissionFullDTO(m *Mission) MissionFullDTO {
	dto := MissionFullDTO{
		ID:          m.MissionID.String(),
		Information: newInformationDTO(&m.Information),
		Data:        newDataDTO(&m.Data),
		Overrides:   m.Overrides,
		IsPublic:    m.IsPublic,
		CreatedAt:   m.CreatedAt,
		ModifiedAt:  m.ModifiedAt,
	}
	if m.Map != nil {
		dto.Map = newMapDTO(m.Map)
	}
	return dto
}

func newMissionShortDTO(m *Mission) MissionShortDTO {
	return MissionShortDTO{
		ID:          m.MissionID.String(),
		Information: newInformationDTO(&m.Information),
		IsPublic:    m.IsPublic,
		CreatedAt:   m.CreatedAt,
	}
}

func newInformationDTO(i *Information) InformationDTO {
	return InformationDTO{
		Name:        i.Name,
		Description: i.Description,
	}
}

func (d InformationDTO) ToDomain() Information {
	return Information{
		Name:        d.Name,
		Description: d.Description,
	}
}

func newDataDTO(dt *Data) DataDTO {
	objectives := make([]ObjectiveDTO, len(dt.Objectives))
	for i := range dt.Objectives {
		objectives[i] = newObjectiveDTO(&dt.Objectives[i])
	}
	return DataDTO{
		Location:   dt.Location,
		Time:       dt.Time,
		Objectives: objectives,
		Conditions: newConditionsDTO(&dt.Conditions),
	}
}

func (d DataDTO) ToDomain() *Data {
	objectives := make([]Objective, len(d.Objectives))
	for i := range d.Objectives {
		objectives[i] = *d.Objectives[i].ToDomain()
	}
	return &Data{
		Location:   d.Location,
		Time:       d.Time,
		Objectives: objectives,
		Conditions: *d.Conditions.ToDomain(),
	}
}

func newHorizontalDTO(h coordinates.Horizontal) HorizontalDTO {
	return HorizontalDTO{
		Azimuth:  coordinates.Deg(h.Az),
		Altitude: coordinates.Deg(h.Alt),
	}
}

func newMapDTO(m *Map) *MapDTO {
	positions := make(map[int64]HorizontalDTO, len(m.Positions))
	for id, p := range m.Positions {
		positions[id] = newHorizontalDTO(p)
	}
	return &MapDTO{
		MoonPosition: newHorizontalDTO(m.MoonPosition),
		Positions:    positions,
		Tour:         m.Tour,
	}
}

func newObjectiveDTO(o *Objective) ObjectiveDTO {
	return ObjectiveDTO{OID: o.OID, Name: o.Name}
}

func (d ObjectiveDTO) ToDomain() *Objective {
	return &Objective{OID: d.OID, Name: d.Name}
}

func newConditionsDTO(c *Conditions) ConditionsDTO {
	return ConditionsDTO{LimitingMagnitude: c.LimitingMagnitude}
}

func (d ConditionsDTO) ToDomain() *Conditions {
	return &Conditions{LimitingMagnitude: d.LimitingMagnitude}
}
