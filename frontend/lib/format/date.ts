export function formatDate(value: Date | string) {
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "full"
  }).format(typeof value === "string" ? new Date(value) : value);
}
