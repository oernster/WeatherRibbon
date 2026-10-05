import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { installBridge, places, snapshot } from './fakeBridge'
import { donateTip, Settings } from './Settings'

function settings() {
  const bridge = installBridge()
  const reload = vi.fn()
  render(<Settings snapshot={snapshot()} startAdding={false} reload={reload} onClose={vi.fn()} />)
  return { bridge, reload }
}

describe('Settings (FR-701, FR-702)', () => {
  it('adds the city chosen in the search by its GeoNames id, then redraws (FR-201, FR-202)', async () => {
    const { bridge, reload } = settings()
    fireEvent.change(screen.getByRole('searchbox', { name: 'Search cities' }), { target: { value: 'lon' } })
    fireEvent.click(await screen.findByText(places[1].description))
    await waitFor(() => expect(bridge.AddCity).toHaveBeenCalledWith(places[1].geoNamesId))
    await waitFor(() => expect(reload).toHaveBeenCalled())
  })

  it('greys a Position choice that would leave the ribbon where it stands, so pressing it does nothing (FR-505)', () => {
    const bridge = installBridge()
    const shown = snapshot()
    shown.choices = shown.choices.map((choice) =>
      choice.label !== 'Position' ? choice : { ...choice, children: choice.children.map((item) => ({ ...item, disabled: item.action === 'right-edge' })) },
    )
    render(<Settings snapshot={shown} startAdding={false} reload={vi.fn()} onClose={vi.fn()} />)
    const right = screen.getByRole('button', { name: 'Centre on right edge' }) as HTMLButtonElement
    expect(right.disabled).toBe(true)
    expect((screen.getByRole('button', { name: 'Centre on left edge' }) as HTMLButtonElement).disabled).toBe(false)
    fireEvent.click(right)
    expect(bridge.Choose).not.toHaveBeenCalled()
  })

  it('lists nothing until something is typed', () => {
    settings()
    expect(screen.queryAllByRole('option')).toHaveLength(0)
  })

  it('asks before removing a city, naming it (FR-205)', async () => {
    const { bridge } = settings()
    fireEvent.click(screen.getByRole('button', { name: 'Remove London' }))
    expect(screen.getByRole('alertdialog', { name: 'Remove city' }).textContent).toContain('Remove London?')
    fireEvent.click(screen.getAllByRole('button', { name: 'Remove' })[0])
    await waitFor(() => expect(bridge.RemoveCity).toHaveBeenCalledWith('london'))
  })

  it('offers the menus\' choices (Units among them) making each through Choose (FR-703)', async () => {
    const { bridge } = settings()
    fireEvent.click(screen.getByRole('radio', { name: 'Imperial' }))
    await waitFor(() => expect(bridge.Choose).toHaveBeenCalledWith('units-imperial'))
  })

  it('sets the time format at once (FR-401, FR-702)', async () => {
    const { bridge } = settings()
    fireEvent.click(screen.getByRole('radio', { name: '12-hour' }))
    await waitFor(() => expect(bridge.SetFormat).toHaveBeenCalledWith('12h'))
  })

  it('hands the donation to the browser from the foot (FR-709)', async () => {
    const { bridge } = settings()
    fireEvent.click(screen.getByRole('button', { name: donateTip }))
    await waitFor(() => expect(bridge.OpenDonation).toHaveBeenCalledOnce())
  })
})
