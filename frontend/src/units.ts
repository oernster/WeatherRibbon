// The page's one home for what each system's measures are written in (FR-703): every number the page
// draws says what it measures, so none is a bare figure.

/** UnitWords are a system's measures: its temperature mark, then what rain and wind are measured in. */
interface UnitWords {
  temperature: string
  rain: string
  wind: string
  /** rainPlaces are the decimals rain is written to. */
  rainPlaces: number
}

const systems: Record<string, UnitWords> = {
  metric: { temperature: '°C', rain: 'mm', wind: 'km/h', rainPlaces: 1 },
  imperial: { temperature: '°F', rain: 'in', wind: 'mph', rainPlaces: 2 },
}

/** unitsOf answers what the system named units is written in; Metric for a name it does not know. */
export function unitsOf(units: string): UnitWords {
  return systems[units] ?? systems.metric
}

/** degrees writes a whole temperature with its scale, as 16°C. */
export const degrees = (temperature: number, units: string) => `${temperature}${unitsOf(units).temperature}`

/** high and low write one temperature marked by its letter, as H 20°C and L 14°C (FR-404, FR-406). */
export const high = (temperature: number, units: string) => `H ${degrees(temperature, units)}`
export const low = (temperature: number, units: string) => `L ${degrees(temperature, units)}`

/** highLow writes a day's high and low on one line, as H 20°C L 14°C (FR-404). */
export const highLow = (top: number, bottom: number, units: string) => `${high(top, units)} ${low(bottom, units)}`

/** highLowSaid is highLow in words, for a screen reader and the tooltip. */
export const highLowSaid = (high: number, low: number, units: string) => `high ${degrees(high, units)}, low ${degrees(low, units)}`

/** rainOf writes an amount of rain with its measure, as 0.4 mm. */
export const rainOf = (rain: number, units: string) => `${rain.toFixed(unitsOf(units).rainPlaces)} ${unitsOf(units).rain}`
