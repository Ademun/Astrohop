const PARSEC_TO_LIGHT_YEARS = 3.26156;
const PLACEHOLDER = "—";

function pad(value: number): string {
    return String(value).padStart(2, "0");
}

function sexagesimal(total: number, unitsPerCycle: number) {
    const tenths = Math.round(total * 3600 * 10) % (unitsPerCycle * 36000);
    return {
        whole: Math.floor(tenths / 36000),
        minutes: Math.floor((tenths % 36000) / 600),
        seconds: ((tenths % 600) / 10).toFixed(1).padStart(4, "0"),
    };
}

export function formatNumber(
    value: number | null,
    digits = 2,
    unit = "",
): string {
    if (value === null) return PLACEHOLDER;
    const formatted = value.toLocaleString("en-US", {
        maximumFractionDigits: digits,
    });
    return unit ? `${formatted}${unit}` : formatted;
}

export function formatRightAscension(degrees: number): string {
    const normalized = ((degrees % 360) + 360) % 360;
    const { whole, minutes, seconds } = sexagesimal(normalized / 15, 24);
    return `${whole}h ${pad(minutes)}m ${seconds}s`;
}

export function formatDeclination(degrees: number): string {
    const { whole, minutes, seconds } = sexagesimal(Math.abs(degrees), 360);
    const sign = degrees < 0 ? "−" : "+";
    return `${sign}${pad(whole)}°${pad(minutes)}′${seconds}″`;
}

export function formatDistance(parsecs: number | null): string {
    if (parsecs === null) return "Unknown";
    const digits = parsecs < 100 ? 2 : 0;
    return `${formatNumber(parsecs, digits, "pc")} (${formatNumber(
        parsecs * PARSEC_TO_LIGHT_YEARS,
        digits,
        "ly",
    )})`;
}

export function formatAngularSize(
    major: number | null,
    minor: number | null,
): string {
    if (major === null && minor === null) return PLACEHOLDER;
    if (major === null || minor === null) {
        return `${formatNumber(major ?? minor, 1)}′`;
    }
    if (Math.abs(major - minor) < 0.05) return `${formatNumber(major, 1)}′`;
    return `${formatNumber(major, 1)}'×${formatNumber(minor, 1)}′`;
}