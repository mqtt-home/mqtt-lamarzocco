import { useState, useEffect, useRef } from 'react'
import { Check, Droplets, X } from 'lucide-react'
import { MachineStatus } from '@/types/status'
import { useI18n } from '@/lib/i18n'

type DoseId = 'Dose1' | 'Dose2'

interface SettingsModalProps {
  isOpen: boolean
  status: MachineStatus | null
  onSaveDose: (doseId: DoseId, weight: number) => Promise<void>
  onBackFlush: () => Promise<void>
  onClose: () => void
}

export function SettingsModal({ isOpen, status, onSaveDose, onBackFlush, onClose }: SettingsModalProps) {
  const i = useI18n()
  const machineOn = status?.machineOn ?? false
  const [values, setValues] = useState<Record<DoseId, string>>({ Dose1: '', Dose2: '' })
  const [saving, setSaving] = useState<DoseId | null>(null)
  const [saved, setSaved] = useState<DoseId | null>(null)
  const [backflushing, setBackflushing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const wasOpen = useRef(false)

  // Only take the values from the server when the sheet opens, not on every status update.
  useEffect(() => {
    if (isOpen && !wasOpen.current && status) {
      const fmt = (w?: number) => (w === undefined ? '' : i.grams(w))
      setValues({ Dose1: fmt(status.dose1?.weight), Dose2: fmt(status.dose2?.weight) })
      setError(null)
    }
    wasOpen.current = isOpen
  }, [isOpen, status, i])

  useEffect(() => {
    if (!isOpen) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [isOpen, onClose])

  const handleSave = async (doseId: DoseId) => {
    // Accept both decimal comma and point.
    const weight = parseFloat(values[doseId].replace(',', '.'))
    if (isNaN(weight) || weight < 5 || weight > 100) {
      setError(i.t('dose.invalid'))
      return
    }
    setSaving(doseId)
    setError(null)
    try {
      await onSaveDose(doseId, weight)
      setSaved(doseId)
      setTimeout(() => setSaved(null), 1500)
    } catch {
      setError(i.t('dose.failed'))
    } finally {
      setSaving(null)
    }
  }

  const handleBackFlush = async () => {
    setBackflushing(true)
    setError(null)
    try {
      await onBackFlush()
      // The paddle has to be engaged within 15 seconds.
      setTimeout(() => setBackflushing(false), 15000)
    } catch {
      setError(i.t('backflush.failed'))
      setBackflushing(false)
    }
  }

  if (!isOpen) return null

  const off = !machineOn ? 'opacity-50 pointer-events-none' : ''

  return (
    <div className="sheet-wrap" role="dialog" aria-modal="true" aria-label={i.t('settings')}>
      <div className="sheet-backdrop" onClick={onClose} />
      <div className="sheet">
        <div className="sheet-h">
          <h2>{i.t('settings')}</h2>
          <button className="icon-btn ml-auto" onClick={onClose} aria-label={i.t('close')} title={i.t('close')}>
            <X className="size-5" />
          </button>
        </div>

        {(error || !machineOn) && (
          <div className="space-y-2 px-[18px] pt-3">
            {error && (
              <p className="text-sm" style={{ color: 'var(--alert)' }}>
                {error}
              </p>
            )}
            {!machineOn && (
              <p className="text-sm" style={{ color: 'var(--warn)' }}>
                {i.t('settings.off')}
              </p>
            )}
          </div>
        )}

        <div className={`sheet-sec ${off}`}>
          <div className="kicker">{i.t('doses')}</div>
          <div className="space-y-2.5">
            {(['Dose1', 'Dose2'] as DoseId[]).map((id) => (
              <div key={id} className="flex items-center gap-3">
                <label htmlFor={id} className="w-16 shrink-0 text-sm font-medium">
                  {i.t(`mode.${id}`)}
                </label>
                <div className="field flex-1">
                  <input
                    id={id}
                    type="text"
                    inputMode="decimal"
                    className="num"
                    value={values[id]}
                    onChange={(e) => setValues((v) => ({ ...v, [id]: e.target.value }))}
                    onKeyDown={(e) => e.key === 'Enter' && handleSave(id)}
                  />
                  <span className="unit">g</span>
                </div>
                <button
                  className="btn sm w-[104px]"
                  onClick={() => handleSave(id)}
                  disabled={saving === id}
                  style={saved === id ? { color: 'var(--ok)' } : undefined}
                >
                  {saving === id ? '…' : saved === id ? <Check className="size-4" /> : i.t('save')}
                </button>
              </div>
            ))}
          </div>
          <p className="mt-2 text-xs" style={{ color: 'var(--ink-3)' }}>
            {i.t('dose.range')}
          </p>
        </div>

        <div className={`sheet-sec ${off}`}>
          <div className="kicker">{i.t('maintenance')}</div>
          {backflushing ? (
            <div className="alert-banner warn">
              <span className="ico sm">
                <Droplets className="animate-pulse" />
              </span>
              <div className="min-w-0">
                <div className="t">{i.t('backflush.active')}</div>
                <div className="s">{i.t('backflush.activeHint')}</div>
              </div>
            </div>
          ) : (
            <>
              <button className="btn" onClick={handleBackFlush}>
                <Droplets className="size-4" />
                {i.t('backflush.start')}
              </button>
              <p className="mt-2 text-xs" style={{ color: 'var(--ink-3)' }}>
                {i.t('backflush.hint')}
              </p>
            </>
          )}
        </div>

        <div className="sheet-sec">
          <button className="btn w-full" onClick={onClose}>
            {i.t('close')}
          </button>
        </div>
      </div>
    </div>
  )
}
