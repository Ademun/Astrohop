interface Equatorial {
    ra: number,
    dec: number
}

interface Horizontal {
    azimuth: number,
    altitude: number
}

interface GeoLocation {
    lat: number,
    long: number
}

export type {Equatorial, Horizontal, GeoLocation}