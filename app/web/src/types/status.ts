export type DoseMode = 'Dose1' | 'Dose2' | 'Continuous';

export interface DoseInfo {
  weight: number; // Weight in grams
}

export interface BoilerInfo {
  ready: boolean;
  remainingSeconds?: number; // Seconds until ready (0 if ready)
  temperature?: number; // Target temperature (coffee)
  level?: string; // Target level (steam): Level1, Level2, etc.
}

export interface BoilersInfo {
  coffee?: BoilerInfo;
  steam?: BoilerInfo;
}

export interface ScaleInfo {
  connected: boolean;
  batteryLevel?: number; // Battery percentage 0-100
}

export interface MachineStatus {
  mode: DoseMode;
  connected: boolean;
  serial?: string;
  model?: string;
  dose1?: DoseInfo;
  dose2?: DoseInfo;
  machineOn?: boolean;
  boilers?: BoilersInfo;
  scale?: ScaleInfo;
  shots: number;
  flushes: number;
}

export function getDoseWeight(mode: DoseMode, status?: MachineStatus): number | undefined {
  if (mode === 'Dose1' && status?.dose1?.weight) {
    return status.dose1.weight;
  }
  if (mode === 'Dose2' && status?.dose2?.weight) {
    return status.dose2.weight;
  }
  return undefined;
}
