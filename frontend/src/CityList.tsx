import { useState } from 'react'
import type { Cell } from './api'

interface Props {
  cells: Cell[]
  onRename: (id: string, label: string) => void
  onChangePlace: (cell: Cell) => void
  onRemove: (id: string) => void
}

/**
 * CityList is the configured cities in the ribbon's order, east from Greenwich (FR-108, FR-204,
 * FR-205, FR-207). The order follows the time, so rows are not moved by hand; Remove asks first,
 * naming the city.
 */
export function CityList({ cells, onRename, onChangePlace, onRemove }: Props) {
  const [confirming, setConfirming] = useState<Cell | null>(null)

  if (cells.length === 0) {
    return <p className="muted">No cities yet. Add one below.</p>
  }
  return (
    <>
      <ol className="cities" aria-label="Cities">
        {cells.map((cell) => (
          <li key={cell.id}>
            <div className="row-main">
              <input
                key={cell.label}
                aria-label={`Label for ${cell.place === '' ? cell.label : cell.place}`}
                defaultValue={cell.label}
                onBlur={(event) => {
                  if (event.target.value !== cell.label) {
                    onRename(cell.id, event.target.value)
                  }
                }}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    event.currentTarget.blur()
                  }
                }}
              />
              <span className={cell.problem !== '' && cell.place === '' ? 'problem' : 'muted place-name'}>
                {cell.place === '' ? cell.problem : cell.place}
              </span>
            </div>
            <button type="button" onClick={() => onChangePlace(cell)} title="Change place">
              Change place
            </button>
            <button type="button" aria-label={`Remove ${cell.label}`} title="Remove" onClick={() => setConfirming(cell)}>
              Remove
            </button>
          </li>
        ))}
      </ol>
      {confirming != null && (
        <div className="dialog" role="alertdialog" aria-modal="true" aria-label="Remove city">
          <p>Remove {confirming.label}?</p>
          <button
            type="button"
            autoFocus
            onClick={() => {
              onRemove(confirming.id)
              setConfirming(null)
            }}
          >
            Remove
          </button>
          <button type="button" onClick={() => setConfirming(null)}>
            Cancel
          </button>
        </div>
      )}
    </>
  )
}
