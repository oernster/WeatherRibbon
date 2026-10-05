import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { Detail } from './Detail'
import { unitsOf } from './units'
import { detail, installBridge } from './fakeBridge'

describe('Detail (FR-410, FR-411)', () => {
  it('shows each hour with its symbol, temperature, rain and wind in the units, with sunrise and sunset', async () => {
    const bridge = installBridge()
    render(<Detail id="london" units="imperial" onClose={vi.fn()} ready />)
    expect(await screen.findByText('09:00')).toBeTruthy()
    expect(bridge.OpenDetail).toHaveBeenCalledWith('london')
    expect(screen.getByText('Sunrise 07:07')).toBeTruthy()
    expect(screen.getByText(`Wind (${unitsOf('imperial').wind})`)).toBeTruthy()
    expect(screen.getByLabelText('from 225 degrees')).toBeTruthy()
  })

  it('leaves out sun times that could not be had; says when there is no forecast', async () => {
    const bridge = installBridge()
    bridge.OpenDetail.mockImplementation(async () => detail({ sunrise: '', sunset: '', hours: [], problem: 'Forecast unavailable' }))
    render(<Detail id="london" units="metric" onClose={vi.fn()} ready />)
    expect(await screen.findByText('Forecast unavailable')).toBeTruthy()
    expect(screen.queryByText(/Sunrise/)).toBeNull()
  })

  it('returns to the ribbon on Close and on Escape', async () => {
    installBridge()
    const onClose = vi.fn()
    render(<Detail id="london" units="metric" onClose={onClose} ready />)
    await screen.findByText('09:00')
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    fireEvent.keyDown(screen.getByRole('main'), { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(2)
  })
})
