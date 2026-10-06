import { describe, expect, it } from 'vitest'
import { degrees, high, highLow, highLowSaid, low, rainOf } from './units'

describe('units (FR-404, FR-406, FR-703)', () => {
  it('writes a temperature with its scale in each system', () => {
    expect(degrees(14, 'metric')).toBe('14°C')
    expect(degrees(58, 'imperial')).toBe('58°F')
    expect(degrees(-3, 'metric')).toBe('-3°C')
  })

  it('marks a high and a low by their letters', () => {
    expect(high(20, 'metric')).toBe('H 20°C')
    expect(low(14, 'metric')).toBe('L 14°C')
    expect(highLow(68, 57, 'imperial')).toBe('H 68°F L 57°F')
    expect(highLowSaid(20, 14, 'metric')).toBe('high 20°C, low 14°C')
  })

  it('writes rain with its measure to the places its system keeps', () => {
    expect(rainOf(0.4, 'metric')).toBe('0.4 mm')
    expect(rainOf(0.016, 'imperial')).toBe('0.02 in')
  })

  it('writes in Metric for a system it does not know', () => {
    expect(degrees(14, 'unknown')).toBe('14°C')
    expect(rainOf(0.4, 'unknown')).toBe('0.4 mm')
  })
})
