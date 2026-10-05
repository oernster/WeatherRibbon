import { useCallback, useEffect, useState } from 'react'
import { api, type Snapshot } from './api'
import { About, Licence, Update, useShell, type Panel } from '@oernster/ribbonkit'
import appIcon from './assets/app-icon.png'
import { Detail } from './Detail'
import { useMeasuredCells } from './measure'
import { Ribbon } from './Ribbon'
import { Settings } from './Settings'

/**
 * WeatherRibbon's own open-panel words: add-city, which app.go sends, opens Settings on the place
 * search; detail is WeatherRibbon's own panel, which main.go sizes (FR-410). The window's words are
 * ribbonkit's (useShell).
 */
const addCity = 'add-city'
const detail = 'detail'
const opens: Record<string, Panel | typeof detail> = { [addCity]: 'settings', [detail]: detail }

/**
 * App is ribbonkit's shell around the cities: it shows the ribbon, else the panel Go or the page
 * opened it at. The snapshot is also taken again at each minute boundary Go names (FR-401).
 */
export function App() {
  const { snapshot, problem, refused, view, at, update, opened, load, openPanel, closePanel } = useShell<Snapshot, typeof detail>({
    calls: api,
    take: api.snapshot,
    opens,
  })
  // shown is the city whose hourly detail is open.
  const [shown, setShown] = useState('')

  // Go widens the cells to the widest text the page really draws, which only it can measure.
  useMeasuredCells(snapshot, load, refused)

  useEffect(() => {
    if (snapshot == null) {
      return
    }
    const timer = window.setTimeout(load, snapshot.refreshInMs)
    return () => window.clearTimeout(timer)
  }, [snapshot, load])

  const openDetail = useCallback(
    (id: string) => {
      setShown(id)
      openPanel(detail)
    },
    [openPanel],
  )

  if (snapshot == null) {
    return <div className="problem">{problem}</div>
  }
  if (view === 'settings') {
    return <Settings snapshot={snapshot} startAdding={at === addCity} reload={load} onClose={closePanel} ready={opened} />
  }
  if (view === detail) {
    return <Detail id={shown} units={snapshot.units} onClose={closePanel} ready={opened} />
  }
  if (view === 'about') {
    return <About onClose={closePanel} calls={api} icon={appIcon} />
  }
  if (view === 'licence') {
    return <Licence onClose={closePanel} calls={api} />
  }
  if (view === 'update' && update != null) {
    return <Update status={update} onClose={closePanel} calls={api} />
  }
  return <Ribbon snapshot={snapshot} onAddCity={() => openPanel(addCity)} onOpen={openDetail} reload={load} refused={refused} />
}
