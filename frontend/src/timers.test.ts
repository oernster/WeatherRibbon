// While the ribbon is shown its page schedules no periodic timer more frequent than once a minute
// (NFR-P-3). The scan and the kit's own timer calls are ribbonkit's; WeatherRibbon names its
// sources, its test support and its own calls.

import { describePageTimers } from '@oernster/ribbonkit/testing'

// The sources are read as text, so the scan sees exactly the files the build bundles. The kit is
// globbed on its own: given beside './**', Vite's glob leaves out everything under node_modules.
const ownSources = import.meta.glob<string>('./**/*.{ts,tsx}', { query: '?raw', import: 'default', eager: true })
const kitSources = import.meta.glob<string>(
  ['../node_modules/@oernster/ribbonkit/web/**/*.{ts,tsx}', '../node_modules/@oernster/ribbonkit/installer/page/*.js'],
  { query: '?raw', import: 'default', eager: true },
)

describePageTimers({
  requirement: 'NFR-P-3',
  sources: { ...ownSources, ...kitSources },
  kitRoot: '../node_modules/@oernster/ribbonkit',
  testSupport: {
    'fakeBridge.ts': 'stands in for Go in the suites',
    'test-setup.ts': "Vitest's setup file",
  },
  allowed: [
    {
      file: 'App.tsx',
      api: 'setTimeout',
      delay: 'snapshot.refreshInMs',
      reason: "one shot to Go's next minute boundary, armed again only by the snapshot it loads (FR-401)",
    },
  ],
})
