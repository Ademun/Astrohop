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

export type ObjectClass = "star" | "dso" | "other";

export interface ObjectRef {
    id: number;
    class: ObjectClass;
    common_name: string | null;
}

export interface Collection {
    id: number;
    name: string;
    description: string | null;
}

export interface SearchHit extends ObjectRef{
    identifier: string;
    collection: string;
}

export interface Equatorial {
    ra: number;
    dec: number;
}

export interface ObjectMetadata {
    type_name: string;
    common_name: string | null;
    description: string | null;
    image_url: string | null;
}

export interface StarData {
    visual_mag: number;
    spectral_class: string | null;
}

export interface DsoData {
    visual_mag: number | null;
    major_axis: number | null;
    minor_axis: number | null;
    pos_angle: number | null;
}

export interface CatalogObject {
    id: number;
    class: ObjectClass;
    position: Equatorial;
    distance_pc: number | null;
    metadata: ObjectMetadata;
    star?: StarData;
    dso?: DsoData;
}

export interface SearchRequest {
    query?: string;
    collection?: number;
    limit: number;
    offset: number;
}