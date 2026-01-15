import { useState } from 'react';
import { Coffee, Sun, Moon, Wifi, WifiOff } from 'lucide-react';
import { useSSE } from '@/hooks/useSSE';
import { setMode } from '@/lib/api';
import { useTheme } from '@/contexts/ThemeContext';
import { DoseMode, getModeDisplayName } from '@/types/status';

export function App() {
  const { status, isConnected, error, reconnect } = useSSE();
  const { theme, toggleTheme } = useTheme();
  const [isLoading, setIsLoading] = useState<DoseMode | null>(null);

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

  const modes: DoseMode[] = ['Dose1', 'Dose2', 'Continuous'];

  return (
    <div className="min-h-screen bg-background p-4 md:p-8">
      <div className="max-w-md mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-3">
            <Coffee className="h-8 w-8 text-primary" />
            <h1 className="text-2xl font-bold text-foreground">La Marzocco</h1>
          </div>
          <div className="flex items-center gap-2">
            {/* Connection status */}
            <div className="p-2" title={isConnected ? 'Connected' : 'Disconnected'}>
              {isConnected ? (
                <Wifi className="h-5 w-5 text-green-500" />
              ) : (
                <WifiOff className="h-5 w-5 text-red-500 cursor-pointer" onClick={reconnect} />
              )}
            </div>
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

        {/* Machine info */}
        {status && (
          <div className="mb-6 p-4 bg-card rounded-lg border border-border">
            <div className="text-sm text-muted-foreground mb-1">
              {status.model || 'La Marzocco'} {status.serial && `(${status.serial})`}
            </div>
            <div className="text-lg font-medium text-foreground">
              Brew by Weight: <span className="text-primary">{getModeDisplayName(status.mode)}</span>
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
                <span className="text-lg font-medium">{getModeDisplayName(mode)}</span>
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
    </div>
  );
}
