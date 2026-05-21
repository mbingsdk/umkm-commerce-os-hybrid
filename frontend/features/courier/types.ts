export type CourierZone = {
  id: string;
  name: string;
  description?: string;
  rate: number;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
};

export type CourierZoneInput = {
  name: string;
  description?: string;
  rate: number;
  isActive: boolean;
  sortOrder?: number;
};
