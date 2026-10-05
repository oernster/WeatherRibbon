// A stand-in for the Go facade in tests: every call is recorded, every answer is canned. The window's
// half is ribbonkit's own stand-in; WeatherRibbon's methods are added to it here.

import { install, windowBridge } from '@oernster/ribbonkit/testing'
import { vi } from 'vitest'
import type { Cell, Day, Detail, MenuChoice, Place, Snapshot } from './wire'

export function day(overrides: Partial<Day> = {}): Day {
  return {
    date: '2026-10-06', weekday: 'Tuesday', high: 16, low: 9, rain: 0.4,
    symbol: { icon: 'cloudy', words: '' }, known: true, ...overrides,
  }
}

export function cell(overrides: Partial<Cell> = {}): Cell {
  return {
    id: 'london', label: 'London', place: 'London, England, United Kingdom', time: '08:36', zoneMark: 'BST',
    temperature: 14, symbol: { icon: 'rain', words: '' }, today: day({ date: '2026-10-05', weekday: 'Monday' }),
    outlook: [day(), day({ date: '2026-10-07', weekday: 'Wednesday' }), day({ date: '2026-10-08', weekday: 'Thursday' })],
    age: '', problem: '', petrichor: false, ...overrides,
  }
}

function item(action: string, label: string, checked?: boolean): MenuChoice {
  return { action, label, checkable: checked != null, checked: checked === true, children: [] }
}

function group(label: string, children: MenuChoice[]): MenuChoice {
  return { action: '', label, checkable: false, checked: false, children }
}

/** The menus' choices as Go sends them, in their order; Colour cut to two schemes. */
export const choices: MenuChoice[] = [
  group('Units', [item('units-metric', 'Metric', true), item('units-imperial', 'Imperial', false)]),
  group('Colour', [item('colour-classic', 'Classic', true), item('colour-neon', 'Neon', false)]),
  group('Orientation', [item('horizontal', 'Horizontal', false), item('vertical', 'Vertical', true)]),
  group('Position', [item('left-edge', 'Centre on left edge'), item('right-edge', 'Centre on right edge')]),
  item('always-on-top', 'Always on top', false),
  item('pin', 'Pin ribbon', true),
]

export function snapshot(overrides: Partial<Snapshot> = {}): Snapshot {
  return {
    cells: [cell(), cell({ id: 'tokyo', label: 'Tokyo', place: 'Tokyo, Japan', time: '16:36', zoneMark: 'JST' })],
    units: 'metric', format: '24h', colour: 'classic', orientation: 'vertical', theme: 'system', alwaysOnTop: false,
    opacity: 100, minOpacity: 20, scale: 100, minScale: 75, maxScale: 200,
    layout: { cell: { width: 196, height: 212 }, prompt: { width: 196, height: 212 }, padding: 6 },
    refreshInMs: 60000, notices: [], scrolls: false, dragThreshold: { width: 4, height: 4 }, startLabel: 'Start with Windows',
    collapsed: false, choices, ...overrides,
  }
}

export const places: Place[] = [
  { geoNamesId: 2643743, name: 'London', description: 'London, England, United Kingdom' },
  { geoNamesId: 6058560, name: 'London', description: 'London, Ontario, Canada' },
]

export function detail(overrides: Partial<Detail> = {}): Detail {
  return {
    id: 'london', label: 'London', place: 'London, England, United Kingdom', sunrise: '07:07', sunset: '18:29', problem: '',
    hours: [{ time: '09:00', symbol: { icon: 'rain', words: '' }, temperature: 14, rain: 0.2, windSpeed: 18, windFrom: 225 }],
    ...overrides,
  }
}

/** installBridge puts a recording facade on window and answers it. */
export function installBridge() {
  return install({
    ...windowBridge(),
    Snapshot: vi.fn(async () => snapshot()),
    SearchPlaces: vi.fn(async () => places),
    AddCity: vi.fn(async () => 'new'),
    RenameCity: vi.fn(async () => undefined),
    ChangePlace: vi.fn(async () => undefined),
    RemoveCity: vi.fn(async () => undefined),
    SetUnits: vi.fn(async () => undefined),
    SetFormat: vi.fn(async () => undefined),
    DismissNotices: vi.fn(async () => undefined),
    DismissPetrichor: vi.fn(async () => undefined),
    OpenDetail: vi.fn(async () => detail()),
    TextSamples: vi.fn(async () => ({ times: ['00:00', '23:59'], weekdays: ['Sunday', 'Wednesday'], temperatures: [-60, 60], labels: ['London'] })),
    SetMeasured: vi.fn(async () => undefined),
  })
}
