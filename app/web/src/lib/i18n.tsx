import { useMemo } from 'react'
import { createI18n, type I18n as BaseI18n } from '@mqtt-home/ui/i18n'
export { LANGS, LangSwitch, resolveText, type Lang, type Text } from '@mqtt-home/ui/i18n'

// German is the house language; English stays one tap away (per device).
const dict = {
  de: {
    title: 'Kaffeemaschine',
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
}

export type Key = keyof (typeof dict)['de']

export type I18n = BaseI18n<Key> & {
  /** Grams with up to one decimal, "16,5". */
  grams: (g: number) => string
}

const base = createI18n(dict)
export const I18nProvider = base.I18nProvider

export function useI18n(): I18n {
  const i = base.useI18n()
  return useMemo<I18n>(() => {
    const gramsFmt = new Intl.NumberFormat(i.locale, { maximumFractionDigits: 1 })
    return { ...i, grams: (g) => gramsFmt.format(g) }
  }, [i])
}
