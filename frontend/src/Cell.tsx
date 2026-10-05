import type { Cell as CellData, Day } from './api'
import { Symbol } from './Symbol'

/** The petrichor line's words, one home (FR-413). */
export const petrichorLine = 'Can you smell petrichor in the air?'

/** degrees marks a whole temperature, whichever units it is in (FR-703). */
export const degrees = (temperature: number) => `${temperature}°`

/** The days an outlook shows (FR-406) and the weekday's first letters each shows, so they fit across a cell. */
export const outlookDays = 3
export const weekdayLetters = 3

interface Props {
  cell: CellData
  onOpen: (id: string) => void
  onDismissPetrichor: (id: string) => void
}

/** Place names the city: its label, then its local time and zone mark (FR-401). */
function Place({ cell }: { cell: CellData }) {
  return (
    <div className="place" title={cell.place === '' ? cell.label : cell.place}>
      <span className="label">{cell.label}</span>
      {cell.time !== '' && (
        <span className="clock">
          {' '}
          {cell.time}
          {cell.zoneMark !== '' && ` ${cell.zoneMark}`}
        </span>
      )}
    </div>
  )
}

/** Outlook is the three days after today, each with its weekday, symbol, high and low (FR-406). */
function Outlook({ days }: { days: Day[] }) {
  return (
    <div className="outlook">
      {days.map((day) => (
        <div key={day.date} className="day" aria-label={day.weekday}>
          <span className="weekday">{day.weekday.slice(0, weekdayLetters)}</span>
          {day.known ? (
            <>
              <Symbol symbol={day.symbol} />
              <span className="range">
                {degrees(day.high)} {degrees(day.low)}
              </span>
            </>
          ) : (
            <span className="muted">?</span>
          )}
        </div>
      ))}
    </div>
  )
}

/**
 * Cell is one city. A city that cannot show the weather says why in words, never another place's
 * weather (FR-305, FR-403, FR-803, FR-804, NFR-U-2). A click opens the hourly detail (FR-410); a click
 * on the petrichor line hides it and opens nothing (FR-413).
 */
export function Cell({ cell, onOpen, onDismissPetrichor }: Props) {
  const open = () => onOpen(cell.id)
  if (cell.problem !== '') {
    return (
      <div className="cell" role="group" aria-label={cell.label} onClick={open}>
        <Place cell={cell} />
        <div className="problem">{cell.problem}</div>
      </div>
    )
  }
  const today = cell.today
  return (
    <div className="cell weather" role="group" aria-label={`${cell.label}, ${cell.time}, ${degrees(cell.temperature)}`} onClick={open}>
      <Place cell={cell} />
      <div className="now">
        <Symbol symbol={cell.symbol} large />
        <span className="temperature">{degrees(cell.temperature)}</span>
      </div>
      {today.known && (
        <div className="today">
          {degrees(today.high)} {degrees(today.low)}
          {today.rain > 0 && <span className="rain"> {today.rain.toFixed(1)}</span>}
        </div>
      )}
      <Outlook days={cell.outlook} />
      {cell.age !== '' && <div className="age">{cell.age}</div>}
      {cell.petrichor && (
        <button
          type="button"
          className="petrichor"
          onClick={(event) => {
            event.stopPropagation()
            onDismissPetrichor(cell.id)
          }}
        >
          {petrichorLine}
        </button>
      )}
    </div>
  )
}
