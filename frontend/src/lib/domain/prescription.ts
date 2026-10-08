export interface PrescribedItem {
  medicine: string;
  concentration: string;
  presentation: string;
  quantity: number;
  unit: string;
}

/** Structured output of the OCR module for one prescription image. */
export interface Prescription {
  id: string;
  mediaId: string;
  patient: { name: string; idNumber: string };
  issuedAt: string;
  doctor: { name: string; registryId: string };
  hasSignature: boolean;
  hasStamp: boolean;
  confidence: number;
  items: PrescribedItem[];
}

export interface DoctorCheck {
  registryId: string;
  name: string;
  registered: boolean;
  active: boolean;
  enabledToPrescribe: boolean;
}

export interface PrescriptionValidation {
  valid: boolean;
  errors: string[];
  doctor: DoctorCheck;
}
