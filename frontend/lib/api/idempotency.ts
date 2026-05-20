export function createIdempotencyKey(prefix: "checkout" | "pos" | "payment") {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `${prefix}_${Date.now()}_${crypto.randomUUID()}`;
  }

  return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2)}`;
}
