import type { ReactNode } from "react";
import { cn } from "@/lib/utils/cn";

export type DataTableColumn<T> = {
  key: string;
  header: string;
  className?: string;
  render: (row: T) => ReactNode;
};

type DataTableProps<T> = {
  columns: Array<DataTableColumn<T>>;
  rows: T[];
  getRowKey: (row: T) => string;
  className?: string;
};

export function DataTable<T>({ columns, rows, getRowKey, className }: DataTableProps<T>) {
  return (
    <div className={cn("overflow-x-auto rounded-2xl border border-neutral-200 bg-white", className)}>
      <table className="min-w-full divide-y divide-neutral-200 text-sm">
        <thead className="bg-neutral-50 text-left text-xs font-semibold uppercase tracking-wide text-neutral-500">
          <tr>
            {columns.map((column) => (
              <th key={column.key} className={cn("px-4 py-3", column.className)}>
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-neutral-100 text-neutral-700">
          {rows.map((row) => (
            <tr key={getRowKey(row)} className="align-top">
              {columns.map((column) => (
                <td key={column.key} className={cn("px-4 py-4", column.className)}>
                  {column.render(row)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
