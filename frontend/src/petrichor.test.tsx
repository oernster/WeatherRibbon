import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { petrichorLine } from './Cell'
import { cell, installBridge, snapshot } from './fakeBridge'
import { Ribbon } from './Ribbon'

describe('the petrichor line (FR-413)', () => {
  it('shows in the city where the rain arrived; a click hides it and opens no detail', async () => {
    const bridge = installBridge()
    const onOpen = vi.fn()
    const reload = vi.fn()
    render(
      <Ribbon
        snapshot={snapshot({ cells: [cell({ petrichor: true }), cell({ id: 'tokyo', label: 'Tokyo' })] })}
        onAddCity={vi.fn()}
        onOpen={onOpen}
        reload={reload}
        refused={vi.fn()}
      />,
    )
    const line = screen.getByRole('button', { name: petrichorLine })
    expect(screen.getAllByText(petrichorLine)).toHaveLength(1)
    fireEvent.click(line)
    expect(bridge.DismissPetrichor).toHaveBeenCalledWith('london')
    expect(onOpen).not.toHaveBeenCalled()
    await waitFor(() => expect(reload).toHaveBeenCalledOnce())
  })

})
