package mission

import (
	"astrohop/internal/astronomy/coordinates"
	"time"
)

// ---------- MissionFullDTO ----------

type MissionFullDTO struct {
	Information InformationDTO `json:"information"`
	Data        DataDTO        `json:"data"`
	Map         *MapDTO        `json:"map,omitempty"`
	Overrides   *Overrides     `json:"overrides,omitempty"`
	IsPublic    bool           `json:"is_public"`
	CreatedAt   time.Time      `json:"created_at"`
	ModifiedAt  *time.Time     `json:"modified_at,omitempty"`
}

func NewMissionFullDTO(m *Mission) MissionFullDTO {
	dto := MissionFullDTO{
		Information: NewInformationDTO(&m.Information),
		Data:        NewDataDTO(&m.Data),
		Overrides:   m.Overrides,
		IsPublic:    m.IsPublic,
		CreatedAt:   m.CreatedAt,
		ModifiedAt:  m.ModifiedAt,
	}
	if m.Map != nil {
		dto.Map = NewMapDTO(m.Map)
	}
	return dto
}

func (d MissionFullDTO) ToDomain() *Mission {
	m := &Mission{
		Information: d.Information.ToDomain(),
		Data:        *d.Data.ToDomain(),
		Overrides:   d.Overrides,
		IsPublic:    d.IsPublic,
		CreatedAt:   d.CreatedAt,
		ModifiedAt:  d.ModifiedAt,
	}
	if d.Map != nil {
		m.Map = d.Map.ToDomain()
	}
	return m
}

type MissionShortDTO struct {
	Information InformationDTO `json:"information"`
	IsPublic    bool           `json:"is_public"`
	CreatedAt   time.Time      `json:"created_at"`
}

func NewMissionShortDTO(m *Mission) MissionShortDTO {
	return MissionShortDTO{
		Information: NewInformationDTO(&m.Information),
		IsPublic:    m.IsPublic,
		CreatedAt:   m.CreatedAt,
	}
}

type InformationDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func NewInformationDTO(i *Information) InformationDTO {
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

type DataDTO struct {
	Location   coordinates.GeoLocation `json:"location"`
	Time       time.Time               `json:"time"`
	Objectives []ObjectiveDTO          `json:"objectives"`
	Conditions ConditionsDTO           `json:"conditions"`
}

func NewDataDTO(dt *Data) DataDTO {
	objectives := make([]ObjectiveDTO, len(dt.Objectives))
	for i := range dt.Objectives {
		objectives[i] = NewObjectiveDTO(&dt.Objectives[i])
	}
	return DataDTO{
		Location:   dt.Location,
		Time:       dt.Time,
		Objectives: objectives,
		Conditions: NewConditionsDTO(&dt.Conditions),
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

type MapDTO struct {
	MoonPosition coordinates.Horizontal           `json:"moon_position"`
	Positions    map[int64]coordinates.Horizontal `json:"positions"`
	Tour         []int64                          `json:"tour"`
}

func NewMapDTO(m *Map) *MapDTO {
	return &MapDTO{
		MoonPosition: m.MoonPosition,
		Positions:    m.Positions,
		Tour:         m.Tour,
	}
}

func (d MapDTO) ToDomain() *Map {
	return &Map{
		MoonPosition: d.MoonPosition,
		Positions:    d.Positions,
		Tour:         d.Tour,
	}
}

type ObjectiveDTO struct {
	OID  int64  `json:"oid" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func NewObjectiveDTO(o *Objective) ObjectiveDTO {
	return ObjectiveDTO{OID: o.OID, Name: o.Name}
}

func (d ObjectiveDTO) ToDomain() *Objective {
	return &Objective{OID: d.OID, Name: d.Name}
}

type ConditionsDTO struct {
	LimitingMagnitude float32 `json:"limiting_magnitude"`
}

func NewConditionsDTO(c *Conditions) ConditionsDTO {
	return ConditionsDTO{LimitingMagnitude: c.LimitingMagnitude}
}

func (d ConditionsDTO) ToDomain() *Conditions {
	return &Conditions{LimitingMagnitude: d.LimitingMagnitude}
}
