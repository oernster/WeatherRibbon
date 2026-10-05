import { describe, expect, it, vi } from 'vitest'
import { cellWidthNeeded } from './measure'

/** Every text jsdom lays out is this wide per character, so widths follow the longest text. */
const perCharacter = 7

describe('cellWidthNeeded (FR-103)', () => {
  it('fits the widest label and time, temperature and outlook, with the cell\'s padding', () => {
    const spy = vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
      const widest = Math.max(0, ...Array.from(this.children, (line) => (line.textContent ?? '').length))
      return { width: widest * perCharacter } as DOMRect
    })
    try {
      const narrow = cellWidthNeeded({ times: ['00:00'], weekdays: ['Sunday'], temperatures: [5], labels: ['Oslo'] })
      const wide = cellWidthNeeded({ times: ['00:00'], weekdays: ['Sunday'], temperatures: [5], labels: ['Llanfairpwllgwyngyll'] })
      const cold = cellWidthNeeded({ times: ['00:00'], weekdays: ['Sunday'], temperatures: [-60, 5], labels: ['Oslo'] })
      expect(wide).toBeGreaterThan(narrow)
      expect(cold).toBeGreaterThan(narrow)
    } finally {
      spy.mockRestore()
    }
  })
})
