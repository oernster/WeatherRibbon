import { render } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { cell, installBridge, snapshot } from './fakeBridge'
import { Ribbon } from './Ribbon'

// No condition is told by colour or picture alone: each weather icon carries an accessible name
// holding the code's words Go sends with it; a code with no icon shows the words themselves (NFR-U-2,
// FR-412). How the words are spelled is Go's (TestBothSpellingsOfLightSleetThunderHaveAnIcon).

/** Every icon the page ships, by name: the files Symbol.tsx draws from. */
const shipped = Object.keys(import.meta.glob('./assets/weather/*.svg')).map((path) => path.slice(path.lastIndexOf('/') + 1, -'.svg'.length))

/** wordsOf stands for the words Go sends with an icon, unlike its file's name so the test can tell them apart. */
const wordsOf = (icon: string) => `words for ${icon}`

/** namesOf answers the accessible name of each weather symbol picture in root. */
function namesOf(root: HTMLElement): string[] {
  return [...root.querySelectorAll<HTMLImageElement>('img.symbol')].map((img) => img.getAttribute('alt') ?? '')
}

function drawn(icons: string[]): HTMLElement {
  installBridge()
  const cells = icons.map((icon) => cell({ id: icon, label: icon, symbol: { icon, words: wordsOf(icon) }, outlook: [] }))
  return render(<Ribbon snapshot={snapshot({ cells })} onAddCity={vi.fn()} onOpen={vi.fn()} reload={vi.fn()} refused={vi.fn()} />).container
}

describe('weather symbols (NFR-U-2)', () => {
  it('ships icons to check', () => {
    expect(shipped.length).toBeGreaterThan(0)
  })

  // Each cell draws one picture with no outlook: the weather now (Cell.tsx; today draws none).
  it('names every icon the page ships by the words Go sends with it', () => {
    expect(namesOf(drawn(shipped)).sort()).toEqual(shipped.map(wordsOf).sort())
  })

  it('shows a code with no icon as its words, as text', () => {
    installBridge()
    const { container } = render(
      <Ribbon snapshot={snapshot({ cells: [cell({ symbol: { icon: '', words: 'lightrainshowers day' }, outlook: [] })] })} onAddCity={vi.fn()} onOpen={vi.fn()} reload={vi.fn()} refused={vi.fn()} />,
    )
    expect(container.textContent).toContain('lightrainshowers day')
  })

  it('finds a symbol picture with no name, so it can catch one', () => {
    const { container } = render(<img className="symbol" src="x.svg" alt="" />)
    expect(namesOf(container)).toEqual([''])
  })
})
