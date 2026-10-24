import { useState, type CSSProperties, type ReactNode } from 'react'
import { Check, Coffee, GlassWater, Power, Scale, Settings, Thermometer, Wind } from 'lucide-react'
import { useSSE } from '@/hooks/useSSE'
import { setMode, setDose, startBackFlush, setPower } from '@/lib/api'
import { LangSwitch as BaseLangSwitch, useI18n, type I18n } from '@/lib/i18n'
import { BoilerInfo, DoseMode, MachineStatus, getDoseWeight } from '@/types/status'
import { SettingsModal } from '@/components/SettingsModal'

const MODES: DoseMode[] = ['Dose1', 'Dose2', 'Continuous']

export function App() {
  const { t } = useI18n()
  const { status, isConnected, reconnect } = useSSE()
  const [showSettings, setShowSettings] = useState(false)

  return (
    <div className="min-h-screen">
      <header className="top-bar">
        <div className="mx-auto flex max-w-2xl items-center gap-3 px-safe py-2.5">
          <span className="ico">
            <Coffee />
          </span>
          <div className="min-w-0 flex-1">
            <div className="truncate text-[17px] leading-tight font-semibold tracking-tight">{t('title')}</div>
            <div className="flex min-w-0 items-center gap-1.5 text-xs" style={{ color: 'var(--ink-3)' }}>
              {isConnected ? (
                <span className="truncate">{status?.model || 'La Marzocco'}</span>
              ) : (
                <button className="inline-flex items-center gap-1.5" onClick={reconnect} title={t('reconnect')}>
                  <span className="dot bad" /> {t('disconnected')}
                </button>
              )}
            </div>
          </div>
          <LangSwitch />
          <button
            className="icon-btn -mr-2"
            onClick={() => setShowSettings(true)}
            aria-label={t('settings')}
            title={t('settings')}
          >
            <Settings className="size-[19px]" />
          </button>
        </div>
      </header>

      <main className="mx-auto max-w-2xl space-y-5 px-safe pt-5 pb-[max(1.5rem,env(safe-area-inset-bottom))]">
        <MachineCard status={status} connected={isConnected} />
        <BrewByWeight status={status} />
        <p className="pt-2 text-center text-xs" style={{ color: 'var(--ink-4)' }}>
          mqtt-lamarzocco{status?.serial ? ` · ${status.serial}` : ''}
        </p>
      </main>

      <SettingsModal
        isOpen={showSettings}
        status={status}
        onSaveDose={setDose}
        onBackFlush={startBackFlush}
        onClose={() => setShowSettings(false)}
      />
    </div>
  )
}

function LangSwitch() {
  return <BaseLangSwitch i18n={useI18n()} />
}

function MachineCard({ status, connected }: { status: MachineStatus | null; connected: boolean }) {
  const i = useI18n()
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const on = status?.machineOn ?? false
  const coffeeReady = status?.boilers?.coffee?.ready ?? false

  // Colour only ever means state.
  const [stateClass, stateLabel] =
    !connected || !status
      ? ['bad', i.t('state.offline')]
      : !on
        ? ['', i.t('state.standby')]
        : coffeeReady
          ? ['ok', i.t('state.ready')]
          : ['warn', i.t('state.heating')]

  const togglePower = async () => {
    if (busy || !status) return
    setBusy(true)
    setError(null)
    try {
      await setPower(!on)
    } catch {
      setError(i.t('power.failed'))
    } finally {
      setTimeout(() => setBusy(false), 1000)
    }
  }

  return (
    <section className="card">
      <div className="card-h">
        <span className="ico sm" style={{ '--c': on ? 'var(--accent)' : 'var(--idle)' } as CSSProperties}>
          <Coffee />
        </span>
        <h3 className="min-w-0 truncate">{status?.model || 'La Marzocco'}</h3>
        <span className={`state ${stateClass} ml-auto`}>
          <span className={`dot ${stateClass ? '' : 'idle'} ${on && !coffeeReady && connected ? 'live' : ''}`} />
          {stateLabel}
        </span>
      </div>

      <div className="card-b space-y-4">
        {status && (
          <div className="stats four">
            <Stat
              icon={<Thermometer />}
              label={i.t('coffeeBoiler')}
              value={status.boilers?.coffee?.temperature ? `${i.num(status.boilers.coffee.temperature)} °C` : '–'}
              hint={<BoilerHint boiler={status.boilers?.coffee} on={on} i={i} />}
            />
            <Stat
              icon={<Wind />}
              label={i.t('steamBoiler')}
              value={steamLevel(status.boilers?.steam?.level, i)}
              hint={<BoilerHint boiler={status.boilers?.steam} on={on} i={i} />}
            />
            <Stat
              icon={<Scale />}
              label={i.t('scale')}
              value={
                status.scale?.connected && status.scale.batteryLevel !== undefined
                  ? `${i.num(status.scale.batteryLevel)} %`
                  : '–'
              }
              hint={
                <>
                  <span className={`dot ${status.scale?.connected ? '' : 'idle'}`} />
                  {status.scale?.connected ? i.t('scale.connected') : i.t('scale.disconnected')}
                </>
              }
            />
            <Stat
              icon={<GlassWater />}
              label={i.t('shots')}
              value={i.num(status.shots ?? 0)}
              hint={i.t('flushes', { n: i.num(status.flushes ?? 0) })}
            />
          </div>
        )}

        {status && !on && (
          <div className="alert-banner warn">
            <span className="ico sm">
              <Power />
            </span>
            <div className="min-w-0">
              <div className="t">{i.t('off.title')}</div>
              <div className="s">{i.t('off.text')}</div>
            </div>
          </div>
        )}

        {error && (
          <p className="text-sm" style={{ color: 'var(--alert)' }}>
            {error}
          </p>
        )}

        {status && (
          <div className="flex justify-end">
            <button className="btn" onClick={togglePower} disabled={busy || !connected}>
              <Power className="size-4" style={{ color: on ? 'var(--ok)' : 'var(--ink-3)' }} />
              {busy ? '…' : on ? i.t('power.off') : i.t('power.on')}
            </button>
          </div>
        )}
      </div>
    </section>
  )
}

function steamLevel(level: string | undefined, i: I18n): string {
  if (!level) return '–'
  const n = level.match(/\d+/)?.[0]
  return n ? i.t('steam.level', { n }) : level
}

function BoilerHint({ boiler, on, i }: { boiler?: BoilerInfo; on: boolean; i: I18n }) {
  if (!boiler || !on) {
    return (
      <>
        <span className="dot idle" /> {i.t('off')}
      </>
    )
  }
  if (boiler.ready) {
    return (
      <>
        <span className="dot" /> {i.t('ready')}
      </>
    )
  }
  const min = boiler.remainingSeconds && boiler.remainingSeconds > 0 ? Math.ceil(boiler.remainingSeconds / 60) : 0
  return (
    <>
      <span className="dot warn" /> {min > 0 ? i.t('heating.min', { n: min }) : i.t('heating')}
    </>
  )
}

function Stat({ icon, label, value, hint }: { icon: ReactNode; label: string; value: string; hint?: ReactNode }) {
  return (
    <div className="stat">
      <div className="k">
        {icon}
        <span className="truncate">{label}</span>
      </div>
      <div className="v">{value}</div>
      {hint && <div className="h">{hint}</div>}
    </div>
  )
}

function BrewByWeight({ status }: { status: MachineStatus | null }) {
  const i = useI18n()
  const [pending, setPending] = useState<DoseMode | null>(null)
  const [error, setError] = useState<string | null>(null)

  const choose = async (mode: DoseMode) => {
    setPending(mode)
    setError(null)
    try {
      await setMode(mode)
    } catch {
      setError(i.t('mode.failed'))
    } finally {
      setTimeout(() => setPending(null), 500)
    }
  }

  return (
    <section>
      <div className="sec-title">
        <h2>{i.t('brewByWeight')}</h2>
      </div>
      <div className="space-y-2.5">
        {MODES.map((mode) => {
          const active = status?.mode === mode
          const weight = getDoseWeight(mode, status ?? undefined)
          return (
            <button
              key={mode}
              className="choice"
              aria-pressed={active}
              disabled={pending === mode}
              onClick={() => choose(mode)}
            >
              <span className="tick">{active && <Check strokeWidth={3} />}</span>
              <span className="min-w-0">
                <span className="name block">{i.t(`mode.${mode}`)}</span>
                <span className="sub block">
                  {weight !== undefined ? i.t('mode.stopsAt', { w: i.grams(weight) }) : i.t('mode.continuous')}
                </span>
              </span>
              {weight !== undefined && (
                <span className="w">
                  {i.grams(weight)}
                  <small>g</small>
                </span>
              )}
            </button>
          )
        })}
      </div>
      {error && (
        <p className="mt-3 text-sm" style={{ color: 'var(--alert)' }}>
          {error}
        </p>
      )}
    </section>
  )
}
