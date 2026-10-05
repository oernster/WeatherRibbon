import { useEffect, useState, type KeyboardEvent } from 'react'
import { usePanelFit } from '@oernster/ribbonkit'
import { api, type Detail as DetailData } from './api'
import { degrees } from './Cell'
import { Symbol } from './Symbol'

/** The words each system's rain and wind are measured in (FR-703), one home. */
export const unitWords: Record<string, { rain: string; wind: string }> = {
  metric: { rain: 'mm', wind: 'km/h' },
  imperial: { rain: 'in', wind: 'mph' },
}

/** The places of decimals rain is written to, in each system. */
const rainPlaces: Record<string, number> = { metric: 1, imperial: 2 }

/** A wind arrow points where the wind blows to: half a turn from where it blows from. */
const halfTurn = 180

interface Props {
  id: string
  units: string
  onClose: () => void
  /** ready is true once Go has made the window this panel, so its content can be measured. */
  ready: boolean
}

/**
 * Detail is one city's next 24 hours, each with its local hour, symbol, temperature, rain and wind
 * (FR-410), with the day's sunrise and sunset where they could be had (FR-411). Close and Escape
 * return to the ribbon.
 */
export function Detail({ id, units, onClose, ready }: Props) {
  const [detail, setDetail] = useState<DetailData | null>(null)
  const [problem, setProblem] = useState('')
  const panel = usePanelFit<HTMLElement>(api, setProblem, ready && detail != null)
  const words = unitWords[units] ?? unitWords.metric

  useEffect(() => {
    void api.openDetail(id, setProblem).then(setDetail)
  }, [id])

  const escape = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === 'Escape') {
      onClose()
    }
  }

  return (
    <main className="settings detail" ref={panel} onKeyDown={escape}>
      <header>
        <h1>{detail?.label ?? ''}</h1>
        <button type="button" autoFocus onClick={onClose}>
          Close
        </button>
      </header>
      {problem !== '' && (
        <p className="problem" role="alert">
          {problem}
        </p>
      )}
      {detail != null && (
        <>
          <p className="muted">{detail.place}</p>
          {(detail.sunrise !== '' || detail.sunset !== '') && (
            <p className="sun">
              {detail.sunrise !== '' && <span>Sunrise {detail.sunrise}</span>}
              {detail.sunset !== '' && <span>Sunset {detail.sunset}</span>}
            </p>
          )}
          {detail.problem !== '' ? (
            <p className="problem">{detail.problem}</p>
          ) : (
            <table className="hours">
              <thead>
                <tr>
                  <th scope="col">Time</th>
                  <th scope="col">Weather</th>
                  <th scope="col">Temperature</th>
                  <th scope="col">Rain ({words.rain})</th>
                  <th scope="col">Wind ({words.wind})</th>
                </tr>
              </thead>
              <tbody>
                {detail.hours.map((hour) => (
                  <tr key={hour.time}>
                    <td>{hour.time}</td>
                    <td>
                      <Symbol symbol={hour.symbol} />
                    </td>
                    <td>{degrees(hour.temperature)}</td>
                    <td>{hour.rain.toFixed(rainPlaces[units] ?? rainPlaces.metric)}</td>
                    <td>
                      <span className="wind" aria-label={`from ${Math.round(hour.windFrom)} degrees`} style={{ rotate: `${hour.windFrom + halfTurn}deg` }}>
                        ↑
                      </span>{' '}
                      {hour.windSpeed}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}
    </main>
  )
}
