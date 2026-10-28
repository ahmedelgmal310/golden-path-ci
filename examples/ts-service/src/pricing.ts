export interface LineItem {
  sku: string;
  unitPriceCents: number;
  quantity: number;
}

export interface Quote {
  subtotalCents: number;
  discountCents: number;
  totalCents: number;
}

/** Bulk discount: 10% off every line with 10 or more units. */
export const BULK_THRESHOLD = 10;
export const BULK_DISCOUNT = 0.1;

export function quote(items: readonly LineItem[]): Quote {
  let subtotalCents = 0;
  let discountCents = 0;
  for (const item of items) {
    if (!Number.isInteger(item.unitPriceCents) || item.unitPriceCents < 0) {
      throw new RangeError(`invalid price for ${item.sku}`);
    }
    if (!Number.isInteger(item.quantity) || item.quantity <= 0) {
      throw new RangeError(`invalid quantity for ${item.sku}`);
    }
    const line = item.unitPriceCents * item.quantity;
    subtotalCents += line;
    if (item.quantity >= BULK_THRESHOLD) {
      discountCents += Math.round(line * BULK_DISCOUNT);
    }
  }
  return { subtotalCents, discountCents, totalCents: subtotalCents - discountCents };
}
