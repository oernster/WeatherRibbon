import { useEffect, useId, useRef, useState, type KeyboardEvent } from 'react'
import { api, type Place, type Refused } from './api'
import { ArtButton } from './ArtButton'

interface Props {
  /** heading names what choosing a place does, such as "Add a city". */
  heading: string
  onChoose: (place: Place) => void
  refused: Refused
  /** autoFocus puts the cursor in the search as it appears. */
  autoFocus: boolean
  /** onCancel gives the search a Cancel button, Escape calling it, for a search opened for one purpose. */
  onCancel?: () => void
  /** picture, for a search that stays open, is drawn beside the box; pressing it chooses the highlighted place. */
  picture?: { art: string; label: string }
}

/** The most results listed at once; typing narrows the rest. */
const shown = 50

/**
 * PlaceSearch finds a city by its name, region or country, offline (FR-202, CON-9). Up and Down move
 * through the results, Enter chooses, Escape cancels. Nothing is listed until something is typed.
 */
export function PlaceSearch({ heading, onChoose, refused, autoFocus, onCancel, picture }: Props) {
  const [typed, setTyped] = useState('')
  const [places, setPlaces] = useState<Place[]>([])
  const [active, setActive] = useState(0)
  const listId = useId()
  const box = useRef<HTMLInputElement>(null)
  const listed = typed.trim() !== ''

  const choose = (place: Place) => {
    onChoose(place)
    if (onCancel == null) {
      setTyped('')
    }
  }

  const pressed = () => {
    if (listed && places[active] != null) {
      choose(places[active])
    } else {
      box.current?.focus()
    }
  }

  useEffect(() => {
    if (!listed) {
      setPlaces([])
      return
    }
    let current = true
    void api.searchPlaces(typed, refused).then((found) => {
      if (current && found != null) {
        setPlaces(found.slice(0, shown))
        setActive(0)
      }
    })
    return () => {
      current = false
    }
  }, [typed, listed, refused])

  const key = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActive((index) => Math.min(index + 1, places.length - 1))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActive((index) => Math.max(index - 1, 0))
    } else if (event.key === 'Enter' && listed && places[active] != null) {
      event.preventDefault()
      choose(places[active])
    } else if (event.key === 'Escape' && (onCancel != null || typed !== '')) {
      event.preventDefault()
      event.stopPropagation()
      if (onCancel != null) {
        onCancel()
      } else {
        setTyped('')
      }
    }
  }

  return (
    <section className="search" aria-label={heading}>
      <h2>{heading}</h2>
      <div className="search-row">
        <input
          ref={box}
          autoFocus={autoFocus}
          type="search"
          placeholder="City, region or country"
          aria-label="Search cities"
          aria-controls={listId}
          aria-activedescendant={listed && places[active] != null ? `${listId}-${active}` : undefined}
          value={typed}
          onChange={(event) => setTyped(event.target.value)}
          onKeyDown={key}
        />
        {picture != null && <ArtButton art={picture.art} label={picture.label} large onClick={pressed} />}
      </div>
      <ul id={listId} role="listbox" aria-label="Cities" hidden={!listed}>
        {listed && places.map((place, index) => (
          <li
            key={place.geoNamesId}
            id={`${listId}-${index}`}
            role="option"
            aria-selected={index === active}
            className={index === active ? 'active' : ''}
            onMouseEnter={() => setActive(index)}
            onClick={() => choose(place)}
          >
            {place.description}
          </li>
        ))}
        {listed && places.length === 0 && <li className="muted">No city matches</li>}
      </ul>
      {onCancel != null && (
        <button type="button" onClick={onCancel}>
          Cancel
        </button>
      )}
    </section>
  )
}
