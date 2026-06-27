export function normalizePhone(raw: string): string {
  const digits = raw.replace(/[^0-9]/g, "");
  if (digits.length === 10) return "1" + digits;
  if (digits.length === 11 && digits.startsWith("1")) return digits;
  if (digits.length > 11) return digits.replace(/^1?/, "1").slice(0, 11);
  return digits;
}

export function isCompletePhone(raw: string): boolean {
  const digits = raw.replace(/[^0-9]/g, "");
  return digits.length >= 11 && digits.startsWith("1");
}

export function formatPhone(raw: string): string {
  const n = normalizePhone(raw);
  if (n.length < 11) return raw;
  return `1+${n.slice(1, 4)}-${n.slice(4, 7)}-${n.slice(7, 11)}`;
}
