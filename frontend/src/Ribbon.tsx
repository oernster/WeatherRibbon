import type { CSSProperties } from 'react'
import { api, type Refused, type Snapshot } from './api'
import { ArtButton, addCityTip } from './ArtButton'
import addCityArt from './assets/add-city.png'
import { Cell } from './Cell'
import { Band, percentOfWhole } from '@oernster/ribbonkit'

/** The grip's words, one home: what dragging and double-clicking it do (NFR-U-4). */
export const scaleGripTip = 'Drag to resize the cities; double-click for their own size'

/** The empty ribbon's words (FR-106). */
export const noCities = 'No cities yet'

interface Props {
  snapshot: Snapshot
  onAddCity: () => void
  onOpen: (id: string) => void
  /** reload takes the snapshot again, once a notice or a petrichor line is dismissed. */
  reload: () => void
  refused: Refused
}

/** Ribbon is the cities in order (FR-108), in ribbonkit's band: its tab, drag, menu and grip. */
export function Ribbon({ snapshot, onAddCity, onOpen, reload, refused }: Props) {
  const cell = snapshot.cells.length === 0 ? snapshot.layout.prompt : snapshot.layout.cell
  const sizing = {
    '--cell-w': `${cell.width}px`,
    '--cell-h': `${cell.height}px`,
    '--pad': `${snapshot.layout.padding}px`,
    // Everything in the cells is drawn at the chosen scale together; Go has sized the window at the
    // same scale (FR-705).
    zoom: snapshot.scale / percentOfWhole,
  } as CSSProperties

  return (
    <Band
      collapsed={snapshot.collapsed}
      vertical={snapshot.orientation === 'vertical'}
      scrolls={snapshot.scrolls}
      lane=""
      className=""
      style={sizing}
      dragThreshold={snapshot.dragThreshold}
      calls={api}
      refused={refused}
      gripTip={scaleGripTip}
    >
      {snapshot.notices.map((notice) => (
        <div key={notice} className="cell" role="alert">
          <div className="problem">{notice}</div>
          <button type="button" onClick={() => void api.dismissNotices(refused).then(reload)}>
            OK
          </button>
        </div>
      ))}
      {snapshot.cells.length === 0 && (
        <div className="cell prompt">
          <div className="place">{noCities}</div>
          <ArtButton art={addCityArt} label={addCityTip} large onClick={onAddCity} />
        </div>
      )}
      {snapshot.cells.map((each) => (
        <Cell key={each.id} cell={each} onOpen={onOpen} onDismissPetrichor={(id) => void api.dismissPetrichor(id, refused).then(reload)} />
      ))}
    </Band>
  )
}
