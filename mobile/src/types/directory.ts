export interface Institution {
  id: string;
  name: string;
}

export interface Agency {
  id: string;
  name: string;
  institutionId: string;
}
