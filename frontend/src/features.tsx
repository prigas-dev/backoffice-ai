import { FC, lazy } from "react";

export const FeatureComponents: Record<string, FC> = {

  "kanban-board": lazy(() => import("./components/kanban-board")),
  "task-statistics-dashboard": lazy(() => import("./components/task-statistics-dashboard")),
};
