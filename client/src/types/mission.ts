import {GeoLocation, Horizontal} from "@/types/coordinates.ts";

interface MissionID {
    mission_id: string
}

interface MissionFull {
    mission_id: string,
    information: Information,
    data: Data,
    map?: Map,
    overrides?: Overrides
    is_public: boolean,
    created_at: string
    modified_at?: string
}

interface MissionShort {
    mission_id: string,
    information: Information,
    is_public: boolean,
    created_at: string
}

interface Information {
    name?: string,
    description?: string
}

interface Data {
    location: GeoLocation,
    time: string,
    objectives: Array<Objective>,
    conditions: Conditions
}

interface Map {
    moon_position: Horizontal,
    positions: Record<number, Horizontal>
    tour: Array<number>
}

interface Objective {
    oid: number,
    name: string
}

interface Conditions {
    limiting_magnitude: number
}

interface Overrides {
}

type ProgressStatus = "building_route" | "done" | "failed";

interface ProgressEvent {
    progress: ProgressStatus;
    payload?: Map;
    error?: string;
}

export type {
    MissionID,
    MissionFull,
    MissionShort,
    Information,
    Data,
    Map,
    Objective,
    Conditions,
    Overrides,
    ProgressStatus,
    ProgressEvent
}