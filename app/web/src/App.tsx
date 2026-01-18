import { useState } from 'react';
import { Coffee, Sun, Moon, Wifi, WifiOff, Settings, Power, PowerOff, Thermometer, Battery, Scale } from 'lucide-react';
import { useSSE } from '@/hooks/useSSE';
import { setMode, setDose, startBackFlush, setPower } from '@/lib/api';
import { useTheme } from '@/contexts/ThemeContext';
import { DoseMode, getModeDisplayName, getDoseWeight } from '@/types/status';
import { SettingsModal } from '@/components/SettingsModal';

export function App() {
  const { status, isConnected, error, reconnect } = useSSE();
  const { theme, toggleTheme } = useTheme();
  const [isLoading, setIsLoading] = useState<DoseMode | null>(null);
  const [showSettings, setShowSettings] = useState(false);
  const [powerLoading, setPowerLoading] = useState(false);

  const handleSetMode = async (mode: DoseMode) => {
    setIsLoading(mode);
    try {
      await setMode(mode);
    } catch (err) {
      console.error('Failed to set mode:', err);
    } finally {
      setTimeout(() => setIsLoading(null), 500);
    }
  };

  const handleSaveDose = async (doseId: 'Dose1' | 'Dose2', weight: number) => {
    await setDose(doseId, weight);
  };

  const handleTogglePower = async () => {
    if (powerLoading || !status) return;
    setPowerLoading(true);
    try {
      await setPower(!machineOn);
    } catch (err) {
      console.error('Failed to toggle power:', err);
    } finally {
      setTimeout(() => setPowerLoading(false), 1000);
    }
  };

  const modes: DoseMode[] = ['Dose1', 'Dose2', 'Continuous'];

  const machineOn = status?.machineOn ?? false;

  return (
    <div className="min-h-screen bg-background p-4 md:p-8">
      <div className="max-w-md mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-3">
            <Coffee className="h-8 w-8 text-primary" />
            <h1 className="text-2xl font-bold text-foreground">La Marzocco</h1>
          </div>
          <div className="flex items-center gap-1">
            {/* Connection status */}
            <div className="p-2" title={isConnected ? 'Connected' : 'Disconnected'}>
              {isConnected ? (
                <Wifi className="h-5 w-5 text-green-500" />
              ) : (
                <WifiOff className="h-5 w-5 text-red-500 cursor-pointer" onClick={reconnect} />
              )}
            </div>
            {/* Settings button */}
            <button
              onClick={() => setShowSettings(true)}
              className="p-2 rounded-lg hover:bg-accent transition-colors"
              aria-label="Settings"
              title="Settings"
            >
              <Settings className="h-5 w-5 text-foreground" />
            </button>
            {/* Theme toggle */}
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg hover:bg-accent transition-colors"
              aria-label="Toggle theme"
            >
              {theme === 'dark' ? (
                <Sun className="h-5 w-5 text-foreground" />
              ) : (
                <Moon className="h-5 w-5 text-foreground" />
              )}
            </button>
          </div>
        </div>

        {/* Error message */}
        {error && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-500 text-sm">
            {error}
            <button
              onClick={reconnect}
              className="ml-2 underline hover:no-underline"
            >
              Retry
            </button>
          </div>
        )}

        {/* Machine status banner */}
        {status && !machineOn && (
          <div className="mb-4 p-4 bg-amber-500/10 border border-amber-500/30 rounded-lg flex items-center gap-3">
            <PowerOff className="h-5 w-5 text-amber-500 flex-shrink-0" />
            <div>
              <div className="font-medium text-amber-600 dark:text-amber-400">Machine is off</div>
              <div className="text-sm text-amber-600/80 dark:text-amber-400/80">
                Turn on the machine to control brew settings
              </div>
            </div>
          </div>
        )}

        {/* Machine info */}
        {status && (
          <div className="mb-6 p-4 bg-card rounded-lg border border-border">
            <div className="flex items-center justify-between mb-1">
              <div className="text-sm text-muted-foreground">
                {status.model || 'La Marzocco'} {status.serial && `(${status.serial})`}
              </div>
              <button
                onClick={handleTogglePower}
                disabled={powerLoading}
                className={`flex items-center gap-1.5 text-xs font-medium px-2 py-1 rounded transition-colors ${
                  machineOn
                    ? 'text-green-500 hover:bg-green-500/10'
                    : 'text-muted-foreground hover:bg-accent'
                } ${powerLoading ? 'opacity-50 cursor-wait' : 'cursor-pointer'}`}
                title={machineOn ? 'Click to turn off' : 'Click to turn on'}
              >
                {machineOn ? (
                  <>
                    <Power className="h-3.5 w-3.5" />
                    <span>{powerLoading ? '...' : 'On'}</span>
                  </>
                ) : (
                  <>
                    <PowerOff className="h-3.5 w-3.5" />
                    <span>{powerLoading ? '...' : 'Off'}</span>
                  </>
                )}
              </button>
            </div>
            <div className="text-lg font-medium text-foreground">
              Brew by Weight: <span className="text-primary">{getModeDisplayName(status.mode)}</span>
            </div>

            {/* Boilers and Scale status */}
            <div className="mt-3 pt-3 border-t border-border flex flex-wrap gap-4 text-sm">
              {status.boilers?.coffee && (
                <div className="flex items-center gap-2">
                  <Thermometer className={`h-4 w-4 ${status.boilers.coffee.ready ? 'text-green-500' : 'text-amber-500'}`} />
                  <span className="text-muted-foreground">
                    Coffee
                    {status.boilers.coffee.temperature ? ` ${status.boilers.coffee.temperature}°C` : ''}
                    {!status.boilers.coffee.ready && (
                      <>
                        {status.boilers.coffee.remainingSeconds !== undefined && status.boilers.coffee.remainingSeconds > 0 ? (
                          <span className="ml-1 tabular-nums">
                            ({Math.ceil(status.boilers.coffee.remainingSeconds / 60)}m)
                          </span>
                        ) : (
                          <span className="ml-1">(heating)</span>
                        )}
                      </>
                    )}
                  </span>
                </div>
              )}
              {status.boilers?.steam && (
                <div className="flex items-center gap-2">
                  <Thermometer className={`h-4 w-4 ${status.boilers.steam.ready ? 'text-green-500' : 'text-amber-500'}`} />
                  <span className="text-muted-foreground">
                    Steam
                    {status.boilers.steam.level && ` ${status.boilers.steam.level.replace('Level', 'L')}`}
                    {!status.boilers.steam.ready && (
                      <>
                        {status.boilers.steam.remainingSeconds !== undefined && status.boilers.steam.remainingSeconds > 0 ? (
                          <span className="ml-1 tabular-nums">
                            ({Math.ceil(status.boilers.steam.remainingSeconds / 60)}m)
                          </span>
                        ) : (
                          <span className="ml-1">(heating)</span>
                        )}
                      </>
                    )}
                  </span>
                </div>
              )}
              {status.scale && (
                <div className="flex items-center gap-2">
                  <Scale className={`h-4 w-4 ${status.scale.connected ? 'text-foreground' : 'text-muted-foreground'}`} />
                  {status.scale.connected ? (
                    <div className="flex items-center gap-1.5 text-muted-foreground">
                      <span>Scale</span>
                      {status.scale.batteryLevel !== undefined && (
                        <span className="flex items-center gap-1">
                          <Battery className={`h-3.5 w-3.5 ${
                            status.scale.batteryLevel > 50 ? 'text-green-500' :
                            status.scale.batteryLevel > 20 ? 'text-amber-500' : 'text-red-500'
                          }`} />
                          <span className="tabular-nums">{status.scale.batteryLevel}%</span>
                        </span>
                      )}
                    </div>
                  ) : (
                    <span className="text-muted-foreground">Scale disconnected</span>
                  )}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Mode buttons */}
        <div className="space-y-3">
          <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-4">
            Select Mode
          </h2>
          {modes.map((mode) => {
            const isActive = status?.mode === mode;
            const isLoadingThis = isLoading === mode;
            const doseWeight = getDoseWeight(mode, status ?? undefined);

            return (
              <button
                key={mode}
                onClick={() => handleSetMode(mode)}
                disabled={isLoadingThis}
                className={`
                  w-full p-4 rounded-lg border-2 transition-all touch-target
                  flex items-center justify-between
                  ${isActive
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-border bg-card text-foreground hover:border-primary/50 hover:bg-accent'
                  }
                  ${isLoadingThis ? 'opacity-50 cursor-wait' : 'cursor-pointer'}
                `}
              >
                <div className="flex items-baseline gap-3">
                  <span className="text-lg font-medium">{getModeDisplayName(mode)}</span>
                  {doseWeight !== undefined && (
                    <span className="text-sm text-muted-foreground tabular-nums">{doseWeight}g</span>
                  )}
                </div>
                {isActive && (
                  <span className="text-xs uppercase tracking-wide bg-primary text-primary-foreground px-2 py-1 rounded">
                    Active
                  </span>
                )}
                {isLoadingThis && (
                  <span className="text-xs uppercase tracking-wide text-muted-foreground">
                    Setting...
                  </span>
                )}
              </button>
            );
          })}
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-xs text-muted-foreground">
          mqtt-lamarzocco
        </div>
      </div>

      {/* Settings Modal */}
      <SettingsModal
        isOpen={showSettings}
        status={status}
        onSaveDose={handleSaveDose}
        onBackFlush={startBackFlush}
        onClose={() => setShowSettings(false)}
      />
    </div>
  );
}
