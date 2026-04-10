import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { Warehouse } from "../types/warehouse";
import { api } from "../api/client";

interface WarehouseState {
  warehouses: Warehouse[];
  activeWarehouseId: number;
  loading: boolean;
  setActiveWarehouse: (id: number) => void;
  fetchWarehouses: () => Promise<void>;
}

export const useWarehouseStore = create<WarehouseState>()(
  persist(
    (set) => ({
      warehouses: [],
      activeWarehouseId: 1,
      loading: false,

      setActiveWarehouse: (id: number) => set({ activeWarehouseId: id }),

      fetchWarehouses: async () => {
        set({ loading: true });
        try {
          const warehouses = await api.listWarehouses();
          set((state) => {
            const ids = warehouses.map((w: Warehouse) => w.id);
            const activeStillValid = ids.includes(state.activeWarehouseId);
            return {
              warehouses,
              activeWarehouseId: activeStillValid
                ? state.activeWarehouseId
                : (warehouses[0]?.id ?? 1),
              loading: false,
            };
          });
        } catch {
          set({ loading: false });
        }
      },
    }),
    {
      name: "wms-warehouse",
      partialize: (state) => ({
        activeWarehouseId: state.activeWarehouseId,
      }),
    },
  ),
);
