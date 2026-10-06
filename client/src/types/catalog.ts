import {Equatorial} from "@/types/coordinates.ts";

interface ObjectRef {
    id: number,
    class: string,
    common_name: string
}

interface Collection {
    id: number,
    name: string,
    description?: string
}

interface CollectionMember extends ObjectRef {
    identifier: string
}

interface SearchHit extends ObjectRef {
    identifier: string,
    collection: string
}

interface Metadata {
    type_name: string,
    common_name: string,
    description: string,
    image_url: string
}

interface StarData {
    visual_mag: number,
    spectral_class?: string
}

interface DSOData {
    visual_mag?: number,
    major_axis?: number,
    minor_axis?: number,
    pos_angle?: number
}

interface CelestialObject {
    id: number,
    class: string,
    position: Equatorial,
    distance?: number,
    metadata: Metadata
    star?: StarData,
    dso?: DSOData
}

interface SearchRequest {
    query?: string,
    collection?: number,
    limit: number,
    offset: number
}

export type {
    ObjectRef, Collection, CollectionMember, SearchHit, Metadata, StarData, DSOData, CelestialObject, SearchRequest
}