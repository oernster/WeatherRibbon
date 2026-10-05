import { useCallback, useEffect, useState, type KeyboardEvent } from 'react'
import { api, type Cell, type Place, type Refused, type Snapshot } from './api'
import { CityList } from './CityList'
import { PlaceSearch } from './PlaceSearch'
import { ArtButton, addCityTip } from './ArtButton'
import addCityArt from './assets/add-city.png'
import donateMark from './assets/donate.png'
import { MenuGroup, MenuToggle, OpacitySlider, usePanelFit } from '@oernster/ribbonkit'

/** The picture alone does not say pressing it leaves the application, so the tip does. */
export const donateTip = 'Buy the author a drink (opens your browser)'

interface Props {
  snapshot: Snapshot
  /** startAdding puts the cursor in the place search, as Add city does. */
  startAdding: boolean
  reload: () => void
  onClose: () => void
  /** ready is true once Go has made the window this panel, so its content can be measured. */
  ready?: boolean
}

interface Choice {
  label: string
  value: string
}

/**
 * The choices Settings alone offers, with their values in the words shown (FR-701). The menus'
 * choices, Units among them, come from Go in the snapshot, so their words have one home.
 */
const choices: { name: string; key: 'format' | 'theme'; options: Choice[] }[] = [
  { name: 'Time format', key: 'format', options: [{ label: '24-hour', value: '24h' }, { label: '12-hour', value: '12h' }] },
  { name: 'Theme', key: 'theme', options: [{ label: 'System', value: 'system' }, { label: 'Light', value: 'light' }, { label: 'Dark', value: 'dark' }] },
]

const setters = { format: api.setFormat, theme: api.setTheme }

/**
 * Settings is the cities plus every choice, the menus' included (FR-701). Every change applies and
 * is kept at once, with no Save step (FR-702). Escape closes it.
 */
export function Settings({ snapshot, startAdding, reload, onClose, ready = true }: Props) {
  const [moving, setMoving] = useState<Cell | null>(null)
  const [problem, setProblem] = useState('')
  const [startAtSignIn, setStartAtSignIn] = useState<boolean | null>(null)
  const panel = usePanelFit<HTMLElement>(api, setProblem, ready)

  useEffect(() => {
    void api.startWithWindows(setProblem).then(setStartAtSignIn)
  }, [])

  const then = useCallback(() => reload(), [reload])
  const refused: Refused = setProblem
  const choose = (action: string) => void api.choose(action, refused).then(then)

  const added = (place: Place) => void api.addCity(place.geoNamesId, refused).then(then)

  const moved = (place: Place) => {
    if (moving != null) {
      void api.changePlace(moving.id, place.geoNamesId, refused).then(then)
    }
    setMoving(null)
  }

  const escape = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === 'Escape') {
      onClose()
    }
  }

  const groups = snapshot.choices.filter((choice) => choice.children.length > 0)
  const toggles = snapshot.choices.filter((choice) => choice.children.length === 0)

  return (
    <main className="settings" ref={panel} onKeyDown={escape}>
      <header>
        <h1>Settings</h1>
        <button type="button" onClick={onClose}>
          Close
        </button>
      </header>
      {problem !== '' && (
        <p className="problem" role="alert">
          {problem}
        </p>
      )}

      <h2>Cities</h2>
      <CityList
        cells={snapshot.cells}
        onRename={(id, label) => void api.renameCity(id, label, refused).then(then)}
        onChangePlace={setMoving}
        onRemove={(id) => void api.removeCity(id, refused).then(then)}
      />
      {moving == null ? (
        <PlaceSearch
          key="add"
          heading="Add a city"
          onChoose={added}
          refused={refused}
          autoFocus={startAdding}
          picture={{ art: addCityArt, label: addCityTip }}
        />
      ) : (
        <PlaceSearch
          key={`move-${moving.id}`}
          heading={`Change the place of ${moving.label}`}
          onChoose={moved}
          onCancel={() => setMoving(null)}
          refused={refused}
          autoFocus
        />
      )}

      <div className="settings-choices">
        {choices.map((choice) => (
          <fieldset key={choice.key}>
            <legend>{choice.name}</legend>
            {choice.options.map((option) => (
              <label key={option.value}>
                <input
                  type="radio"
                  name={choice.key}
                  value={option.value}
                  checked={snapshot[choice.key] === option.value}
                  onChange={() => void setters[choice.key](option.value, refused).then(then)}
                />
                {option.label}
              </label>
            ))}
          </fieldset>
        ))}

        {groups.map((choice) => (
          <MenuGroup key={choice.label} choice={choice} choose={choose} />
        ))}

        <OpacitySlider opacity={snapshot.opacity} minOpacity={snapshot.minOpacity} calls={api} refused={refused} then={then} />

        <fieldset>
          <legend>Window</legend>
          {toggles.map((choice) => (
            <MenuToggle key={choice.action} choice={choice} choose={choose} />
          ))}
          <label>
            <input
              type="checkbox"
              disabled={startAtSignIn == null}
              checked={startAtSignIn === true}
              onChange={(event) => {
                const on = event.target.checked
                void api.setStartWithWindows(on, refused).then(() => api.startWithWindows(refused).then(setStartAtSignIn))
              }}
            />
            {snapshot.startLabel}
          </label>
        </fieldset>
      </div>

      <footer className="settings-foot">
        <ArtButton art={donateMark} label={donateTip} onClick={() => void api.openDonation(refused)} />
        <p className="muted">Free to use and staying free: nothing is held back behind a donation.</p>
      </footer>
    </main>
  )
}
