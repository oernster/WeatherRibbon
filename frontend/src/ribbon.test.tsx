import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { cell, installBridge, snapshot } from './fakeBridge'
import { noCities, Ribbon } from './Ribbon'

function ribbon(shown = snapshot(), onOpen = vi.fn(), onAddCity = vi.fn()) {
  installBridge()
  render(<Ribbon snapshot={shown} onAddCity={onAddCity} onOpen={onOpen} reload={vi.fn()} refused={vi.fn()} />)
  return { onOpen, onAddCity }
}

describe('Ribbon', () => {
  it('draws each city with its time, its weather now, today and three days ahead (FR-401 to FR-406)', () => {
    ribbon()
    const london = screen.getByRole('group', { name: 'London, 08:36, 14°C' })
    expect(london.textContent).toContain('08:36 BST')
    expect(london.textContent).toContain('H 16°C L 9°C')
    expect(london.querySelectorAll('.day')).toHaveLength(3)
    expect(screen.getAllByRole('img', { name: 'rain' }).length).toBeGreaterThan(0)
  })

  it('shows a symbol with no icon as its words (FR-412)', () => {
    ribbon(snapshot({ cells: [cell({ symbol: { icon: '', words: 'lightrainshowers day' } })] }))
    expect(screen.getByText('lightrainshowers day')).toBeTruthy()
  })

  it('says why a city cannot show the weather, showing none (FR-305, FR-403, NFR-U-2)', () => {
    ribbon(snapshot({ cells: [cell({ problem: 'Forecast unavailable' })] }))
    expect(screen.getByText('Forecast unavailable')).toBeTruthy()
    expect(screen.queryByText('14°C')).toBeNull()
  })

  it('tells a stale forecast\'s age (FR-305)', () => {
    ribbon(snapshot({ cells: [cell({ age: 'Updated 3 hours ago' })] }))
    expect(screen.getByText('Updated 3 hours ago')).toBeTruthy()
  })

  it('offers Add city on an empty ribbon (FR-106)', () => {
    const { onAddCity } = ribbon(snapshot({ cells: [] }))
    expect(screen.getByText(noCities)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: /Add a city/ }))
    expect(onAddCity).toHaveBeenCalledOnce()
  })

  it('opens a city\'s hourly detail on a click (FR-410)', () => {
    const { onOpen } = ribbon()
    fireEvent.click(screen.getByRole('group', { name: 'London, 08:36, 14°C' }))
    expect(onOpen).toHaveBeenCalledWith('london')
  })
})
