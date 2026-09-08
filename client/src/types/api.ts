export interface AstroObject {
  oid: number;
  name: string;
  type: string;
}

export interface GeoLocation {
  lat: number;
  long: number;
}

export interface Objective {
  oid: number;
  name: string;
}

export interface Conditions {
  limiting_magnitude: number;
}

export interface MissionData {
  location: GeoLocation;
  time: string; // ISO 8601 date-time
  objectives: Objective[];
  conditions: Conditions;
}

export interface Horizontal {
  alt: number;
  az: number;
}

export interface MapData {
  moon_position: Horizontal;
  positions: Record<number, Horizontal>;
  tour: number[];
}

export interface Mission {
  mission_id: string;
  data?: MissionData | null;
  map_data?: MapData | null;
  created_at: string;
}

export interface CreateMissionResponse {
  mission_id: string;
}

export type MissionProgressStatus = "building_route" | "done" | "failed";

export interface MissionProgressEvent {
  progress: MissionProgressStatus;
  payload?: MapData;
  error?: string;
}

export interface ApiErrorBody {
  error: string;
}

export interface UnauthorizedErrorBody {
  message: string;
}
