// Typed access to the Go facade: ribbonkit's window calls plus WeatherRibbon's own, over one guarded
// call (ribbonkit's bridge). Wrapping them here keeps the binding shape in one file.

import { connect, windowCalls, type Refused, type WindowBridge } from '@oernster/ribbonkit'
import type { Detail, Measured, Place, Snapshot, TextSamples } from './wire'

export type { Refused } from '@oernster/ribbonkit'
export type { Cell, Day, Detail, Hour, Measured, MenuChoice, Place, Snapshot, TextSamples, WeatherSymbol } from './wire'

/** Bridge is the App Wails binds: the window's methods (WindowBridge) and WeatherRibbon's own (app.go). */
interface Bridge extends WindowBridge {
  Snapshot(): Promise<Snapshot>
  SearchPlaces(typed: string): Promise<Place[]>
  AddCity(geoNamesId: number): Promise<string>
  RenameCity(id: string, typed: string): Promise<void>
  ChangePlace(id: string, geoNamesId: number): Promise<void>
  RemoveCity(id: string): Promise<void>
  SetUnits(units: string): Promise<void>
  SetFormat(format: string): Promise<void>
  DismissNotices(): Promise<void>
  DismissPetrichor(id: string): Promise<void>
  OpenDetail(id: string): Promise<Detail>
  TextSamples(): Promise<TextSamples>
  SetMeasured(measured: Measured): Promise<void>
}

const call = connect<Bridge>('WeatherRibbon')

export const api = {
  ...windowCalls(call),
  snapshot: (refused: Refused) => call((b) => b.Snapshot(), refused),
  searchPlaces: (typed: string, refused: Refused) => call((b) => b.SearchPlaces(typed), refused),
  addCity: (geoNamesId: number, refused: Refused) => call((b) => b.AddCity(geoNamesId), refused),
  renameCity: (id: string, typed: string, refused: Refused) => call((b) => b.RenameCity(id, typed), refused),
  changePlace: (id: string, geoNamesId: number, refused: Refused) => call((b) => b.ChangePlace(id, geoNamesId), refused),
  removeCity: (id: string, refused: Refused) => call((b) => b.RemoveCity(id), refused),
  setUnits: (units: string, refused: Refused) => call((b) => b.SetUnits(units), refused),
  setFormat: (format: string, refused: Refused) => call((b) => b.SetFormat(format), refused),
  dismissNotices: (refused: Refused) => call((b) => b.DismissNotices(), refused),
  dismissPetrichor: (id: string, refused: Refused) => call((b) => b.DismissPetrichor(id), refused),
  openDetail: (id: string, refused: Refused) => call((b) => b.OpenDetail(id), refused),
  textSamples: (refused: Refused) => call((b) => b.TextSamples(), refused),
  setMeasured: (measured: Measured, refused: Refused) => call((b) => b.SetMeasured(measured), refused),
}
