package mission

import (
	"astrohop/internal/astronomy/coordinates"
	"astrohop/pkg/algo"
	"time"

	"github.com/google/uuid"
)

type MissionDTO struct {
	MissionID uuid.UUID   `json:"mission_id"`
	Data      *DataDTO    `json:"data,omitempty"`
	MapData   *MapDataDTO `json:"map_data,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

func (m *Mission) ToDTO() *MissionDTO {
	if m == nil {
		return nil
	}
	return &MissionDTO{
		MissionID: m.MissionID,
		Data:      m.Data.ToDTO(),
		MapData:   m.MapData.ToDTO(),
		CreatedAt: m.CreatedAt,
	}
}

func (dto *MissionDTO) ToDomain() *Mission {
	if dto == nil {
		return nil
	}
	return &Mission{
		MissionID: dto.MissionID,
		Data:      dto.Data.ToDomain(),
		MapData:   dto.MapData.ToDomain(),
		CreatedAt: dto.CreatedAt,
	}
}

type DataDTO struct {
	Location   coordinates.GeoLocation `json:"location"`
	Time       time.Time               `json:"time"`
	Objectives []ObjectiveDTO          `json:"objectives"`
	Conditions ConditionsDTO           `json:"conditions"`
}

func (d *Data) ToDTO() *DataDTO {
	if d == nil {
		return nil
	}
	objectives := make([]ObjectiveDTO, len(d.Objectives))
	for i, obj := range d.Objectives {
		objectives[i] = obj.ToDTO()
	}
	return &DataDTO{
		Location:   d.Location,
		Time:       d.Time,
		Objectives: objectives,
		Conditions: d.Conditions.ToDTO(),
	}
}

func (dto *DataDTO) ToDomain() *Data {
	if dto == nil {
		return nil
	}
	objectives := make([]Objective, len(dto.Objectives))
	for i, obj := range dto.Objectives {
		objectives[i] = obj.ToDomain()
	}
	return &Data{
		Location:   dto.Location,
		Time:       dto.Time,
		Objectives: objectives,
		Conditions: dto.Conditions.ToDomain(),
	}
}

type MapDataDTO struct {
	Positions []coordinates.Horizontal `json:"positions"`
	Tour      *algo.Tour               `json:"tour"`
}

func (md *MapData) ToDTO() *MapDataDTO {
	if md == nil {
		return nil
	}
	return &MapDataDTO{
		Positions: md.Positions,
		Tour:      md.Tour,
	}
}

func (dto *MapDataDTO) ToDomain() *MapData {
	if dto == nil {
		return nil
	}
	return &MapData{
		Positions: dto.Positions,
		Tour:      dto.Tour,
	}
}

type ObjectiveDTO struct {
	OID  int64  `json:"oid" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (o Objective) ToDTO() ObjectiveDTO {
	return ObjectiveDTO{
		OID:  o.OID,
		Name: o.Name,
	}
}

func (dto ObjectiveDTO) ToDomain() Objective {
	return Objective{
		OID:  dto.OID,
		Name: dto.Name,
	}
}

type ConditionsDTO struct {
	LimitingMagnitude *float32 `json:"limiting_magnitude" binding:"required"`
}

func (c Conditions) ToDTO() ConditionsDTO {
	return ConditionsDTO{
		LimitingMagnitude: &c.LimitingMagnitude,
	}
}

func (dto ConditionsDTO) ToDomain() Conditions {
	return Conditions{
		LimitingMagnitude: *dto.LimitingMagnitude,
	}
}
