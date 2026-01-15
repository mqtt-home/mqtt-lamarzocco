export type DoseMode = 'Dose1' | 'Dose2' | 'Continuous';

export interface MachineStatus {
  mode: DoseMode;
  connected: boolean;
  serial?: string;
  model?: string;
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
