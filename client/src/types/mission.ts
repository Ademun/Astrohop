import {Horizontal} from "@/types/coordinates.ts";
import {MapData} from "@/types/api.ts";

interface MissionID {
    mission_id: string
}

interface MissionFull {
    id: string,
    information: Information,
    data: Data,
    map?: Map,
    overrides?: Overrides
    is_public: boolean,
    created_at: string
    modified_at?: string
}

interface MissionShort {
    id: string,
    information: Information,
    is_public: boolean,
    created_at: string
}

interface Information {
    name?: string,
    description?: string
}

interface Data {
    location: Geolocation,
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
    payload?: MapData;
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