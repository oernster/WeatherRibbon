import type { WeatherSymbol } from './api'

/** Every Yr icon the page ships, by its file's path; Go names each symbol by its icon's file name. */
const files = import.meta.glob<string>('./assets/weather/*.svg', { eager: true, query: '?url', import: 'default' })

/** iconOf answers where the icon named name was bundled; undefined where none was. */
export function iconOf(name: string): string | undefined {
  return files[`./assets/weather/${name}.svg`]
}

interface Props {
  symbol: WeatherSymbol
  /** large draws the current conditions' icon; otherwise a day's or an hour's. */
  large?: boolean
}

/**
 * Symbol draws a weather symbol: its icon where Go found one, else the code's words in its place
 * (FR-412). The icon's alt text is the code's words as Go wrote them, so the weather is never told by
 * picture alone (NFR-U-2).
 */
export function Symbol({ symbol, large = false }: Props) {
  const source = symbol.icon !== '' ? iconOf(symbol.icon) : undefined
  if (source == null) {
    return symbol.words === '' ? null : <span className="symbol-words">{symbol.words}</span>
  }
  return <img className={large ? 'symbol large' : 'symbol'} src={source} alt={symbol.words} title={symbol.words} draggable={false} />
}
