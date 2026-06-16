export function fmtNum(value: number | null | undefined): string {
  return (value ?? 0).toLocaleString();
}

export function fmtMoney(value: number | null | undefined, currency = '$'): string {
  return `${currency}${(value ?? 0).toFixed(2)}`;
}

export function fmtEtb(value: number | null | undefined): string {
  return `ETB ${(value ?? 0).toLocaleString()}`;
}
