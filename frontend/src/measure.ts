import { useEffect } from 'react'
import { api, type Refused, type Snapshot, type TextSamples } from './api'
import { outlookDays, weekdayLetters } from './Cell'
import { degrees, high, highLow, low } from './units'

/**
 * widestLine answers the width of the widest of texts laid out with className inside parent: one
 * box that shrinks to its widest line, so a single layout pass measures them all.
 */
function widestLine(parent: HTMLElement, className: string, texts: string[]): number {
  const box = parent.ownerDocument.createElement('div')
  box.style.cssText = 'position: absolute; width: max-content'
  for (const text of texts) {
    const line = parent.ownerDocument.createElement('div')
    line.className = className
    line.textContent = text
    box.appendChild(line)
  }
  parent.appendChild(box)
  return box.getBoundingClientRect().width
}

/** sides answers the sum of a computed style's two horizontal lengths named by prefix and suffix. */
function sides(style: CSSStyleDeclaration, prefix: string, suffix: string): number {
  return (parseFloat(style.getPropertyValue(`${prefix}-left${suffix}`)) || 0) + (parseFloat(style.getPropertyValue(`${prefix}-right${suffix}`)) || 0)
}

/**
 * cellWidthNeeded measures, in DIP, how wide a cell must be to show every sample whole in the font
 * this web engine really draws with (FR-103): a label with the widest time beside it, the widest
 * current temperature, the widest day's high and low, the outlook's days side by side, each in the classes a cell's own text carries,
 * plus the cell's padding and border. It measures a cell that follows another, so the divider between
 * cells is counted.
 */
export function cellWidthNeeded(samples: TextSamples, units: string, doc: Document = document): number {
  const ribbon = doc.createElement('div')
  ribbon.className = 'ribbon horizontal'
  ribbon.style.cssText = 'position: absolute; visibility: hidden; left: 0; top: 0; width: auto; height: auto'
  const first = doc.createElement('div')
  const cell = doc.createElement('div')
  first.className = cell.className = 'cell weather'
  ribbon.append(first, cell)
  doc.body.appendChild(ribbon)
  try {
    const temperatures = samples.temperatures.map((temperature) => degrees(temperature, units))
    const ranges = samples.temperatures.map((temperature) => highLow(temperature, temperature, units))
    const outlookLines = samples.temperatures.flatMap((temperature) => [high(temperature, units), low(temperature, units)])
    const text = Math.max(
      widestLine(cell, 'place label', samples.labels) + widestLine(cell, 'place clock', samples.times),
      widestLine(cell, 'temperature', temperatures),
      widestLine(cell, 'today', ranges),
      outlookDays * Math.max(widestLine(cell, 'weekday', samples.weekdays.map((day) => day.slice(0, weekdayLetters))), widestLine(cell, 'range', outlookLines)),
    )
    const style = doc.defaultView?.getComputedStyle(cell)
    const chrome = style == null ? 0 : sides(style, 'padding', '') + sides(style, 'border', '-width')
    return Math.ceil(text + chrome)
  } finally {
    ribbon.remove()
  }
}

/**
 * useMeasuredCells tells Go how wide a cell must be for its widest text whenever the units, the time
 * format or the labels change, then takes the snapshot again so the cells are drawn at the width Go
 * settles on (FR-103). A measurement overtaken by a newer change is dropped.
 */
export function useMeasuredCells(snapshot: Snapshot | null, load: () => void, refused: Refused): void {
  const units = snapshot?.units
  const format = snapshot?.format
  const labels = snapshot == null ? null : JSON.stringify(snapshot.cells.map((cell) => cell.label))
  useEffect(() => {
    if (units == null || format == null || labels == null) {
      return
    }
    let overtaken = false
    const measure = async () => {
      await document.fonts?.ready
      const samples = await api.textSamples(refused)
      if (samples == null || overtaken) {
        return
      }
      const cellWidth = cellWidthNeeded(samples, units)
      const taken = await api.setMeasured({ units, format, labels: samples.labels, cellWidth }, refused)
      if (taken === null || overtaken) {
        return
      }
      load()
    }
    void measure()
    return () => {
      overtaken = true
    }
  }, [units, format, labels, load, refused])
}
