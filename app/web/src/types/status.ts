export type DoseMode = 'Dose1' | 'Dose2' | 'Continuous';

export interface DoseInfo {
  weight: number; // Weight in grams
}

export interface BoilerInfo {
  ready: boolean;
  remainingSeconds?: number; // Seconds until ready (0 if ready)
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
  boiler?: BoilerInfo;
  scale?: ScaleInfo;
}

export function getModeDisplayName(mode: DoseMode): string {
  switch (mode) {
    case 'Dose1':
      return 'Dose 1';
    case 'Dose2':
      return 'Dose 2';
    case 'Continuous':
      return 'Continuous';
    default:
      return mode;
  }
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
