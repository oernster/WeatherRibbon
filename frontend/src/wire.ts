// WeatherRibbon's half of the wire between Go and the page, stated a second time here. dto.go is the
// other statement; ribbonkit states the window's half. A structural test compares each pair.

import type { MenuChoice } from '@oernster/ribbonkit'

export interface Size {
  width: number
  height: number
}

export interface Layout {
  cell: Size
  prompt: Size
  padding: number
}

/** Every text a cell can show whose width varies, for the page to measure (FR-103). */
export interface TextSamples {
  times: string[]
  weekdays: string[]
  temperatures: number[]
  labels: string[]
}

/** The cell width the page measured its widest text to need, with what it measured under (FR-103). */
export interface Measured {
  units: string
  format: string
  labels: string[]
  cellWidth: number
}

/** A weather symbol: the icon drawn for it, else its words in its place (FR-412). */
export interface WeatherSymbol {
  icon: string
  words: string
}

/** One day of a cell (FR-404, FR-406); date is written yyyy-mm-dd. */
export interface Day {
  date: string
  weekday: string
  high: number
  low: number
  rain: number
  symbol: WeatherSymbol
  known: boolean
}

export interface Cell {
  id: string
  label: string
  place: string
  time: string
  zoneMark: string
  temperature: number
  symbol: WeatherSymbol
  today: Day
  outlook: Day[]
  age: string
  problem: string
  petrichor: boolean
}

export interface Snapshot {
  cells: Cell[]
  units: string
  format: string
  colour: string
  orientation: string
  theme: string
  alwaysOnTop: boolean
  /** How opaque the window is drawn in percent; minOpacity the least it may be (FR-704). */
  opacity: number
  minOpacity: number
  /** The percent the ribbon is drawn at; minScale and maxScale bound it (FR-705). */
  scale: number
  minScale: number
  maxScale: number
  layout: Layout
  refreshInMs: number
  notices: string[]
  scrolls: boolean
  dragThreshold: Size
  startLabel: string
  /** True while the window is an unpinned ribbon's tab (FR-706). */
  collapsed: boolean
  /** The menus' choices, which Settings offers as well (FR-701). */
  choices: MenuChoice[]
}

/** One entry of the place search (FR-202). */
export interface Place {
  geoNamesId: number
  name: string
  description: string
}

/** One hour of the detail panel (FR-410). */
export interface Hour {
  time: string
  symbol: WeatherSymbol
  temperature: number
  rain: number
  windSpeed: number
  windFrom: number
}

/** The detail panel for one city (FR-410, FR-411). */
export interface Detail {
  id: string
  label: string
  place: string
  hours: Hour[]
  sunrise: string
  sunset: string
  problem: string
}
