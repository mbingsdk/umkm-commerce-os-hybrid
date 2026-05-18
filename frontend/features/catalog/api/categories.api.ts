import { apiFetch } from "@/lib/api/client";
import type { Category } from "@/features/catalog/types";

type ApiCategory = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  parent_id?: string | null;
  sort_order: number;
  is_active: boolean;
};

export type CreateCategoryInput = {
  name: string;
  slug: string;
  description?: string;
  parent_id?: string | null;
  sort_order: number;
  is_active: boolean;
};

export type UpdateCategoryInput = Partial<CreateCategoryInput>;

export async function listCategories(filters?: { isActive?: boolean }): Promise<Category[]> {
  const searchParams = new URLSearchParams();

  if (filters?.isActive !== undefined) {
    searchParams.set("is_active", String(filters.isActive));
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  const categories = await apiFetch<ApiCategory[]>(`/api/v1/categories${suffix}`);

  return categories.map(normalizeCategory);
}

export async function createCategory(input: CreateCategoryInput): Promise<Category> {
  const category = await apiFetch<ApiCategory>("/api/v1/categories", {
    method: "POST",
    body: JSON.stringify(input)
  });

  return normalizeCategory(category);
}

export async function updateCategory(categoryId: string, input: UpdateCategoryInput): Promise<Category> {
  const category = await apiFetch<ApiCategory>(`/api/v1/categories/${categoryId}`, {
    method: "PATCH",
    body: JSON.stringify(input)
  });

  return normalizeCategory(category);
}

export async function deleteCategory(categoryId: string) {
  await apiFetch<void>(`/api/v1/categories/${categoryId}`, {
    method: "DELETE"
  });
}

function normalizeCategory(category: ApiCategory): Category {
  return {
    id: category.id,
    name: category.name,
    slug: category.slug,
    description: category.description,
    parentId: category.parent_id,
    sortOrder: category.sort_order,
    isActive: category.is_active
  };
}
