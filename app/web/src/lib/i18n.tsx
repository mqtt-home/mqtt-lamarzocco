import { createContext, useCallback, useContext, useMemo, useState } from 'react'

export type Lang = 'de' | 'en'
export const LANGS: Lang[] = ['de', 'en']

const LANG_KEY = 'home.lang'

// German is the house language; English stays one tap away. Per device.
function initialLang(): Lang {
  try {
    const v = localStorage.getItem(LANG_KEY)
    if (v === 'de' || v === 'en') return v
  } catch {
    // storage blocked
  }
  return 'de'
}

const dict = {
  de: {
    title: 'Kaffeemaschine',
    language: 'Sprache',
    settings: 'Einstellungen',
    close: 'Schließen',
    disconnected: 'Keine Verbindung',
    reconnect: 'Neu verbinden',
    'state.ready': 'Bereit',
    'state.heating': 'Heizt auf',
    'state.standby': 'Standby',
    'state.offline': 'Offline',
    'power.on': 'Einschalten',
    'power.off': 'Ausschalten',
    'power.failed': 'Die Kaffeemaschine konnte nicht geschaltet werden.',
    'off.title': 'Die Kaffeemaschine ist aus',
    'off.text': 'Zum Brühen und für die Einstellungen einschalten.',
    coffeeBoiler: 'Kaffeekessel',
    steamBoiler: 'Dampfkessel',
    scale: 'Waage',
    shots: 'Bezüge',
    flushes: '{n} Spülungen',
    ready: 'bereit',
    heating: 'heizt',
    'heating.min': 'heizt, noch {n} min',
    off: 'aus',
    'steam.level': 'Stufe {n}',
    'scale.connected': 'verbunden',
    'scale.disconnected': 'nicht verbunden',
    battery: 'Akku {pct} %',
    brewByWeight: 'Brühen nach Gewicht',
    'mode.Dose1': 'Dosis 1',
    'mode.Dose2': 'Dosis 2',
    'mode.Continuous': 'Durchlauf',
    'mode.stopsAt': 'Stoppt bei {w} g',
    'mode.continuous': 'Ohne Gewichtsstopp',
    'mode.failed': 'Der Modus konnte nicht gesetzt werden.',
    active: 'Aktiv',
    doses: 'Dosen',
    'dose.range': 'Zwischen 5 und 100 g.',
    'dose.invalid': 'Das Gewicht muss zwischen 5 und 100 g liegen.',
    'dose.failed': 'Die Dosis konnte nicht gespeichert werden.',
    save: 'Speichern',
    saved: 'Gespeichert',
    maintenance: 'Wartung',
    'backflush.start': 'Rückspülen starten',
    'backflush.hint': 'Vorher das Blindsieb in den Siebträger einsetzen.',
    'backflush.active': 'Hebel innerhalb von 15 Sekunden auf EIN stellen',
    'backflush.activeHint': 'Das Rückspülen beginnt, sobald der Hebel umgelegt ist.',
    'backflush.failed': 'Das Rückspülen konnte nicht gestartet werden.',
    'settings.off': 'Die Kaffeemaschine ist aus. Zum Ändern der Einstellungen einschalten.',
  },
  en: {
    title: 'Coffee machine',
    language: 'Language',
    settings: 'Settings',
    close: 'Close',
    disconnected: 'Disconnected',
    reconnect: 'Reconnect',
    'state.ready': 'Ready',
    'state.heating': 'Heating',
    'state.standby': 'Standby',
    'state.offline': 'Offline',
    'power.on': 'Turn on',
    'power.off': 'Turn off',
    'power.failed': 'Could not switch the coffee machine.',
    'off.title': 'The coffee machine is off',
    'off.text': 'Turn it on to brew and to change settings.',
    coffeeBoiler: 'Coffee boiler',
    steamBoiler: 'Steam boiler',
    scale: 'Scale',
    shots: 'Shots',
    flushes: '{n} flushes',
    ready: 'ready',
    heating: 'heating',
    'heating.min': 'heating, {n} min left',
    off: 'off',
    'steam.level': 'Level {n}',
    'scale.connected': 'connected',
    'scale.disconnected': 'not connected',
    battery: 'Battery {pct} %',
    brewByWeight: 'Brew by weight',
    'mode.Dose1': 'Dose 1',
    'mode.Dose2': 'Dose 2',
    'mode.Continuous': 'Continuous',
    'mode.stopsAt': 'Stops at {w} g',
    'mode.continuous': 'No weight stop',
    'mode.failed': 'Could not set the mode.',
    active: 'Active',
    doses: 'Doses',
    'dose.range': 'Between 5 and 100 g.',
    'dose.invalid': 'The weight must be between 5 and 100 g.',
    'dose.failed': 'Could not save the dose.',
    save: 'Save',
    saved: 'Saved',
    maintenance: 'Maintenance',
    'backflush.start': 'Start back flush',
    'backflush.hint': 'Insert the blind filter into the portafilter first.',
    'backflush.active': 'Move the paddle to ON within 15 seconds',
    'backflush.activeHint': 'The back flush starts as soon as the paddle is engaged.',
    'backflush.failed': 'Could not start the back flush.',
    'settings.off': 'The coffee machine is off. Turn it on to change settings.',
  },
} satisfies Record<Lang, Record<string, string>>

export type Key = keyof (typeof dict)['de']

export interface I18n {
  lang: Lang
  locale: string
  setLang: (l: Lang) => void
  /** A UI string with {placeholders}. */
  t: (key: Key, vars?: Record<string, string | number>) => string
  /** A number in the current locale (8,5 in German). */
  num: (n: number, decimals?: number) => string
  /** Grams with up to one decimal, "16,5". */
  grams: (g: number) => string
}

const Ctx = createContext<I18n | null>(null)

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [lang, setLangState] = useState<Lang>(initialLang)

  const setLang = useCallback((l: Lang) => {
    setLangState(l)
    try {
      localStorage.setItem(LANG_KEY, l)
    } catch {
      // convenience only
    }
  }, [])

  const value = useMemo<I18n>(() => {
    const locale = lang === 'de' ? 'de-DE' : 'en-GB'
    document.documentElement.lang = lang
    const t: I18n['t'] = (key, vars) =>
      (dict[lang][key] ?? dict.en[key] ?? key).replace(/\{(\w+)\}/g, (_, k) => String(vars?.[k] ?? ''))
    const fmt = new Map<number, Intl.NumberFormat>()
    const num: I18n['num'] = (n, decimals = 0) => {
      let f = fmt.get(decimals)
      if (!f) {
        f = new Intl.NumberFormat(locale, { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
        fmt.set(decimals, f)
      }
      return f.format(n)
    }
    const gramsFmt = new Intl.NumberFormat(locale, { maximumFractionDigits: 1 })
    const grams: I18n['grams'] = (g) => gramsFmt.format(g)
    return { lang, locale, setLang, t, num, grams }
  }, [lang, setLang])

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

export function useI18n() {
  const v = useContext(Ctx)
  if (!v) throw new Error('useI18n outside I18nProvider')
  return v
}
