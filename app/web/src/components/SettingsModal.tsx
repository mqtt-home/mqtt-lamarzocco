import { useState, useEffect } from 'react';
import { X, Droplets } from 'lucide-react';
import { MachineStatus } from '@/types/status';

interface SettingsModalProps {
  isOpen: boolean;
  status: MachineStatus | null;
  onSaveDose: (doseId: 'Dose1' | 'Dose2', weight: number) => Promise<void>;
  onBackFlush: () => Promise<void>;
  onClose: () => void;
}

export function SettingsModal({ isOpen, status, onSaveDose, onBackFlush, onClose }: SettingsModalProps) {
  const machineOn = status?.machineOn ?? false;
  const [dose1, setDose1] = useState(status?.dose1?.weight ?? 0);
  const [dose2, setDose2] = useState(status?.dose2?.weight ?? 0);
  const [saving, setSaving] = useState<'dose1' | 'dose2' | null>(null);
  const [backflushing, setBackflushing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (status) {
      setDose1(status.dose1?.weight ?? 0);
      setDose2(status.dose2?.weight ?? 0);
    }
    setError(null);
  }, [status, isOpen]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return;
      if (e.key === 'Escape') {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  const handleSave = async (doseId: 'Dose1' | 'Dose2') => {
    const weight = doseId === 'Dose1' ? dose1 : dose2;

    if (weight < 5 || weight > 100) {
      setError('Weight must be between 5 and 100 grams');
      return;
    }

    setSaving(doseId === 'Dose1' ? 'dose1' : 'dose2');
    setError(null);

    try {
      await onSaveDose(doseId, weight);
    } catch (err) {
      setError('Failed to save dose weight');
    } finally {
      setSaving(null);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative bg-card border border-border rounded-lg shadow-lg w-full max-w-sm mx-4 p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold text-foreground">Settings</h2>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-accent transition-colors"
          >
            <X className="h-5 w-5 text-muted-foreground" />
          </button>
        </div>

        {/* Error */}
        {error && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-500 text-sm">
            {error}
          </div>
        )}

        {/* Machine off warning */}
        {!machineOn && (
          <div className="mb-4 p-3 bg-amber-500/10 border border-amber-500/20 rounded-lg text-amber-600 dark:text-amber-400 text-sm">
            Machine is off. Turn on the machine to change settings.
          </div>
        )}

        {/* Dose Settings */}
        <div className={`space-y-4 ${!machineOn ? 'opacity-50 pointer-events-none' : ''}`}>
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            Dose Weights
          </h3>

          {/* Dose 1 */}
          <div className="flex items-center gap-3">
            <label htmlFor="dose1" className="w-20 text-sm font-medium text-foreground">
              Dose 1
            </label>
            <div className="flex-1 flex items-center gap-2">
              <input
                id="dose1"
                type="number"
                min={5}
                max={100}
                step={0.1}
                value={dose1}
                onChange={(e) => setDose1(parseFloat(e.target.value) || 0)}
                className="flex-1 px-3 py-2 bg-background border border-border rounded-lg text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
              />
              <span className="text-sm text-muted-foreground">g</span>
              <button
                onClick={() => handleSave('Dose1')}
                disabled={saving === 'dose1'}
                className="px-3 py-2 bg-primary text-primary-foreground rounded-lg text-sm hover:bg-primary/90 transition-colors disabled:opacity-50"
              >
                {saving === 'dose1' ? '...' : 'Save'}
              </button>
            </div>
          </div>

          {/* Dose 2 */}
          <div className="flex items-center gap-3">
            <label htmlFor="dose2" className="w-20 text-sm font-medium text-foreground">
              Dose 2
            </label>
            <div className="flex-1 flex items-center gap-2">
              <input
                id="dose2"
                type="number"
                min={5}
                max={100}
                step={0.1}
                value={dose2}
                onChange={(e) => setDose2(parseFloat(e.target.value) || 0)}
                className="flex-1 px-3 py-2 bg-background border border-border rounded-lg text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
              />
              <span className="text-sm text-muted-foreground">g</span>
              <button
                onClick={() => handleSave('Dose2')}
                disabled={saving === 'dose2'}
                className="px-3 py-2 bg-primary text-primary-foreground rounded-lg text-sm hover:bg-primary/90 transition-colors disabled:opacity-50"
              >
                {saving === 'dose2' ? '...' : 'Save'}
              </button>
            </div>
          </div>

          <p className="text-xs text-muted-foreground">
            Min: 5g, Max: 100g
          </p>
        </div>

        {/* Back Flush */}
        <div className={`mt-6 pt-4 border-t border-border ${!machineOn ? 'opacity-50 pointer-events-none' : ''}`}>
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-3">
            Maintenance
          </h3>
          {backflushing ? (
            <div className="p-4 bg-amber-500/20 border border-amber-500/40 rounded-lg">
              <div className="flex items-center gap-2 text-amber-600 dark:text-amber-400 font-medium">
                <Droplets className="h-5 w-5 animate-pulse" />
                Move paddle to ON within 15 seconds
              </div>
              <p className="mt-2 text-sm text-amber-600/80 dark:text-amber-400/80">
                The back flush cycle will start when you engage the paddle
              </p>
            </div>
          ) : (
            <>
              <button
                onClick={async () => {
                  setBackflushing(true);
                  setError(null);
                  try {
                    await onBackFlush();
                    // Reset after 15 seconds (paddle engagement window)
                    setTimeout(() => setBackflushing(false), 15000);
                  } catch (err) {
                    setError('Failed to start back flush');
                    setBackflushing(false);
                  }
                }}
                className="flex items-center gap-2 px-4 py-2 bg-amber-500/10 border border-amber-500/30 text-amber-600 dark:text-amber-400 rounded-lg hover:bg-amber-500/20 transition-colors"
              >
                <Droplets className="h-4 w-4" />
                Start Back Flush
              </button>
              <p className="mt-2 text-xs text-muted-foreground">
                Insert blind filter before starting
              </p>
            </>
          )}
        </div>

        {/* Close button */}
        <div className="mt-6">
          <button
            onClick={onClose}
            className="w-full px-4 py-2 border border-border rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
