import { DashboardView } from "./DashboardPage.view";
import { useDashboard } from "./useDashboard";

export function DashboardPage() {
  const { stats, activities, trafficHistory, topology, loading, error } = useDashboard();

  return (
    <DashboardView
      stats={stats}
      activities={activities}
      trafficHistory={trafficHistory}
      topology={topology}
      loading={loading}
      error={error}
    />
  );
}
