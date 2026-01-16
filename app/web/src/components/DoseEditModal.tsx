import { useState, useEffect, useRef } from 'react';
import { X } from 'lucide-react';

interface DoseEditModalProps {
  isOpen: boolean;
  doseId: 'Dose1' | 'Dose2';
  currentWeight: number;
  onSave: (weight: number) => void;
  onClose: () => void;
}

export function DoseEditModal({ isOpen, doseId, currentWeight, onSave, onClose }: DoseEditModalProps) {
  const [weight, setWeight] = useState(currentWeight);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setWeight(currentWeight);
    setError(null);
  }, [currentWeight, isOpen]);

  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isOpen]);

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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (weight < 5 || weight > 100) {
      setError('Weight must be between 5 and 100 grams');
      return;
    }

    onSave(weight);
  };

  if (!isOpen) return null;

  const displayName = doseId === 'Dose1' ? 'Dose 1' : 'Dose 2';

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
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-foreground">Edit {displayName}</h2>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-accent transition-colors"
          >
            <X className="h-5 w-5 text-muted-foreground" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="weight" className="block text-sm font-medium text-muted-foreground mb-2">
              Weight (grams)
            </label>
            <input
              ref={inputRef}
              id="weight"
              type="number"
              min={5}
              max={100}
              step={0.1}
              value={weight}
              onChange={(e) => {
                setWeight(parseFloat(e.target.value) || 0);
                setError(null);
              }}
              className="w-full px-3 py-2 bg-background border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
            />
            {error && (
              <p className="mt-1 text-sm text-red-500">{error}</p>
            )}
            <p className="mt-1 text-xs text-muted-foreground">
              Min: 5g, Max: 100g
            </p>
          </div>

          {/* Buttons */}
          <div className="flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-border rounded-lg text-foreground hover:bg-accent transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
            >
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
